package customerrors

type ErrorBuilder struct {
	args      map[string]string
	errorType ErrorCode
	message   string
	parentErr error
}

func NewErrorBuilder(errorCode ErrorCode) *ErrorBuilder {
	return &ErrorBuilder{
		args:      make(map[string]string),
		errorType: errorCode,
	}
}

func (builder *ErrorBuilder) With(key, value string) *ErrorBuilder {
	builder.args[key] = value
	return builder
}

func (builder *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	builder.message = message
	return builder
}

func (builder *ErrorBuilder) Wrap(err error) *ErrorBuilder {
	builder.parentErr = err
	return builder
}

func (builder *ErrorBuilder) Build() error {
	return Error{
		errorCode: builder.errorType,
		Message:   builder.message,
		arguments: builder.args,
		parentErr: builder.parentErr,
	}
}
