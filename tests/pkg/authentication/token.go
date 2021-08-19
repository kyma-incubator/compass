package authentication

import (
	"testing"

	"github.com/form3tech-oss/jwt-go"
	"github.com/stretchr/testify/require"
)

type Claims struct {
	Scopes string `json:"scopes"`
	ZID    string `json:"zid"`
	jwt.StandardClaims
}

func CreateNotSingedToken(t *testing.T) string {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		Scopes: "Callback",
		ZID:    "id-zone",
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return signedToken
}
