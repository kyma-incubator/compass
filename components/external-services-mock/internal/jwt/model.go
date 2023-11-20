package jwt

import (
	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"
)

type Claims struct {
	CustomerID string `json:"customerId,omitempty"`
	jwt.StandardClaims
}

func (c Claims) GetTenant() string {
	return c.CustomerID
}

func (c Claims) Valid() error {
	if len(c.CustomerID) == 0 {
		return errors.New("CustomerID should not be null")
	}

	return nil
}
