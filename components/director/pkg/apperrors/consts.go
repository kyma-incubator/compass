package apperrors

// ErrorType missing godoc
//go:generate stringer -type ErrorType
type ErrorType int

const (
	// InternalError missing godoc
	InternalError ErrorType = 10
	// UnknownError missing godoc
	UnknownError ErrorType = 11
	// NotFound missing godoc
	NotFound ErrorType = 20
	// NotUnique missing godoc
	NotUnique ErrorType = 21
	// InvalidData missing godoc
	InvalidData ErrorType = 22
	// InsufficientScopes missing godoc
	InsufficientScopes ErrorType = 23
	// TenantRequired missing godoc
	TenantRequired ErrorType = 24
	// TenantNotFound missing godoc
	TenantNotFound ErrorType = 25
	// Unauthorized missing godoc
	Unauthorized ErrorType = 26
	// InvalidOperation missing godoc
	InvalidOperation ErrorType = 27
	// OperationTimeout missing godoc
	OperationTimeout ErrorType = 28
	// EmptyData missing godoc
	EmptyData ErrorType = 29
	// InconsistentData missing godoc
	InconsistentData ErrorType = 30
	// NotUniqueName missing godoc
	NotUniqueName ErrorType = 31
	// ConcurrentOperation missing godoc
	ConcurrentOperation ErrorType = 32
	// InvalidStatusCondition missing godoc
	InvalidStatusCondition ErrorType = 33
	// CannotUpdateObjectInManyBundles missing godoc
	CannotUpdateObjectInManyBundles ErrorType = 34
	// ConcurrentUpdate error code
	ConcurrentUpdate ErrorType = 35
)

const (
	// NotFoundMsg missing godoc
	NotFoundMsg = "Object not found"
	// NotFoundMsgF missing godoc
	NotFoundMsgF = "Object not found: %s"
	// InvalidDataMsg missing godoc
	InvalidDataMsg = "Invalid data"
	// InternalServerErrMsgF missing godoc
	InternalServerErrMsgF = "Internal Server Error: %s"
	// NotUniqueMsg missing godoc
	NotUniqueMsg = "Object is not unique"
	// TenantRequiredMsg missing godoc
	TenantRequiredMsg = "Tenant is required"
	// TenantNotFoundMsg missing godoc
	TenantNotFoundMsg = "Tenant not found"
	// InsufficientScopesMsg missing godoc
	InsufficientScopesMsg = "insufficient scopes provided"
	// NoScopesInContextMsg missing godoc
	NoScopesInContextMsg = "cannot read scopes from context"
	// NoRequiredScopesInContextMsg missing godoc
	NoRequiredScopesInContextMsg = "required scopes are not defined"
	// KeyDoesNotExistMsg missing godoc
	KeyDoesNotExistMsg = "the key does not exist in the source object"
	// CannotReadTenantMsg missing godoc
	CannotReadTenantMsg = "cannot read tenant from context"
	// CannotReadClientUserMsg missing godoc
	CannotReadClientUserMsg = "cannot read client_user from context"
	// InvalidOperationMsg missing godoc
	InvalidOperationMsg = "The operation is not allowed"
	// UnauthorizedMsg missing godoc
	UnauthorizedMsg = "Unauthorized"
	// OperationTimeoutMsg missing godoc
	OperationTimeoutMsg = "operation has timed out"
	// EmptyDataMsg missing godoc
	EmptyDataMsg = "Some required data was left out"
	// InconsistentDataMsg missing godoc
	InconsistentDataMsg = "Inconsistent or out-of-range data"
	// NotUniqueNameMsg missing godoc
	NotUniqueNameMsg = "Object name is not unique"
	// ConcurrentOperationMsg missing godoc
	ConcurrentOperationMsg = "Concurrent operation"
	// InvalidStatusConditionMsg missing godoc
	InvalidStatusConditionMsg = "Invalid status condition"
	// CannotUpdateObjectInManyBundlesMsg missing godoc
	CannotUpdateObjectInManyBundlesMsg = "Can not update object that is part of more than one bundle"
	// ConcurrentUpdateMsg is the error message for NewConcurrentUpdate
	ConcurrentUpdateMsg = "Could not update object due to concurrent update"
	// ShouldUpdateSingleRowButUpdatedMsg is the message which indicates that the query did not update a single row
	ShouldUpdateSingleRowButUpdatedMsg = "should update single row, but updated"
	// ShouldUpdateSingleRowButUpdatedMsgF  is the format of the message for ShouldUpdateSingleRowButUpdatedMsg
	ShouldUpdateSingleRowButUpdatedMsgF = "should update single row, but updated %d rows"
)
