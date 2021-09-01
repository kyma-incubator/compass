package oauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenIDConfigHandler(t *testing.T) {
	//GIVEN
	baseURL := "http://base:8080"
	jwksPath := "/jwks.json"

	expectedConfig := map[string]interface{}{
		"issuer":   baseURL,
		"jwks_uri": baseURL + jwksPath,
	}

	req := httptest.NewRequest(http.MethodGet, "http://target.com/.well-known/openid-configuration", nil)

	h := oauth.NewOpenIDConfigHandler(baseURL, jwksPath)
	r := httptest.NewRecorder()

	//WHEN
	h.Handle(r, req)
	resp := r.Result()

	//THEN
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	config := map[string]interface{}{}
	err := json.NewDecoder(resp.Body).Decode(&config)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, config)
}
