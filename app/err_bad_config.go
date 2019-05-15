package app

type ErrBadConfig struct {
	problem string
}

func (e *ErrBadConfig) Error() string {
	return "there was an error in your config file: " + e.problem
}

// NewErrRecordNotFound returns new error
func NewErrBadConfig(problem string) *ErrBadConfig {
	return &ErrBadConfig{
		problem: problem,
	}
}
