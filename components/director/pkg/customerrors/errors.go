package customerrors

import (
	"fmt"
	"strings"
)

type Error struct {
	errorCode ErrorType
	Message   string
	arguments map[string]string
	parentErr error
}

func (err Error) Error() string {
	builder := strings.Builder{}
	builder.WriteString(err.Message)

	if len(err.arguments) != 0 {
		builder.WriteString(" [")
		for key, value := range err.arguments {
			builder.WriteString(fmt.Sprintf("%s: %s; ", key, value))
		}
		builder.WriteString("] ")
	}
	if err.errorCode == InternalError && err.parentErr != nil {
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

func IsNotFoundErr(err error) bool {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode == NotFound
	} else {
		return false
	}
}

func NewNotUniqueErr(reason string) error {
	return NewBuilder().NotUnique(reason).Build()
}

func NewNotFoundError(objectType ResourceType, objectID string) error {
	return NewBuilder().NotFound(objectType, objectID).Build()
}

func NewInvalidDataError(msg string) error {
	return NewBuilder().InvalidData(msg).Build()
}

func NewInternalError(msg string) error {
	return NewBuilder().InternalError(msg).Build()
}

func NewTenantNotFound(tenantID string) error {
	return NewBuilder().TenantNotFound(tenantID).Build()
}

func GetErrorCode(err error) ErrorType {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode
	} else {
		return UnhandledError
	}
}
