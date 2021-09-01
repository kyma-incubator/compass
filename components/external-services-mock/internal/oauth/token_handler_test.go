package oauth_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/form3tech-oss/jwt-go"
	oauth2 "github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Generate(t *testing.T) {
	//GIVEN
	secret, id := "secret", "id"
	req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", strings.NewReader(""))

	encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
	req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))

	h := oauth2.NewHandler(secret, id)
	r := httptest.NewRecorder()

	//WHEN
	h.Generate(r, req)
	resp := r.Result()

	//THEN
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	var response oauth2.TokenResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
}

func TestHandler_GenerateWithSigningKey(t *testing.T) {
	//GIVEN
	expectedClaims := map[string]interface{}{
		"test-claim": "test-value",
	}
	claimsBody, err := json.Marshal(expectedClaims)
	require.NoError(t, err)

	secret, id := "secret", "id"
	req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer(claimsBody))

	encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
	req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	h := oauth2.NewHandlerWithSigningKey(secret, id, key)
	r := httptest.NewRecorder()

	//WHEN
	h.Generate(r, req)
	resp := r.Result()

	//THEN
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)

	var response oauth2.TokenResponse
	err = json.Unmarshal(body, &response)

	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)

	claims := map[string]interface{}{}

	_, err = jwt.ParseWithClaims(response.AccessToken, jwt.MapClaims(claims), func(token *jwt.Token) (interface{}, error) {
		return key.Public(), nil
	})

	require.NoError(t, err)
	require.Equal(t, expectedClaims, claims)
}
