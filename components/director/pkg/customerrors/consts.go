package customerrors

type ErrorType int

const (
	UnhandledError     ErrorType = 10
	InternalError      ErrorType = 11
	NotFound           ErrorType = 20
	NotUnique          ErrorType = 21
	TenantNotFound     ErrorType = 22
	InvalidData        ErrorType = 23
	UnsuffcientScopes  ErrorType = 24
	ConstraintVolation ErrorType = 25
	TenantNotExist     ErrorType = 26
)

type ResourceType string

const (
	Application ResourceType = "Application"
	Runtime     ResourceType = "Runtime"
	//TODO: Add more types like package and etc
)
