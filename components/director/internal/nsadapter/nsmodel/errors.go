package nsmodel

import "fmt"

const systemNotFoundErrMsgFormat = "system with subaccount [%s], location id [%s] and host [%s] does not exist"

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
