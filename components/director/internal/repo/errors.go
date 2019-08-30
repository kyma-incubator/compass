package repo

import "github.com/pkg/errors"

type NotFound interface {
	IsNotFound() bool
}

type notFoundError struct{}

func NewNotFoundError() *notFoundError {
	return &notFoundError{}
}

func (e *notFoundError) Error() string {
	return "object not found in DB"
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

type notUniqueError struct{}

func (e *notUniqueError) Error() string {
	return "unique constraint violation"
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
