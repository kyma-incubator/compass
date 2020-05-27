package customerrors

import "fmt"

type ArgsStep interface {
	With(key, value string) ArgsStep
	Wrap(err error) BuildStep
	Build() error
}

type MsgStep interface {
	WithMessage(msg string) ArgsStep
	Wrap(err error) BuildStep
	Build() error
}

type TypeStep interface {
	InternalError(msg string) ArgsStep
	NotUnique(reason string) ArgsStep
	NotFound(resourceType ResourceType, resourceID string) ArgsStep
	InvalidData(msg string) ArgsStep
	TenantNotFound(tenantID string) ArgsStep
	WithStatusCode(errType ErrorType) MsgStep
	Build() error
}

type BuildStep interface {
	Build() error
}

type Builder struct {
	args      map[string]string
	errorType ErrorType
	message   string
	parentErr error
}

func NewBuilder() TypeStep {
	return &Builder{args: make(map[string]string)}
}

func (builder *Builder) InternalError(msg string) ArgsStep {
	builder.errorType = InternalError
	builder.message = fmt.Sprintf("Internal Server Error; %s", msg)
	return builder
}

func (builder *Builder) NotFound(objectType ResourceType, objectID string) ArgsStep {
	builder.args["object"] = string(objectType)
	builder.args["ID"] = objectID
	builder.message = "Object not found"
	builder.errorType = NotFound
	return builder
}

func (builder *Builder) InvalidData(msg string) ArgsStep {
	builder.args["reason"] = msg
	builder.message = "Invalid input"
	builder.errorType = InvalidData
	return builder
}

func (builder *Builder) NotUnique(msg string) ArgsStep {
	builder.args["reason"] = msg
	builder.message = "Object is not unique"
	builder.errorType = NotUnique
	return builder
}

func (builder *Builder) TenantNotFound(tenantID string) ArgsStep {
	builder.message = "Tenant not found"
	builder.args["tenantID"] = tenantID
	builder.errorType = TenantNotFound
	return builder
}

func (builder *Builder) WithStatusCode(errorType ErrorType) MsgStep {
	builder.errorType = errorType
	return builder
}

func (builder *Builder) With(key, value string) ArgsStep {
	builder.args[key] = value
	return builder
}

func (builder *Builder) WithMessage(message string) ArgsStep {
	builder.message = message
	return builder
}

func (builder *Builder) Wrap(err error) BuildStep {
	builder.parentErr = err
	return builder
}

func (builder *Builder) Build() error {
	return Error{
		errorCode: builder.errorType,
		Message:   builder.message,
		arguments: builder.args,
		parentErr: builder.parentErr,
	}
}
