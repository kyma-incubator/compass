package dberr

import "fmt"

const (
	CodeInternal      = 1
	CodeNotFound      = 2
	CodeAlreadyExists = 3
	CodeConflict      = 4
)

type Error interface {
	Append(string, ...interface{}) Error
	Code() int
	Error() string
}

type dbError struct {
	code    int
	message string
}

func errorf(code int, format string, a ...interface{}) Error {
	return dbError{code: code, message: fmt.Sprintf(format, a...)}
}

func Internal(format string, a ...interface{}) Error {
	return errorf(CodeInternal, format, a...)
}

func NotFound(format string, a ...interface{}) Error {
	return errorf(CodeNotFound, format, a...)
}

func AlreadyExists(format string, a ...interface{}) Error {
	return errorf(CodeAlreadyExists, format, a...)
}

func Conflict(format string, a ...interface{}) Error {
	return errorf(CodeConflict, format, a...)
}

func (e dbError) Append(additionalFormat string, a ...interface{}) Error {
	format := additionalFormat + ", " + e.message
	return errorf(e.code, format, a...)
}

func (e dbError) Code() int {
	return e.code
}

func (e dbError) Error() string {
	return e.message
}
