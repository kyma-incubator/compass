package errors

import (
	"errors"
	"fmt"
)

var (
	As   = errors.As
	Is   = errors.Is
	New  = errors.New
	Newf = fmt.Errorf
	Join = errors.Join
)

var (
	EntityNotFound         = New("entity not found")
	EntityAlreadyExists    = New("entity already exists")
	Internal               = New("internal error")
	InvalidAccessToken     = New("invalid access token")
	IASApplicationNotFound = New("application in IAS not found")
	S4CertificateNotFound  = New("S/4 certificate not found")
)
