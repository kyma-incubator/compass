package apperrors

//go:generate stringer -type ErrorType
type ErrorType int

const (
	InternalError      ErrorType = 10
	UnknownError       ErrorType = 11
	NotFound           ErrorType = 20
	NotUnique          ErrorType = 21
	InvalidData        ErrorType = 22
	InsufficientScopes ErrorType = 23
	TenantRequired     ErrorType = 24
	TenantNotFound     ErrorType = 25
	Unauthorized       ErrorType = 26
	InvalidOperation   ErrorType = 27
	OperationTimeout   ErrorType = 28
	EmptyData          ErrorType = 29
	InconsistentData   ErrorType = 30
	NotUniqueName      ErrorType = 31
)

const (
	NotFoundMsg                  = "Object not found"
	InvalidDataMsg               = "Invalid data"
	InternalServerErrMsgF        = "Internal Server Error: %s"
	NotUniqueMsg                 = "Object is not unique"
	TenantRequiredMsg            = "Tenant is required"
	TenantNotFoundMsg            = "Tenant not found"
	InsufficientScopesMsg        = "insufficient scopes provided"
	NoScopesInContextMsg         = "cannot read scopes from context"
	NoRequiredScopesInContextMsg = "required scopes are not defined"
	KeyDoesNotExistMsg           = "the key does not exist in the source object"
	CannotReadTenantMsg          = "cannot read tenant from context"
	CannotReadClientUserMsg      = "cannot read client_user from context"
	InvalidOperationMsg          = "The operation is not allowed"
	UnauthorizedMsg              = "Unauthorized"
	OperationTimeoutMsg          = "operation has timed out"
	EmptyDataMsg                 = "Some required data was left out"
	InconsistentDataMsg          = "Inconsistent or out-of-range data"
	NotUniqueNameMsg             = "Object name is not unique"
)
