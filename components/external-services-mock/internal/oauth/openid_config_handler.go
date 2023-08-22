package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
)

type openIDConfigHandler struct {
	baseURL  string
	jwksPath string
}

func NewOpenIDConfigHandler(baseURL, jwksPath string) *openIDConfigHandler {
	return &openIDConfigHandler{
		baseURL:  baseURL,
		jwksPath: jwksPath,
	}
}

func (h *openIDConfigHandler) Handle(writer http.ResponseWriter, r *http.Request) {
	openIDConfig := map[string]interface{}{
		"issuer":   h.baseURL,
		"jwks_uri": h.baseURL + h.jwksPath,
	}

	writer.Header().Set(httphelpers.ContentTypeHeaderKey, httphelpers.ContentTypeApplicationJSON)

	err := json.NewEncoder(writer).Encode(openIDConfig)
	if err != nil {
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
}
