package apperrors

import "fmt"

const (
	CodeInternal   errCode = 500
	CodeForbidden  errCode = 403
	CodeBadRequest errCode = 400
)

type errCode int

type AppError interface {
	Append(string, ...interface{}) AppError
	Code() errCode
	Error() string
}

type appError struct {
	code    errCode
	message string
}

func errorf(code errCode, format string, a ...interface{}) AppError {
	return appError{code: code, message: fmt.Sprintf(format, a...)}
}

func Internal(format string, a ...interface{}) AppError {
	return errorf(CodeInternal, format, a...)
}

func Forbidden(format string, a ...interface{}) AppError {
	return errorf(CodeForbidden, format, a...)
}

func BadRequest(format string, a ...interface{}) AppError {
	return errorf(CodeBadRequest, format, a...)
}

func (ae appError) Append(additionalFormat string, a ...interface{}) AppError {
	format := additionalFormat + ", " + ae.message
	return errorf(ae.code, format, a...)
}

func (ae appError) Code() errCode {
	return ae.code
}

func (ae appError) Error() string {
	return ae.message
}
