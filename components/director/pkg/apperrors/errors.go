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
	if e.identifier == "" {
		return "Object was not found"
	}
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

type ConstraintViolation interface {
	ConstraintViolation()
}

type constraintViolationError struct {
	table string
}

func NewConstraintViolationError(table string) *constraintViolationError {
	return &constraintViolationError{
		table: table,
	}
}

func (e *constraintViolationError) Error() string {
	return fmt.Sprintf("this object still has %s referenced by it", e.table)
}

func (constraintViolationError) ConstraintViolation() {}

func IsConstraintViolation(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}
	if _, ok := err.(ConstraintViolation); ok {
		return true
	}
	return false
}

type InvalidCast interface {
	InvalidCast()
}

type invalidStringCastError struct{}

func NewInvalidStringCastError() *invalidStringCastError {
	return &invalidStringCastError{}
}

func (e *invalidStringCastError) Error() string {
	return "unable to cast the value to a string type"
}

func (invalidStringCastError) InvalidCast() {}

func IsInvalidCast(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(InvalidCast)
	return ok
}

type KeyDoesNotExist interface {
	KeyDoesNotExist()
}

type keyDoesNotExistError struct {
	key string
}

func NewKeyDoesNotExistError(key string) *keyDoesNotExistError {
	return &keyDoesNotExistError{
		key: key,
	}
}

func (e *keyDoesNotExistError) Error() string {
	return fmt.Sprintf("the key (%s) does not exist in source object", e.key)
}

func (keyDoesNotExistError) KeyDoesNotExist() {}

func IsKeyDoesNotExist(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(KeyDoesNotExist)
	return ok
}
