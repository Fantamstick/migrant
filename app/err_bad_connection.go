package app

type ErrBadConnection struct {
	problem string
}

func (e *ErrBadConnection) Error() string {
	return "There was a connection error: " + e.problem
}

// NewErrRecordNotFound returns new error
func NewErrBadConnection(problem string) *ErrBadConnection {
	return &ErrBadConnection{
		problem: problem,
	}
}
