package customerrors

//go:generate stringer -type ErrorType
type ErrorType int

const (
	UnknownError        ErrorType = 0
	InternalError       ErrorType = 10
	NotFound            ErrorType = 20
	NotUnique           ErrorType = 21
	InvalidData         ErrorType = 22
	InsufficientScopes  ErrorType = 23
	ConstraintViolation ErrorType = 24
	TenantIsRequired    ErrorType = 25
	TenantNotFound      ErrorType = 26
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
