package authenticator_tnt

import (
	"github.com/form3tech-oss/jwt-go"
)

type Claims struct {
	Scopes []string `json:"scope"`
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
