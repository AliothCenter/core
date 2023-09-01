package errors

import (
	"errors"
)

type AliothError interface {
	Error() string
	Derive(target any) bool
	Equal(target error) bool
}

type basicAliothError struct{}

func (basicAliothError) Error() string {
	return "unimplemented alioth error"
}

func (bae basicAliothError) Derive(target any) bool {
	return errors.As(bae, &target)
}

func (bae basicAliothError) Equal(target error) bool {
	return errors.Is(bae, target)
}

type GenerationAliothError struct {
	basicAliothError
	err error
}

func (gae *GenerationAliothError) Error() string {
	return gae.err.Error()
}

func NewAliothError(err error) AliothError {
	var aliothErr AliothError
	if errors.As(err, &aliothErr) {
		return err.(AliothError)
	}

	return &GenerationAliothError{
		err: err,
	}
}
