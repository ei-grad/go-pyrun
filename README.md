Go PyRun - simple interactation with Python
===========================================

It runs a Python interpreter in dedicated thread, and allows to run Python code
and get result from it.

Install
-------

    go get github.com/ei-grad/go-pyrun

Example
-------

```go
package main

import "github.com/ei-grad/go-pyrun"
import "fmt"
import "log"

func main() {
    py := pyrun.NewPython()
    err := py.Execute(`
    def hello():
        return 'Hello, world!'
    `)
    if err != nil {
        log.Fatalf("Execute failed: %s", err)
    }
    ret, err := py.EvalToString("hello()")
    if err != nil {
        log.Fatalf("EvalToString failed: %s", err)
    }
    fmt.Printf("Hello from Python: %s\n", ret)
}
```
