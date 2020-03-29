package error

import "fmt"

type TemporaryError struct {
	message string
}

func NewTemporaryError(message string, err error) *TemporaryError {
	var msg string
	if err != nil {
		msg = fmt.Sprintf("%s: %s", message, err.Error())
	} else {
		msg = fmt.Sprintf("%s", message)
	}

	return &TemporaryError{message: msg}
}

func (te TemporaryError) Error() string { return te.message }
func (TemporaryError) Temporary() bool  { return true }

func IsTemporaryError(err error) bool {
	nfe, ok := err.(interface {
		Temporary() bool
	})
	return ok && nfe.Temporary()
}

func Wrapf(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	switch {
	case IsTemporaryError(err):
		return TemporaryError{message: fmt.Sprintf("%s: %s", err.Error(), msg)}
	default:
		return fmt.Errorf("%s: %s", err.Error(), msg)
	}
}
