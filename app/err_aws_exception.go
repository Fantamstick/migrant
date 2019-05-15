package app

type ErrAwsException struct {
	code    string
	message string
}

func (e *ErrAwsException) Error() string {
	return "something went wrong while interfacing with aws: " + e.code + " -- " + e.message
}

// NewErrRecordNotFound returns new error
func NewErrAwsException(code, message string) *ErrAwsException {
	return &ErrAwsException{
		code:    code,
		message: message,
	}
}
