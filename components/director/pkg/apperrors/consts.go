package apperrors

// ErrorType represents an integer error code.
//go:generate stringer -type ErrorType
type ErrorType int

const (
	// InternalError is the error code for Internal errors.
	InternalError ErrorType = 10
	// UnknownError is the error code for Unknown errors.
	UnknownError ErrorType = 11
	// NotFound is the error code for NotFound errors.
	NotFound ErrorType = 20
	// NotUnique is the error code for NotUnique errors.
	NotUnique ErrorType = 21
	// InvalidData is the error code for InvalidData errors.
	InvalidData ErrorType = 22
	// InsufficientScopes is the error code for InsufficientScopes errors.
	InsufficientScopes ErrorType = 23
	// TenantRequired is the error code for TenantRequired errors.
	TenantRequired ErrorType = 24
	// TenantNotFound is the error code for TenantNotFound errors.
	TenantNotFound ErrorType = 25
	// Unauthorized is the error code for Unauthorized errors.
	Unauthorized ErrorType = 26
	// InvalidOperation is the error code for InvalidOperation errors.
	InvalidOperation ErrorType = 27
	// OperationTimeout is the error code for Timeout errors.
	OperationTimeout ErrorType = 28
	// EmptyData is the error code for EmptyData errors.
	EmptyData ErrorType = 29
	// InconsistentData is the error code for InconsistentData errors.
	InconsistentData ErrorType = 30
	// NotUniqueName is the error code for NotUniqueName errors.
	NotUniqueName ErrorType = 31
	// ConcurrentOperation is the error code for ConcurrentOperation errors.
	ConcurrentOperation ErrorType = 32
	// InvalidStatusCondition is the error code for InvalidStatusCondition errors.
	InvalidStatusCondition ErrorType = 33
	// CannotUpdateObjectInManyBundles is the error code for CannotUpdateObjectInManyBundles errors.
	CannotUpdateObjectInManyBundles ErrorType = 34
	// ConcurrentUpdate is the error code for ConcurrentUpdate errors.
	ConcurrentUpdate ErrorType = 35
	// BadRequest is the error code for BadRequest errors.
	BadRequest ErrorType = 400
	// Conflict is the error code for Conflict errors.
	Conflict ErrorType = 409
)

const (
	// NotFoundMsg is the error message for NotFound errors.
	NotFoundMsg = "Object not found"
	// NotFoundMsgF is the error message format for NotFound errors.
	NotFoundMsgF = "Object not found: %s"
	// InvalidDataMsg is the error message for InvalidData errors.
	InvalidDataMsg = "Invalid data"
	// InternalServerErrMsgF is the error message format for InternalServer errors.
	InternalServerErrMsgF = "Internal Server Error: %s"
	// NotUniqueMsg is the error message for NotUnique errors.
	NotUniqueMsg = "Object is not unique"
	// NotUniqueMsgF is the error message format for NotUnique errors with custom message.
	NotUniqueMsgF = "Object is not unique: %s"
	// TenantRequiredMsg is the error message for TenantRequired errors.
	TenantRequiredMsg = "Tenant is required"
	// TenantNotFoundMsg is the error message for TenantNotFound errors.
	TenantNotFoundMsg = "Tenant not found"
	// InsufficientScopesMsg is the error message for InsufficientScopes errors.
	InsufficientScopesMsg = "insufficient scopes provided"
	// NoScopesInContextMsg is the error message for NoScopesInContext errors.
	NoScopesInContextMsg = "cannot read scopes from context"
	// NoRequiredScopesInContextMsg is the error message for NoRequiredScopesInContext errors.
	NoRequiredScopesInContextMsg = "required scopes are not defined"
	// KeyDoesNotExistMsg is the error message for KeyDoesNotExist errors.
	KeyDoesNotExistMsg = "the key does not exist in the source object"
	// CannotReadTenantMsg is the error message for CannotReadTenant errors.
	CannotReadTenantMsg = "cannot read tenant from context"
	// CannotReadClientUserMsg is the error message for CannotReadClientUser errors.
	CannotReadClientUserMsg = "cannot read client_user from context"
	// InvalidOperationMsg is the error message for InvalidOperation errors.
	InvalidOperationMsg = "The operation is not allowed"
	// UnauthorizedMsg is the error message for Unauthorized errors.
	UnauthorizedMsg = "Unauthorized"
	// OperationTimeoutMsg is the error message for Timeout errors.
	OperationTimeoutMsg = "operation has timed out"
	// EmptyDataMsg is the error message for EmptyData errors.
	EmptyDataMsg = "Some required data was left out"
	// InconsistentDataMsg is the error message for InconsistentData errors.
	InconsistentDataMsg = "Inconsistent or out-of-range data"
	// NotUniqueNameMsg is the error message for NotUniqueName errors.
	NotUniqueNameMsg = "Object name is not unique"
	// ConcurrentOperationMsg is the error message for ConcurrentOperation errors.
	ConcurrentOperationMsg = "Concurrent operation"
	// InvalidStatusConditionMsg is the error message for InvalidStatusCondition errors.
	InvalidStatusConditionMsg = "Invalid status condition"
	// CannotUpdateObjectInManyBundlesMsg is the error message for CannotUpdateObjectInManyBundles errors.
	CannotUpdateObjectInManyBundlesMsg = "Can not update object that is part of more than one bundle"
	// ConcurrentUpdateMsg is the error message for NewConcurrentUpdate errors.
	ConcurrentUpdateMsg = "Could not update object due to concurrent update"
	// ShouldUpdateSingleRowButUpdatedMsgF  is the error message for ShouldUpdateSingleRowButUpdated errors.
	ShouldUpdateSingleRowButUpdatedMsgF = "should update single row, but updated %d rows"
	// ShouldUpsertSingleRowButUpsertedMsgF  is the error message returned when upsert resulted in rows affected count not equal to 1.
	ShouldUpsertSingleRowButUpsertedMsgF = "should upsert single row, but upserted %d rows"
	// ShouldBeOwnerMsg is the error message for unauthorized due to missing owner access errors.
	ShouldBeOwnerMsg = "Owner access is needed for resource modification"
)
