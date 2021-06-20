// This package adds async errors handling support in GO. See the README for
// more details.
package aerrors

import (
	"errors"
	"fmt"
	"sync"
)

var defaultErrorChanLen = 10

// AsyncError queues your added errors in chan and handle them by provided handler method.
// It may be started and stopped and run your func in a panic-safe goroutine.
type AsyncError struct {
	logger       Logger
	stopCh       chan struct{}
	baseError    error
	handler      ErrorHandler
	errorChanLen int
	errorChan    chan error
	closed       bool
	running      bool
	mu           sync.Mutex
	wg           sync.WaitGroup
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
		stopCh:    make(chan struct{}),
		logger:    DefaultLogger,
		errorChan: make(chan error, defaultErrorChanLen),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// IsClosed checks if it closed
func (e *AsyncError) IsClosed() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.closed
}

// IsRunning checks if errors handles by handle func
func (e *AsyncError) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.running
}

// Add puts your error in queue to handle. It blocks if we reached chan length
func (e *AsyncError) Add(err error) error {
	if e.IsClosed() {

		return errors.New("aerrors: can't add error to closed chan")
	}

	e.wg.Add(1)
	e.errorChan <- err

	return nil
}

// AddAsync puts your error in queue in goroutine. It not blocks when we reached chan length
func (e *AsyncError) AddAsync(err error) error {
	if e.IsClosed() {

		return errors.New("aerrors: can't async add error to closed chan")
	}
	e.wg.Add(1)
	go func() { e.errorChan <- err }()

	return nil
}

// Stop handle errors
func (e *AsyncError) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stop()
}

func (e *AsyncError) stop() {
	if !e.running {
		return
	}

	e.running = false
	e.stopCh <- struct{}{}
}

// Close the aerror gracefully. It waits to handle all errors from queue.
// You can't use it after closing and have to create a new one.
func (e *AsyncError) Close() {
	if e.IsClosed() {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.closed = true
	if e.running {
		e.wg.Wait()
		e.stop()
	}

	close(e.errorChan)
	close(e.stopCh)
}

// StartHandle starts handling errors
func (e *AsyncError) StartHandle() error {
	if e.IsClosed() {

		return errors.New("aerrors: can't start handle for closed aerror")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.running = true
	go e.start()

	return nil
}

func (e *AsyncError) start() {
	for {
		select {
		case newError := <-e.errorChan:
			if newError != nil {
				e.handle(newError)
				e.wg.Done()
			}

		case <-e.stopCh:
			e.logger.Info("aerrors: stop")
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
		_ = e.Add(err)
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
