package oauth

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type jwksHandler struct {
	key *rsa.PublicKey
}

func NewJWKSHandler(key *rsa.PublicKey) *jwksHandler {
	return &jwksHandler{
		key: key,
	}
}

type JWKSResponse struct {
	Keys []Key
}

type Key struct {
	Kty string
	Alg string
	Use string
	Kid string
	*rsa.PublicKey
}

func (h *jwksHandler) Handle(writer http.ResponseWriter, r *http.Request) {
	resp := JWKSResponse{
		Keys: []Key{
			{
				Kty:       "RSA",
				Alg:       "RS256",
				Use:       "sig",
				Kid:       "key-id",
				PublicKey: h.key,
			},
		},
	}

	writer.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(writer).Encode(resp)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
}
