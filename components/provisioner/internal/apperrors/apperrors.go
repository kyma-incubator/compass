package apperrors

import "fmt"

const (
	CodeInternal   ErrCode = 500
	CodeForbidden  ErrCode = 403
	CodeBadRequest ErrCode = 400
)

type ErrCode int

type AppError interface {
	Append(string, ...interface{}) AppError
	Code() ErrCode
	Error() string
}

type appError struct {
	code    ErrCode
	message string
}

func errorf(code ErrCode, format string, a ...interface{}) AppError {
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

func (ae appError) Code() ErrCode {
	return ae.code
}

func (ae appError) Error() string {
	return ae.message
}
