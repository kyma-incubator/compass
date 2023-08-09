package oauth

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"

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

func (h *jwksHandler) Handle(writer http.ResponseWriter, r *http.Request) {
	jwksKey := jwk.NewRSAPublicKey()
	err := jwksKey.FromRaw(h.key)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while parsing key"), http.StatusInternalServerError)
		return
	}

	keySet := jwk.NewSet()
	keySet.Add(jwksKey)

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)

	err = json.NewEncoder(writer).Encode(keySet)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
}
