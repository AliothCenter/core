package errors

import "fmt"

type RestorationExternalResponseError struct {
	basicAliothError
	code int
}

func (e *RestorationExternalResponseError) Error() string {
	return fmt.Sprintf("restoration external response error, status %d", e.code)
}

func NewRestorationExternalResponseError(code int) AliothError {
	return &RestorationExternalResponseError{
		code: code,
	}
}
