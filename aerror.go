// This package adds async errors handling support in GO. See the README for
// more details.
package aerrors

import (
	"fmt"
)

var defaultErrorChanLen = 10

// AsyncError queues your added errors in chan and handle them by provided handler method.
// It may be started and stopped and run your func in a panic-safe goroutine.
type AsyncError struct {
	logger       Logger
	add          chan error
	stop         chan struct{}
	baseError    error
	handler      ErrorHandler
	errorChanLen int
	errorChan    chan error
}

// ErrorHandler is an interface for defining error handler
type ErrorHandler interface {
	HandleError(err error)
}

// New returns a new AsyncError, modified by the given options.
//
// Available Settings
//
//   Logger
//     Description: The logger which will log your error messages
//     Default:     PrintfLogger
//
//   errorChanLen
//     Description: Available chan length for adding errors to handle queue
//     Default:     10
//
//   baseError
//     Description: All added errors will be chained with provided baseError
//     Default:     nil
//
//   ErrorHandler
//     Description: Handler function for added errors. By default handler will log your error with defined logger
//     Default:     nil
//
// See "aerrors.With*" to modify the default behavior.
func New(opts ...Option) *AsyncError {
	a := &AsyncError{
		add:       make(chan error),
		stop:      make(chan struct{}),
		logger:    DefaultLogger,
		errorChan: make(chan error, defaultErrorChanLen),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Add puts your error in queue to handle. It blocks if we reached chan length
func (e *AsyncError) Add(err error) {
	e.errorChan <- err
}

// AddAsync puts your error in queue in goroutine. It not blocks when we reached chan length
func (e *AsyncError) AddAsync(err error) {
	go e.Add(err)
}

// Stop handle errors
func (e *AsyncError) Stop() {
	e.stop <- struct{}{}
}

// StartHandle starts handling errors
func (e *AsyncError) StartHandle() {
	go e.start()
}

func (e *AsyncError) start() {
	for {
		select {
		case newError := <-e.errorChan:
			e.handle(newError)

		case <-e.stop:
			e.logger.Info("stop")
			return
		}
	}
}

// Wrap your error
func Wrap(errp *error, format string, args ...interface{}) {
	if errp != nil && *errp != nil {
		s := fmt.Sprintf(format, args...)
		*errp = fmt.Errorf("%s: %w", s, *errp)
	}
}

// PanicToError recovers panic and creates an error from it
func (e *AsyncError) PanicToError() {
	if p := recover(); p != nil {
		err := fmt.Errorf("%v", p)
		Wrap(&err, "recoverToError()")
		e.errorChan <- err
	}
}

// Go runs your function in panic-safe goroutine
func (e *AsyncError) Go(f func()) {
	go func() {
		defer e.PanicToError()
		f()
	}()
}

func (e *AsyncError) handle(err error) {
	err = fmt.Errorf("%w: %v", e.baseError, err)
	Wrap(&err, "HandleError()")

	if e.handler != nil {
		go func() {
			e.handler.HandleError(err)
		}()
	} else {
		e.logger.Error(err, "aerror handled error")
	}
}
