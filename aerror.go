package aerrors

import (
	"fmt"
)

var ErrorChanLen = 5
var errorChan = make(chan error, ErrorChanLen)

type AsyncError struct {
	logger    Logger
	add       chan error
	stop      chan struct{}
	baseError error
	handler   ErrorHandler
}

func New(opts ...Option) *AsyncError {
	c := &AsyncError{
		add:    make(chan error),
		stop:   make(chan struct{}),
		logger: DefaultLogger,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (e *AsyncError) Add(err error) {
	errorChan <- err
}

func (e *AsyncError) Stop() {
	e.stop <- struct{}{}
}

func (e *AsyncError) StartHandle() {
	go e.start()
}

func (e *AsyncError) start() {
	for {
		select {
		case newError := <-errorChan:
			e.handle(newError)

		case <-e.stop:
			e.logger.Info("stop")
			return
		}
	}
}

func Wrap(errp *error, format string, args ...interface{}) {
	if errp != nil {
		s := fmt.Sprintf(format, args...)
		*errp = fmt.Errorf("%s: %w", s, *errp)
	}
}

func PanicToError() {
	if p := recover(); p != nil {
		err := fmt.Errorf("%v", p)
		Wrap(&err, "recoverToError()")
		errorChan <- err
	}
}

func Go(f func()) {
	go func() {
		defer PanicToError()
		f()
	}()
}

func (e *AsyncError) handle(err error) {
	err = fmt.Errorf("%w: %v", e.baseError, err)
	e.logger.Error(err, "error handled")
	if e.handler != nil {
		go func() {
			e.handler.Handle(err)
		}()
	}
}
