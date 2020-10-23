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
)

const (
	notFoundMsg                  = "Object not found"
	invalidDataMsg               = "Invalid data"
	internalServerErrMsgF        = "Internal Server Error: %s"
	notUniqueMsg                 = "Object is not unique"
	tenantRequiredMsg            = "Tenant is required"
	tenantNotFoundMsg            = "Tenant not found"
	insufficientScopesMsg        = "insufficient scopes provided"
	noScopesInContextMsg         = "cannot read scopes from context"
	noRequiredScopesInContextMsg = "required scopes are not defined"
	keyDoesNotExistMsg           = "the key does not exist in the source object"
	cannotReadTenantMsg          = "cannot read tenant from context"
	invalidOperationMsg          = "The operation is not allowed"
	unauthorizedMsg              = "Unauthorized"
	operationTimeoutMsg          = "operation has timed out"
	emptyDataMsg                 = "Some required data was left out"
	inconsistentDataMsg          = "Inconsistent or out-of-range data"
)

type SQLOperation string

const (
	Create SQLOperation = "Create"
	Update SQLOperation = "Update"
	Upsert SQLOperation = "Upsert"
	Delete SQLOperation = "Delete"
	Exists SQLOperation = "Exists"
	Get    SQLOperation = "Get"
	List   SQLOperation = "List"
)
