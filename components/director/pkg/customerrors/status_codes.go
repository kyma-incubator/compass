package customerrors

type StatusCode int

const (
	ExternalError = iota
	InternalError
	NotFound
	NotUnique
	TenantNotFound
	InvalidData
)
