package app

type ErrSecretsAlreadyDefined struct {
	source string
}

func (e *ErrSecretsAlreadyDefined) Error() string {
	return "source already exists so was not imported: " + e.source
}

// NewErrRecordNotFound returns new error
func NewErrSecretsAlreadyDefined(source string) *ErrSecretsAlreadyDefined {
	return &ErrSecretsAlreadyDefined{
		source: source,
	}
}
