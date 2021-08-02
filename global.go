package aerrors

import "errors"

var gAerror *AsyncError
var hasInitialized = errors.New("error: try to initialize global aerror while other is running")

// Init a single globally async error handler
func Init(opts ...Option) error {
	if gAerror != nil && gAerror.IsRunning() {
		return hasInitialized
	}
	a := New(opts...)
	gAerror = a
	return nil
}

// Get initialized global async error handler
func Get() *AsyncError {
	return gAerror
}
