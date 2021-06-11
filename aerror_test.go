package aerrors

import (
	"errors"
	"sync"
	"testing"
	"time"
)

const sleepTime = 1 * time.Millisecond

type testHandler struct {
	errs []error
	mu   sync.Mutex
}

func (th *testHandler) HandleError(err error) {
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
	aerror := New(WithHandler(&th))
	aerror.StartHandle()
	defer aerror.Stop()
	defer th.Reset()

	aerror.Add(errors.New("testing handler"))

	if len(aerror.errorChan) != 1 {
		t.Error("expected chan len of 1")
	}

	time.Sleep(sleepTime)
	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}

	if len(aerror.errorChan) != 0 {
		t.Error("chan is not empty")
	}
}

func TestErrorHandlerMultipleAdd(t *testing.T) {
	aerror := New(WithHandler(&th))
	aerror.StartHandle()
	defer aerror.Stop()
	defer th.Reset()

	aerror.Add(errors.New("testing handler 1"))
	aerror.Add(errors.New("testing handler 2"))
	aerror.Add(errors.New("testing handler 3"))

	if len(aerror.errorChan) != 3 {
		t.Error("expected chan len of 3", len(aerror.errorChan))
	}

	time.Sleep(sleepTime)
	if len(th.errs) != 3 {
		t.Error("expected to have 3 errors")
	}

	if len(aerror.errorChan) != 0 {
		t.Error("chan is not empty")
	}
}

func TestWithBaseError(t *testing.T) {
	aerror := New(WithHandler(&th), WithBaseError(&testErr))
	aerror.StartHandle()
	defer aerror.Stop()
	defer th.Reset()

	aerror.Add(errors.New("testing handler with base error"))

	if len(aerror.errorChan) != 1 {
		t.Error("expected chan len of 1")
	}

	time.Sleep(sleepTime)
	if len(th.errs) != 1 {
		t.Error("expected to have an errors")
	}

	if len(aerror.errorChan) != 0 {
		t.Error("chan is not empty")
	}

	if !errors.Is(th.errs[0], &testErr) {
		t.Error("expected test error type")
	}

	if errors.Unwrap(errors.Unwrap(th.errs[0])).Error() != testErr.Error() {
		t.Error("expected test error next to the current in chain")
	}
}

func TestPanicInGoroutine(t *testing.T) {
	aerror := New(WithHandler(&th), WithBaseError(&testErr))
	aerror.StartHandle()
	defer aerror.Stop()
	defer th.Reset()

	if len(th.errs) != 0 {
		t.Error("expected to have 0 errors")
	}

	go func() {
		defer aerror.PanicToError()
		panic("test panic")
	}()
	time.Sleep(sleepTime)

	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}
}

func TestPanicInGoroutineWrapper(t *testing.T) {
	aerror := New(WithHandler(&th), WithBaseError(&testErr))
	aerror.StartHandle()
	defer aerror.Stop()
	defer th.Reset()

	if len(th.errs) != 0 {
		t.Error("expected to have 0 errors")
	}

	f := func() {
		panic("test panic in Go wrapper")
	}
	aerror.Go(f)
	time.Sleep(sleepTime)

	if len(th.errs) != 1 {
		t.Error("expected to have an error")
	}
}

func TestOverflowErrorChan(t *testing.T) {
	aerror := New(WithHandler(&th), WithBaseError(&testErr), WithErrorChanLen(2))
	defer aerror.Stop()
	defer th.Reset()

	aerror.Add(errors.New("testing handler 1"))
	aerror.Add(errors.New("testing handler 2"))
	aerror.AddAsync(errors.New("testing handler 3"))
	aerror.AddAsync(errors.New("testing handler 4"))
	aerror.AddAsync(errors.New("testing handler 5"))

	if len(aerror.errorChan) != 2 {
		t.Error("expected to have 2 errors in chan")
	}

	aerror.StartHandle()

	time.Sleep(sleepTime)
	if len(th.errs) != 5 {
		t.Error("expected to have 2 errors in chan")
	}
}