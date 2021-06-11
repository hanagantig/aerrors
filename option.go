package aerrors

// Option represents a modification to the default behavior of a serro.
type Option func(s *AsyncError)

// WithLogger uses the provided logger.
func WithLogger(logger Logger) Option {
	return func(e *AsyncError) {
		e.logger = logger
	}
}

// WithBaseError uses the provided base error.
func WithBaseError(err error) Option {
	return func(e *AsyncError) {
		e.baseError = err
	}
}

// WithHandler uses the provided error handler.
func WithHandler(h ErrorHandler) Option {
	return func(e *AsyncError) {
		e.handler = h
	}
}

// WithErrorChanLen sets the error chan length.
func WithErrorChanLen(l int) Option {
	return func(e *AsyncError) {
		e.errorChan = make(chan error, l)
	}
}
