package pyrun

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	py := NewPython()
	err := py.Execute("a = lambda: 'Hello, world!'")
	if err != nil {
		t.Fatalf("Execute failed: %s", err)
	}
	ret, err := py.EvalToString("a()")
	if err != nil {
		t.Fatalf("EvalToString failed: %s", err)
	}
	if ret != "Hello, world!" {
		t.Fatalf("Get unexpected result: %s", ret)
	}
}
