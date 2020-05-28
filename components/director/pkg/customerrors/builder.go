package customerrors

import "fmt"

type buildStep interface {
	build() error
}

type typeStep interface {
	internalError(msg string) argsStep
	notUnique(resourceType ResourceType) argsStep
	notFound(resourceType ResourceType, resourceID string) argsStep
	invalidData(msg string) argsStep
	tenantNotFound(tenantID string) argsStep
	tenantIsRequired() argsStep
	constraintViolation(resourceType ResourceType) argsStep
	withStatusCode(errType ErrorType) msgStep
	buildStep
}

type argsStep interface {
	with(key, value string) argsStep
	wrapStep
	buildStep
}

type msgStep interface {
	withMessage(msg string) argsStep
	wrapStep
	buildStep
}
type wrapStep interface {
	wrap(err error) buildStep
}

type builder struct {
	args      map[string]string
	errorType ErrorType
	message   string
	parentErr error
}

func newBuilder() typeStep {
	return &builder{args: make(map[string]string)}
}

func (builder *builder) internalError(msg string) argsStep {
	builder.errorType = InternalError
	builder.message = fmt.Sprintf("Internal Server Error: %s", msg)
	return builder
}

func (builder *builder) notFound(resourceType ResourceType, objectID string) argsStep {
	builder.args["object"] = string(resourceType)
	builder.args["ID"] = objectID
	builder.message = "Object not found"
	builder.errorType = NotFound
	return builder
}

func (builder *builder) invalidData(msg string) argsStep {
	builder.args["reason"] = msg
	builder.message = "Invalid input"
	builder.errorType = InvalidData
	return builder
}

func (builder *builder) notUnique(resourceType ResourceType) argsStep {
	builder.args["object"] = string(resourceType)
	builder.message = "Object is not unique"
	builder.errorType = NotUnique
	return builder
}

func (builder *builder) tenantIsRequired() argsStep {
	builder.message = "Tenant is required"
	builder.errorType = TenantIsRequired
	return builder
}

func (builder *builder) tenantNotFound(tenantID string) argsStep {
	builder.message = "Tenant not found"
	builder.args["tenantID"] = tenantID
	builder.errorType = TenantNotFound
	return builder
}

func (builder *builder) constraintViolation(resourceType ResourceType) argsStep {
	builder.message = "Object already exist"
	builder.args["object"] = string(resourceType)
	return builder

}

func (builder *builder) withStatusCode(errorType ErrorType) msgStep {
	builder.errorType = errorType
	return builder
}

func (builder *builder) with(key, value string) argsStep {
	builder.args[key] = value
	return builder
}

func (builder *builder) withMessage(message string) argsStep {
	builder.message = message
	return builder
}

func (builder *builder) wrap(err error) buildStep {
	builder.parentErr = err
	return builder
}

func (builder *builder) build() error {
	return Error{
		errorCode: builder.errorType,
		Message:   builder.message,
		arguments: builder.args,
		parentErr: builder.parentErr,
	}
}
