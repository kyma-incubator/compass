package authenticator

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type invalidTenantError struct {}

func (err *invalidTenantError) Error() string {
	return "Invalid tenant"
}

func isInvalidTenantError(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}
	_, ok := err.(*invalidTenantError)
	return ok
}

type Claims struct {
	Tenant string `json:"tenant"`
	Scopes string `json:"scopes"`
	jwt.StandardClaims
}

func (c Claims) Valid() error {
	err := c.StandardClaims.Valid()
	if err != nil {
		return err
	}

	if c.Tenant == "" {
		return &invalidTenantError{}
	}

	return nil
}
