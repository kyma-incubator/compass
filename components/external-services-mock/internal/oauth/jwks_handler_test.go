package oauth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWKSHandler(t *testing.T) {
	//GIVEN
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://target.com/jwks.json", nil)

	h := oauth.NewJWKSHandler(&key.PublicKey)
	r := httptest.NewRecorder()

	//WHEN
	h.Handle(r, req)
	resp := r.Result()

	//THEN
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	response := jwk.NewSet()
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	jwksKey := jwk.NewRSAPublicKey()
	require.NoError(t, jwksKey.FromRaw(&key.PublicKey))

	expectedKeySet := jwk.NewSet()
	expectedKeySet.Add(jwksKey)

	require.Equal(t, expectedKeySet, response)
}
