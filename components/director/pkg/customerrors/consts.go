package customerrors

//go:generate stringer -type ErrorType
type ErrorType int

const (
	UnhandledError      ErrorType = 10
	InternalError       ErrorType = 11
	NotFound            ErrorType = 20
	NotUnique           ErrorType = 21
	InvalidData         ErrorType = 22
	InsufficientScopes  ErrorType = 23
	ConstraintViolation ErrorType = 24
	TenantNotFound      ErrorType = 25
	TenantNotExist      ErrorType = 26
)

type ResourceType string

const (
	Application         ResourceType = "Application"
	ApplicationTemplate ResourceType = "ApplicationTemplate"
	Runtime             ResourceType = "Runtime"
	LabelDefinition     ResourceType = "LabelDefinition"
	Package             ResourceType = "Package"
	IntegrationSystem   ResourceType = "IntegrationSystem"
)
