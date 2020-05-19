package customerrors

type ErrorCode int

const (
	UnhandledError ErrorCode = iota
	InternalError
	NotFound
	NotUnique
	TenantNotFound
	InvalidData
	ConstaintVolation
)
