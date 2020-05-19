package customerrors

import (
	"fmt"
	"strings"
)

type Error struct {
	errorCode ErrorCode
	Message   string
	arguments map[string]string
	parentErr error
	//isInterval bool
}

func (err Error) Error() string {
	builder := strings.Builder{}
	builder.WriteString(err.Message)
	builder.WriteString("[")
	for key, value := range err.arguments {
		builder.WriteString(fmt.Sprintf("%s: %s;", key, value))
	}
	builder.WriteString("]")
	if err.errorCode == InternalError {
		builder.WriteString(": ")
		builder.WriteString(err.parentErr.Error())
	}
	return builder.String()
}

func (e Error) Is(err error) bool {
	if customErr, ok := err.(Error); ok {
		return e.errorCode == customErr.errorCode
	} else {
		return false
	}
}

//TODO: We can provide more constructors for different types and use format prepared messages in contructors.
func NewNotUniqueErr() Error {
	err := Error{errorCode: NotUnique}
	return err
}

func GetErrorCode(err error) ErrorCode {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode
	} else {
		return UnhandledError
	}
}

func IsNotFoundErr(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound
	} else {
		return false
	}
}

type GraphqlError struct {
	StatusCode ErrorCode `json:"status_code"`
	Message    string    `json:"message"`
}

func (g GraphqlError) Error() string {
	return g.Message
}
