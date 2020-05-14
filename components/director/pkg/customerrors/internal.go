package customerrors

import (
	"fmt"
	"strings"
)

type Error struct {
	statusCode StatusCode
	Message    string
	arguments  map[string]string
	parentErr  error
	//isInterval bool
}

func (err Error) Error() string {
	builder := strings.Builder{}
	builder.WriteString("[")
	for key, value := range err.arguments {
		builder.WriteString(fmt.Sprintf("%s: %s;", key, value))
	}
	builder.WriteString("]")
	return builder.String()
}


func (e Error) Is(err error) bool {
	if customErr, ok := err.(Error); ok {
		return e.statusCode == customErr.statusCode
	} else {
		return false
	}
}

var (
	NotFoundErr  = Error{statusCode: NotFound}
	NotUniqueErr = Error{statusCode: NotUnique}
	InternalErr  = Error{statusCode: InternalError}
)

type ErrorBuilder struct {
	args      map[string]string
	errorType StatusCode
}

func NewErrorBuilder(errorType StatusCode) *ErrorBuilder {
	return &ErrorBuilder{
		args:      make(map[string]string),
		errorType: errorType,
	}
}

func (builder *ErrorBuilder) With(key, value string) *ErrorBuilder {
	builder.args[key] = value
	return builder
}

func (b *ErrorBuilder) Build() error {
	return NewError(b.errorType, b.args)
}

func NewError(code StatusCode, args map[string]string) Error {
	return Error{
		statusCode: code,
		Message:    "",
		arguments:  args,
		parentErr:  nil,
	}
}

func NewNotFoundErr(objectType, objectID string) Error {
	return Error{statusCode: NotFoundErr.statusCode, Message: fmt.Sprintf(NotFoundErr.Message, objectType, objectID)}
}

func NewNotUniqueErr() Error {
	err := Error{statusCode: NotUnique}
	return err
}


func ErrorCode(err error) StatusCode {
	if customErr, ok := err.(Error); ok {
		return customErr.statusCode
	} else {
		return ExternalError
	}
}
