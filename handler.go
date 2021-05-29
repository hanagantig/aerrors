package aerrors

type ErrorHandler interface {
	Handle(err error)
}

