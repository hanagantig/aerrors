package serro

type ErrorHandler interface {
	Handle(err error)
}

