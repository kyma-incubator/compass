package nsmodel

import "fmt"

const systemNotFoundErrMsgFormat = "system with subaccount [%s], location id [%s] and host [%s] does not exist"

type NSError struct {
	message string
}

func NewNSError(msg string) *NSError {
	return &NSError{
		message: msg,
	}
}

func (m *NSError) Error() string {
	return m.message
}

type SystemNotFoundError struct {
	msg string
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*SystemNotFoundError)
	return ok
}

func NewSystemNotFoundError(subaccount, locationId, host string) *SystemNotFoundError {
	return &SystemNotFoundError{msg: fmt.Sprintf(systemNotFoundErrMsgFormat, subaccount, locationId, host)}
}

func (a *SystemNotFoundError) Error() string {
	return a.msg
}
