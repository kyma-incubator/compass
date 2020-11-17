package apperrors

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type Error struct {
	errorCode ErrorType
	Message   string
	arguments map[string]string
	parentErr error
}

func (err Error) Error() string {
	builder := strings.Builder{}
	builder.WriteString(err.Message)

	var i = 0
	if len(err.arguments) != 0 {
		builder.WriteString(" [")
		keys := sortMapKey(err.arguments)
		for _, key := range keys {
			builder.WriteString(fmt.Sprintf("%s=%s", key, err.arguments[key]))
			i++
			if len(err.arguments) != i {
				builder.WriteString("; ")
			}
		}
		builder.WriteString("]")
	}
	if err.errorCode == InternalError && err.parentErr != nil {
		builder.WriteString(": ")
		builder.WriteString(err.parentErr.Error())
	}

	return builder.String()
}

func (e Error) Is(err error) bool {
	if customErr, ok := err.(Error); ok {
		return e.errorCode == customErr.errorCode
	} else {
		return false
	}
}

func ErrorCode(err error) ErrorType {
	var customErr Error
	found := errors.As(err, &customErr)
	if found {
		return customErr.errorCode
	}
	return UnknownError

}

func NewNotNullViolationError(resourceType resource.Type) error {
	return Error{
		errorCode: EmptyData,
		Message:   emptyDataMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

func NewCheckViolationError(resourceType resource.Type) error {
	return Error{
		errorCode: InconsistentData,
		Message:   inconsistentDataMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

func NewOperationTimeoutError() error {
	return Error{
		errorCode: OperationTimeout,
		Message:   operationTimeoutMsg,
	}
}

func NewNotUniqueError(resourceType resource.Type) error {
	return Error{
		errorCode: NotUnique,
		Message:   notUniqueMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

func NewNotFoundError(resourceType resource.Type, objectID string) error {
	return Error{
		errorCode: NotFound,
		Message:   notFoundMsg,
		arguments: map[string]string{"object": string(resourceType), "ID": objectID},
	}
}

func NewNotFoundErrorWithType(resourceType resource.Type) error {
	return Error{
		errorCode: NotFound,
		Message:   notFoundMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

func NewInvalidDataError(msg string, args ...interface{}) error {
	return Error{
		errorCode: InvalidData,
		Message:   invalidDataMsg,
		arguments: map[string]string{"reason": fmt.Sprintf(msg, args...)},
	}
}

func NewInvalidDataErrorWithFields(fields map[string]error, objType string) error {
	if len(fields) == 0 {
		return nil
	}

	err := Error{
		errorCode: InvalidData,
		Message:   fmt.Sprintf("%s %s", invalidDataMsg, objType),
		arguments: map[string]string{},
	}

	for k, v := range fields {
		err.arguments[k] = v.Error()
	}
	return err
}

func NewInternalError(msg string, args ...interface{}) error {
	errMsg := fmt.Sprintf(msg, args...)
	return Error{
		errorCode: InternalError,
		Message:   fmt.Sprintf(internalServerErrMsgF, errMsg),
		arguments: map[string]string{},
	}
}

func InternalErrorFrom(err error, msg string, args ...interface{}) error {
	errMsg := fmt.Sprintf(msg, args...)
	return Error{
		errorCode: InternalError,
		Message:   fmt.Sprintf(internalServerErrMsgF, errMsg),
		arguments: map[string]string{},
		parentErr: err,
	}
}

func NewTenantNotFoundError(externalTenant string) error {
	return Error{
		errorCode: TenantNotFound,
		Message:   tenantNotFoundMsg,
		arguments: map[string]string{"externalTenant": externalTenant},
	}
}

func NewTenantRequiredError() error {
	return Error{
		errorCode: TenantRequired,
		Message:   tenantRequiredMsg,
		arguments: map[string]string{},
	}
}

func NewInvalidOperationError(reason string) error {
	return Error{
		errorCode: InvalidOperation,
		Message:   invalidOperationMsg,
		arguments: map[string]string{"reason": reason},
	}
}

func NewForeignKeyInvalidOperationError(sqlOperation resource.SQLOperation, resourceType resource.Type) error {
	var reason string
	switch sqlOperation {
	case resource.Create, resource.Update, resource.Upsert:
		reason = "The referenced entity does not exists"
	case resource.Delete:
		reason = "The record cannot be deleted because another record refers to it"
	}

	return Error{
		errorCode: InvalidOperation,
		Message:   invalidOperationMsg,
		arguments: map[string]string{"reason": reason, "object": string(resourceType)},
	}
}

const valueNotFoundInConfigMsg = "value under specified path not found in configuration"

func NewValueNotFoundInConfigurationError() error {
	return Error{
		errorCode: NotFound,
		Message:   valueNotFoundInConfigMsg,
		arguments: map[string]string{},
	}
}

func NewNoScopesInContextError() error {
	return Error{
		errorCode: NotFound,
		Message:   noScopesInContextMsg,
		arguments: map[string]string{},
	}
}

func NewRequiredScopesNotDefinedError() error {
	return Error{
		errorCode: InsufficientScopes,
		Message:   noRequiredScopesInContextMsg,
		arguments: map[string]string{},
	}
}

func NewKeyDoesNotExistError(key string) error {
	return Error{
		errorCode: NotFound,
		Message:   keyDoesNotExistMsg,
		arguments: map[string]string{"key": key},
	}
}

func NewInsufficientScopesError(requiredScopes, actualScopes []string) error {
	return Error{
		errorCode: InsufficientScopes,
		Message:   insufficientScopesMsg,
		arguments: map[string]string{"required": strings.Join(requiredScopes, ";"),
			"actual": strings.Join(actualScopes, ";")},
		parentErr: nil,
	}
}

func NewCannotReadTenantError() error {
	return Error{
		errorCode: InternalError,
		Message:   cannotReadTenantMsg,
		arguments: map[string]string{},
	}
}

func NewCannotReadClientUserError() error {
	return Error{
		errorCode: InternalError,
		Message:   cannotReadClientUserMsg,
		arguments: map[string]string{},
	}
}

func NewUnauthorizedError(msg string) error {
	return Error{
		errorCode: Unauthorized,
		Message:   unauthorizedMsg,
		arguments: map[string]string{"reason": msg},
	}
}

func IsValueNotFoundInConfiguration(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound && customErr.Message == valueNotFoundInConfigMsg
	} else {
		return false
	}
}

func IsKeyDoesNotExist(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound && customErr.Message == keyDoesNotExistMsg
	} else {
		return false
	}
}

func IsCannotReadTenant(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == InternalError && customErr.Message == cannotReadTenantMsg
	} else {
		return false
	}
}

func IsNewInvalidDataError(err error) bool {
	return ErrorCode(err) == InvalidOperation
}

func IsNotFoundError(err error) bool {
	return ErrorCode(err) == NotFound
}

func IsTenantRequired(err error) bool {
	return ErrorCode(err) == TenantRequired
}

func IsTenantNotFoundError(err error) bool {
	return ErrorCode(err) == TenantNotFound
}

func IsNotUniqueError(err error) bool {
	return ErrorCode(err) == NotUnique
}

func IsNewNotNullViolationError(err error) bool {
	return ErrorCode(err) == EmptyData
}

func IsNewCheckViolationError(err error) bool {
	return ErrorCode(err) == InconsistentData
}

func sortMapKey(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
