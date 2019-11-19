package authenticator

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Tenant     string `json:"tenant"`
	Scopes     string `json:"scopes"`
	ObjectID   string `json:"objectID"`
	ObjectType string `json:"objectType"`
	*jwt.StandardClaims
}

func (c Claims) Valid() error {
	if c.Tenant == "" {
		return errors.New("Tenant cannot be empty")
	}

	return nil
}
