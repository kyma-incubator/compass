package apperrors

import (
	"fmt"

	"github.com/pkg/errors"
)

type NotFound interface {
	IsNotFound() bool
}

type notFoundError struct {
	identifier string
}

func NewNotFoundError(identifier string) *notFoundError {
	return &notFoundError{
		identifier: identifier,
	}
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("Object %s not found", e.identifier)
}

func (e *notFoundError) IsNotFound() bool {
	return true
}

func IsNotFoundError(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}
	if _, ok := err.(NotFound); ok {
		return true
	}
	return false
}

type NotUnique interface {
	IsNotUnique()
}

type notUniqueError struct {
	identifier string
}

func NewNotUniqueError(identifier string) *notUniqueError {
	return &notUniqueError{
		identifier: identifier,
	}
}

func (e *notUniqueError) Error() string {
	if e.identifier == "" {
		return "Object is not unique"
	}
	return fmt.Sprintf("Object %s is not unique", e.identifier)
}

func (notUniqueError) IsNotUnique() {}

func IsNotUnique(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}
	if _, ok := err.(NotUnique); ok {
		return true
	}
	return false
}
