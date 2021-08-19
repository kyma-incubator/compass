package authenticator

import (
	"github.com/form3tech-oss/jwt-go"
)

type Claims struct {
	Scopes []string `json:"scopes"`
	ZID    string   `json:"zid"`
	jwt.StandardClaims
}

func (c Claims) Valid() error {
	err := c.StandardClaims.Valid()
	if err != nil {
		return err
	}

	return nil
}
