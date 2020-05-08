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
	_, ok := err.(NotFound)
	return ok
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
	_, ok := err.(NotUnique)
	return ok
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
	_, ok := err.(ConstraintViolation)
	return ok
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

// ValueNotFoundInConfiguration

type ValueNotFoundInConfiguration interface {
	ValueNotFoundInConfiguration()
}

type valueNotFoundInConfigurationError struct{}

func NewValueNotFoundInConfigurationError() *valueNotFoundInConfigurationError {
	return &valueNotFoundInConfigurationError{}
}

func (e *valueNotFoundInConfigurationError) Error() string {
	return "value under specified path not found in configuration"
}

func (valueNotFoundInConfigurationError) ValueNotFoundInConfiguration() {}

func IsValueNotFoundInConfiguration(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(ValueNotFoundInConfiguration)
	return ok
}

// NoScopesInContextError

type NoScopesInContext interface {
	NoScopesInContext()
}

type noScopesInContextError struct{}

func NewNoScopesInContextError() *noScopesInContextError {
	return &noScopesInContextError{}
}

func (e *noScopesInContextError) Error() string {
	return "cannot read scopes from context"
}

func (noScopesInContextError) NoScopesInContext() {}

func IsNoScopesInContext(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(NoScopesInContext)
	return ok
}

// RequiredScopesNotDefinedError

type RequiredScopesNotDefined interface {
	RequiredScopesNotDefined()
}

type requiredScopesNotDefinedError struct{}

func NewRequiredScopesNotDefinedError() *requiredScopesNotDefinedError {
	return &requiredScopesNotDefinedError{}
}

func (e *requiredScopesNotDefinedError) Error() string {
	return "required scopes are not defined"
}

func (requiredScopesNotDefinedError) RequiredScopesNotDefined() {}

func IsRequiredScopesNotDefined(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(RequiredScopesNotDefined)
	return ok
}

// InsufficientScopesError

type InsufficientScopes interface {
	InsufficientScopes()
}

type insufficientScopesError struct {
	required []string
	actual   []string
}

func NewInsufficientScopesError(requiredScopes, actualScopes []string) *insufficientScopesError {
	return &insufficientScopesError{
		required: requiredScopes,
		actual:   actualScopes,
	}
}

func (e *insufficientScopesError) Error() string {
	return fmt.Sprintf("insufficient scopes provided, required: %v, actual: %v", e.required, e.actual)
}

func (insufficientScopesError) InsufficientScopes() {}

func IsInsufficientScopes(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(InsufficientScopes)
	return ok
}

type NoTenant interface {
	NoTenant()
}

type noTenantError struct {
}

func NewNoTenantError() *noTenantError {
	return &noTenantError{}
}

func (e *noTenantError) Error() string {
	return "cannot read tenant from context"
}

func (noTenantError) NoTenant() {}

func IsNoTenant(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(NoTenant)
	return ok
}

type EmptyTenant interface {
	EmptyTenant()
}

type emptyTenantError struct {
}

func NewEmptyTenantError() *emptyTenantError {
	return &emptyTenantError{}
}

func (e *emptyTenantError) Error() string {
	return "internal tenantID is empty"
}

func (emptyTenantError) EmptyTenant() {}

func IsEmptyTenant(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}

	_, ok := err.(EmptyTenant)
	return ok
}
