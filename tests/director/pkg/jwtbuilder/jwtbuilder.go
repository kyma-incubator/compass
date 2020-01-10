package jwtbuilder

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type jwtTokenClaims struct {
	Scopes string `json:"scopes"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

func Do(tenant string, scopes []string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwtTokenClaims{
		Tenant: tenant,
		Scopes: strings.Join(scopes, " "),
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", errors.Wrap(err, "while signing token")
	}

	return signedToken, nil
}
