package authenticator

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/pkg/errors"
)

type invalidTenantError struct{}

func (err *invalidTenantError) Error() string {
	return "invalid tenant"
}

func isInvalidTenantError(err error) bool {
	if cause := errors.Cause(err); cause != nil {
		err = cause
	}
	_, ok := err.(*invalidTenantError)
	return ok
}

type Claims struct {
	Tenant       string                `json:"tenant"`
	Scopes       string                `json:"scopes"`
	ConsumerID   string                `json:"consumerID"`
	ConsumerType consumer.ConsumerType `json:"consumerType"`
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
