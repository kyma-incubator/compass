package nsmodel

import (
	"fmt"
	"strings"
)

const systemNotFoundErrMsgFormat = "system with subaccount [%s], location id [%s] and host [%s] does not exist"

//TODO check if needed
type SystemNotFoundError struct {
	msg string
}

func IsNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "Object not found")
}

func NewSystemNotFoundError(subaccount, locationId, host string) *SystemNotFoundError {
	return &SystemNotFoundError{msg: fmt.Sprintf(systemNotFoundErrMsgFormat, subaccount, locationId, host)}
}

func (a *SystemNotFoundError) Error() string {
	return a.msg
}
