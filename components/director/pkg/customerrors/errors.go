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

func NewNotUniqueErr(resourceType ResourceType) error {
	return newBuilder().notUnique(resourceType).build()
}

func NewNotFoundError(objectType ResourceType, objectID string) error {
	return newBuilder().notFound(objectType, objectID).build()
}

func NewInvalidDataError(msg string) error {
	return newBuilder().invalidData(msg).build()
}

func NewInternalError(msg string) error {
	return newBuilder().internalError(msg).build()
}

func InternalErrorFrom(msg string, err error) error {
	return newBuilder().internalError(msg).wrap(err).build()
}

func NewTenantNotFound(tenantID string) error {
	return newBuilder().tenantNotFound(tenantID).build()
}

func GetErrorCode(err error) ErrorType {
	if customErr, ok := err.(Error); ok {
		return customErr.errorCode
	} else {
		return UnknownError
	}
}
