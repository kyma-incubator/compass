package apperrors

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Error missing godoc
type Error struct {
	errorCode ErrorType
	Message   string
	arguments map[string]string
	parentErr error
}

// Error missing godoc
func (e Error) Error() string {
	builder := strings.Builder{}
	builder.WriteString(e.Message)

	var i = 0
	if len(e.arguments) != 0 {
		builder.WriteString(" [")
		keys := sortMapKey(e.arguments)
		for _, key := range keys {
			builder.WriteString(fmt.Sprintf("%s=%s", key, e.arguments[key]))
			i++
			if len(e.arguments) != i {
				builder.WriteString("; ")
			}
		}
		builder.WriteString("]")
	}
	if e.errorCode == InternalError && e.parentErr != nil {
		builder.WriteString(": ")
		builder.WriteString(e.parentErr.Error())
	}

	return builder.String()
}

// Is missing godoc
func (e Error) Is(err error) bool {
	if customErr, ok := err.(Error); ok {
		return e.errorCode == customErr.errorCode
	}
	return false
}

// ErrorCode missing godoc
func ErrorCode(err error) ErrorType {
	var customErr Error
	found := errors.As(err, &customErr)
	if found {
		return customErr.errorCode
	}
	return UnknownError
}

// NewNotNullViolationError missing godoc
func NewNotNullViolationError(resourceType resource.Type) error {
	return Error{
		errorCode: EmptyData,
		Message:   EmptyDataMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewCheckViolationError missing godoc
func NewCheckViolationError(resourceType resource.Type) error {
	return Error{
		errorCode: InconsistentData,
		Message:   InconsistentDataMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewOperationTimeoutError missing godoc
func NewOperationTimeoutError() error {
	return Error{
		errorCode: OperationTimeout,
		Message:   OperationTimeoutMsg,
	}
}

// NewNotUniqueError missing godoc
func NewNotUniqueError(resourceType resource.Type) error {
	return Error{
		errorCode: NotUnique,
		Message:   NotUniqueMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewNotUniqueErrorWithMessage constructs a new NotUniqueError with a custom message
func NewNotUniqueErrorWithMessage(resourceType resource.Type, message string) error {
	return Error{
		errorCode: NotUnique,
		Message:   fmt.Sprintf(NotUniqueMsgF, message),
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewNotUniqueNameError missing godoc
func NewNotUniqueNameError(resourceType resource.Type) error {
	return Error{
		errorCode: NotUniqueName,
		Message:   NotUniqueNameMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewNotFoundError missing godoc
func NewNotFoundError(resourceType resource.Type, objectID string) error {
	return Error{
		errorCode: NotFound,
		Message:   NotFoundMsg,
		arguments: map[string]string{"object": string(resourceType), "ID": objectID},
		parentErr: nil,
	}
}

// NewNotFoundErrorWithMessage missing godoc
func NewNotFoundErrorWithMessage(resourceType resource.Type, objectID string, message string) error {
	return Error{
		errorCode: NotFound,
		Message:   fmt.Sprintf(NotFoundMsgF, message),
		arguments: map[string]string{"object": string(resourceType), "ID": objectID},
		parentErr: nil,
	}
}

// NewNotFoundErrorWithType missing godoc
func NewNotFoundErrorWithType(resourceType resource.Type) error {
	return Error{
		errorCode: NotFound,
		Message:   NotFoundMsg,
		arguments: map[string]string{"object": string(resourceType)},
		parentErr: nil,
	}
}

// NewInvalidDataError missing godoc
func NewInvalidDataError(msg string, args ...interface{}) error {
	return Error{
		errorCode: InvalidData,
		Message:   InvalidDataMsg,
		arguments: map[string]string{"reason": fmt.Sprintf(msg, args...)},
	}
}

// NewInvalidDataErrorWithFields missing godoc
func NewInvalidDataErrorWithFields(fields map[string]error, objType string) error {
	if len(fields) == 0 {
		return nil
	}

	err := Error{
		errorCode: InvalidData,
		Message:   fmt.Sprintf("%s %s", InvalidDataMsg, objType),
		arguments: map[string]string{},
	}

	for k, v := range fields {
		err.arguments[k] = v.Error()
	}
	return err
}

// NewInternalError missing godoc
func NewInternalError(msg string, args ...interface{}) error {
	errMsg := fmt.Sprintf(msg, args...)
	return Error{
		errorCode: InternalError,
		Message:   fmt.Sprintf(InternalServerErrMsgF, errMsg),
		arguments: map[string]string{},
	}
}

// InternalErrorFrom missing godoc
func InternalErrorFrom(err error, msg string, args ...interface{}) error {
	errMsg := fmt.Sprintf(msg, args...)
	return Error{
		errorCode: InternalError,
		Message:   fmt.Sprintf(InternalServerErrMsgF, errMsg),
		arguments: map[string]string{},
		parentErr: err,
	}
}

// NewTenantNotFoundError missing godoc
func NewTenantNotFoundError(externalTenant string) error {
	return Error{
		errorCode: TenantNotFound,
		Message:   TenantNotFoundMsg,
		arguments: map[string]string{"externalTenant": externalTenant},
	}
}

// NewTenantRequiredError missing godoc
func NewTenantRequiredError() error {
	return Error{
		errorCode: TenantRequired,
		Message:   TenantRequiredMsg,
		arguments: map[string]string{},
	}
}

// NewInvalidOperationError missing godoc
func NewInvalidOperationError(reason string) error {
	return Error{
		errorCode: InvalidOperation,
		Message:   InvalidOperationMsg,
		arguments: map[string]string{"reason": reason},
	}
}

// NewForeignKeyInvalidOperationError missing godoc
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
		Message:   InvalidOperationMsg,
		arguments: map[string]string{"reason": reason, "object": string(resourceType)},
	}
}

const valueNotFoundInConfigMsg = "value under specified path not found in configuration"

// NewValueNotFoundInConfigurationError missing godoc
func NewValueNotFoundInConfigurationError() error {
	return Error{
		errorCode: NotFound,
		Message:   valueNotFoundInConfigMsg,
		arguments: map[string]string{},
	}
}

// NewNoScopesInContextError missing godoc
func NewNoScopesInContextError() error {
	return Error{
		errorCode: NotFound,
		Message:   NoScopesInContextMsg,
		arguments: map[string]string{},
	}
}

// NewRequiredScopesNotDefinedError missing godoc
func NewRequiredScopesNotDefinedError() error {
	return Error{
		errorCode: InsufficientScopes,
		Message:   NoRequiredScopesInContextMsg,
		arguments: map[string]string{},
	}
}

// NewKeyDoesNotExistError missing godoc
func NewKeyDoesNotExistError(key string) error {
	return Error{
		errorCode: NotFound,
		Message:   KeyDoesNotExistMsg,
		arguments: map[string]string{"key": key},
	}
}

// NewInsufficientScopesError missing godoc
func NewInsufficientScopesError(requiredScopes, actualScopes []string) error {
	return Error{
		errorCode: InsufficientScopes,
		Message:   InsufficientScopesMsg,
		arguments: map[string]string{"required": strings.Join(requiredScopes, ";"),
			"actual": strings.Join(actualScopes, ";")},
		parentErr: nil,
	}
}

// NewCannotReadTenantError missing godoc
func NewCannotReadTenantError() error {
	return Error{
		errorCode: InternalError,
		Message:   CannotReadTenantMsg,
		arguments: map[string]string{},
	}
}

// NewCannotReadClientUserError missing godoc
func NewCannotReadClientUserError() error {
	return Error{
		errorCode: InternalError,
		Message:   CannotReadClientUserMsg,
		arguments: map[string]string{},
	}
}

// NewUnauthorizedError missing godoc
func NewUnauthorizedError(msg string) error {
	return Error{
		errorCode: Unauthorized,
		Message:   UnauthorizedMsg,
		arguments: map[string]string{"reason": msg},
	}
}

// NewConcurrentOperationInProgressError missing godoc
func NewConcurrentOperationInProgressError(msg string) error {
	return Error{
		errorCode: ConcurrentOperation,
		Message:   ConcurrentOperationMsg,
		arguments: map[string]string{"reason": msg},
	}
}

// NewInvalidStatusCondition missing godoc
func NewInvalidStatusCondition(resourceType resource.Type) error {
	return Error{
		errorCode: InvalidStatusCondition,
		Message:   InvalidStatusConditionMsg,
		arguments: map[string]string{"object": string(resourceType)},
	}
}

// NewCannotUpdateObjectInManyBundles missing godoc
func NewCannotUpdateObjectInManyBundles() error {
	return Error{
		errorCode: CannotUpdateObjectInManyBundles,
		Message:   CannotUpdateObjectInManyBundlesMsg,
		arguments: map[string]string{},
	}
}

// NewConcurrentUpdate returns ConcurrentUpdate error
func NewConcurrentUpdate() error {
	return Error{
		errorCode: ConcurrentUpdate,
		Message:   ConcurrentUpdateMsg,
		arguments: map[string]string{},
	}
}

// NewCustomErrorWithCode returns Error with a given code and message
func NewCustomErrorWithCode(code int, msg string) error {
	return Error{
		errorCode: ErrorType(code),
		Message:   msg,
		arguments: map[string]string{},
	}
}

// IsValueNotFoundInConfiguration missing godoc
func IsValueNotFoundInConfiguration(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound && customErr.Message == valueNotFoundInConfigMsg
	}
	return false
}

// IsKeyDoesNotExist missing godoc
func IsKeyDoesNotExist(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound && customErr.Message == KeyDoesNotExistMsg
	}
	return false
}

// IsCannotReadTenant missing godoc
func IsCannotReadTenant(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == InternalError && customErr.Message == CannotReadTenantMsg
	}
	return false
}

// IsConcurrentUpdate indicates if the provided error is thrown in case of concurrent update
func IsConcurrentUpdate(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == Unauthorized && strings.Contains(customErr.Message, UnauthorizedMsg)
	}
	return false
}

// IsNewInvalidOperationError missing godoc
func IsNewInvalidOperationError(err error) bool {
	return ErrorCode(err) == InvalidOperation
}

// IsNotFoundError missing godoc
func IsNotFoundError(err error) bool {
	return ErrorCode(err) == NotFound
}

// IsTenantRequired missing godoc
func IsTenantRequired(err error) bool {
	return ErrorCode(err) == TenantRequired
}

// IsTenantNotFoundError missing godoc
func IsTenantNotFoundError(err error) bool {
	return ErrorCode(err) == TenantNotFound
}

// IsNotUniqueError missing godoc
func IsNotUniqueError(err error) bool {
	return ErrorCode(err) == NotUnique
}

// IsNewNotNullViolationError missing godoc
func IsNewNotNullViolationError(err error) bool {
	return ErrorCode(err) == EmptyData
}

// IsNewCheckViolationError missing godoc
func IsNewCheckViolationError(err error) bool {
	return ErrorCode(err) == InconsistentData
}

// IsInvalidStatusCondition missing godoc
func IsInvalidStatusCondition(err error) bool {
	return ErrorCode(err) == InvalidStatusCondition
}

// IsCannotUpdateObjectInManyBundlesError missing godoc
func IsCannotUpdateObjectInManyBundlesError(err error) bool {
	return ErrorCode(err) == CannotUpdateObjectInManyBundles
}

func sortMapKey(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
