package pyrun

// #cgo pkg-config: python2
// #include "pyrun.h"
import "C"

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"unsafe"
)

// Python is an interpreter abstraction. Each Python{} object run commands in
// its own context. But really they share a single CPython interpreter
// instance.
type Python struct {
	ctx *C.PyObject
}

// NewPython creates new context and ensures that the Python interpreter is
// initialized.
func NewPython() *Python {
	once.Do(Initialize)
	ret := Python{ctx: C.GoPyRunGetContext()}
	if ret.ctx == nil {
		log.Fatal("Can't initialize Python!")
	}
	return &ret
}

type taskKind int

const (
	taskFileInput taskKind = iota
	taskEvalInput
)

type task struct {
	kind    taskKind
	command *C.char
	ctx     *C.PyObject
	out     chan *C.PyObject
}

var tasks = make(chan *task)
var once sync.Once

// Execute will a python code and return an error, if any.
func (py *Python) Execute(cmd string) error {
	t := task{
		kind:    taskFileInput,
		command: C.CString(cmd),
		ctx:     py.ctx,
		out:     make(chan *C.PyObject),
	}
	defer C.free(unsafe.Pointer(t.command))
	tasks <- &t
	obj := <-t.out
	if obj == nil {
		return errors.New("PyRun_String finished with error.")
	}
	defer C.GoPyRunDecref(obj)
	return nil
}

// EvalToString executes a single isolated Python expression and returns result
// of its evaluation as string.
func (py *Python) EvalToString(cmd string) (string, error) {
	t := task{
		kind:    taskEvalInput,
		command: C.CString(cmd),
		ctx:     py.ctx,
		out:     make(chan *C.PyObject),
	}
	defer C.free(unsafe.Pointer(t.command))
	tasks <- &t
	obj := <-t.out
	if obj == nil {
		return "", errors.New("PyRun_String finished with error.")
	}
	defer C.GoPyRunDecref(obj)
	return C.GoString(C.PyString_AsString(obj)), nil
}

// Initialize a python interpreter, it is called automaticaly with sync.Once
// when needed.
func Initialize() {
	ready := make(chan struct{})
	defer close(ready)
	go func() {
		runtime.LockOSThread()
		C.Py_Initialize()
		ready <- struct{}{}
	loop:
		for {
			select {
			case task := <-tasks:
				var ret *C.PyObject
				switch task.kind {
				case taskEvalInput:
					ret = C.GoPyRunEvalInput(task.ctx, task.command)
				case taskFileInput:
					ret = C.GoPyRunFileInput(task.ctx, task.command)
				}
				task.out <- ret
			case <-shutdown:
				break loop
			}
		}
		C.Py_Finalize()
		shutdownDone <- struct{}{}
	}()
	<-ready
}

var shutdown = make(chan struct{})
var shutdownDone = make(chan struct{})
var shutdownMutex sync.Mutex

// Finalize a python interpreter.
func Finalize() {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()
	close(shutdown)
	<-shutdownDone
	shutdown = make(chan struct{})
	once = sync.Once{}
}
