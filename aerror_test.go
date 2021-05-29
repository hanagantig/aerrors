package serro

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type testHandler struct {
	errs []error
	mu   sync.Mutex
}

func (th *testHandler) Handle(err error) {
	th.mu.Lock()
	defer th.mu.Unlock()

	th.errs = append(th.errs, err)
}

func (th *testHandler) Reset() {
	th.mu.Lock()
	defer th.mu.Unlock()

	th.errs = []error{}
}

var th = testHandler{
	errs: []error{},
}

type testErrorType struct{}

func (e *testErrorType) Error() string {
	return "[Test Error Type]"
}

var testErr = testErrorType{}

func TestErrorHandler(t *testing.T) {
	errHandler := New(WithHandler(&th))
	errHandler.StartHandle()
	defer errHandler.Stop()
	defer th.Reset()

	errHandler.Add(errors.New("testing handler"))

	if len(errorChan) != 1 {
		t.Error("expected chan len of 3")
	}

	time.Sleep(1 * time.Second)
	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}

	if len(errorChan) != 0 {
		t.Error("chan is not empty")
	}
}

func TestErrorHandlerMultipleAdd(t *testing.T) {
	errHandler := New(WithHandler(&th))
	errHandler.StartHandle()
	defer errHandler.Stop()
	defer th.Reset()

	errHandler.Add(errors.New("testing handler 1"))
	errHandler.Add(errors.New("testing handler 2"))
	errHandler.Add(errors.New("testing handler 3"))

	if len(errorChan) != 3 {
		t.Error("expected chan len of 3", len(errorChan))
	}

	time.Sleep(1 * time.Second)
	if len(th.errs) != 3 {
		t.Error("expected to have 3 errors")
	}

	if len(errorChan) != 0 {
		t.Error("chan is not empty")
	}
}

func TestWithBaseError(t *testing.T) {
	errHandler := New(WithHandler(&th), WithBaseError(&testErr))
	errHandler.StartHandle()
	defer errHandler.Stop()
	defer th.Reset()

	errHandler.Add(errors.New("testing handler with base error"))

	if len(errorChan) != 1 {
		t.Error("expected chan len of 1")
	}

	time.Sleep(1 * time.Second)
	if len(th.errs) != 1 {
		t.Error("expected to have an errors")
	}

	if len(errorChan) != 0 {
		t.Error("chan is not empty")
	}

	if !errors.Is(th.errs[0], &testErr) {
		t.Error("expected test error type")
	}

	if errors.Unwrap(th.errs[0]).Error() != testErr.Error() {
		t.Error("expected test error next to the current in chain")
	}
}

func TestPanicInGoroutine(t *testing.T) {
	errHandler := New(WithHandler(&th), WithBaseError(&testErr))
	errHandler.StartHandle()
	defer errHandler.Stop()
	defer th.Reset()

	if len(th.errs) != 0 {
		t.Error("expected to have 0 errors")
	}

	go func() {
		defer PanicToError()
		panic("test panic")
	}()
	time.Sleep(1 * time.Second)

	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}
}

func TestPanicInGoroutineWrapper(t *testing.T) {
	errHandler := New(WithHandler(&th), WithBaseError(&testErr))
	errHandler.StartHandle()
	defer errHandler.Stop()
	defer th.Reset()

	if len(th.errs) != 0 {
		t.Error("expected to have 0 errors")
	}

	f := func() {
		panic("test panic in Go wrapper")
	}
	Go(f)
	time.Sleep(1 * time.Second)

	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}
}
