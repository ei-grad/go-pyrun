package pyrun

// #cgo pkg-config: python-2.7
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

// Execute a Python code and return an error if any.
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

// EvalToString evaluates a single Python expression and returns the result as
// string.
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

var shutdown chan struct{}
var shutdownDone chan struct{}
var initializeMutex sync.Mutex

// Initialize a Python interpreter. It is called automaticaly during
// NewPython() call if needed. It does nothing if interpreter is already
// initialized.
func Initialize() {

	initializeMutex.Lock()
	defer initializeMutex.Unlock()

	if shutdown != nil {
		return
	}

	shutdown = make(chan struct{})
	shutdownDone = make(chan struct{})

	ready := make(chan struct{})
	defer close(ready)

	go func() {
		runtime.LockOSThread()
		C.Py_Initialize()
		defer func() {
			C.Py_Finalize()
			close(shutdownDone)
		}()
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
	}()
	<-ready
}

var once sync.Once

var shutdownMutex sync.Mutex

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

// Finalize a python interpreter. If interpreter is not running it does nothing.
// Should be called when there is no need to execute a Python code anymore.
func Finalize() {

	// don't allow several Finalize calls to control the shutdown/shutdownDone channels
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()

	// if no interpreter is running, do nothing
	if shutdown == nil {
		return
	}

	// notify a worker goroutine from Initialize that it should finish its task
	// and break the loop
	close(shutdown)

	// wait for it to finish
	<-shutdownDone

	// reinitialize
	shutdown = nil
	shutdownDone = nil
	once = sync.Once{}
}
