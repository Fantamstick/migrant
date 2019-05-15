package app

import "fmt"

type ErrSecretNotFound struct {
	key    string
	source string
}

func (e *ErrSecretNotFound) Error() string {
	return fmt.Sprintf("The specified secret was not found: %s - %s", e.source, e.key)
}

// NewErrRecordNotFound returns new error
func NewErrSecretNotFound(source, key string) *ErrSecretNotFound {
	return &ErrSecretNotFound{
		source: source,
		key:    key,
	}
}
