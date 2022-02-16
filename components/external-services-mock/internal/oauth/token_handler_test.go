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
	"net/url"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/form3tech-oss/jwt-go"
	oauth2 "github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/stretchr/testify/require"
)

func TestHandler_Generate(t *testing.T) {
	//GIVEN
	data := url.Values{}
	data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)

	id, secret := "id", "secret"
	req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", strings.NewReader(data.Encode()))

	encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
	req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))
	req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

	h := oauth2.NewHandler(secret, id)
	r := httptest.NewRecorder()

	//WHEN
	h.Generate(r, req)
	resp := r.Result()

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	var response oauth2.TokenResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response.AccessToken)
}

func TestHandler_GenerateWithSigningKey(t *testing.T) {
	t.Run("Failed generate token if the content type is not application URL encoded", func(t *testing.T) {
		// GIVEN
		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte{}))
		handler := oauth2.NewHandlerWithSigningKey("", "", "", "", "", nil, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		handler.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("Failed generate token if the grant type is not client_credentials or password", func(t *testing.T) {
		// GIVEN
		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, "invalid")

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)
		handler := oauth2.NewHandlerWithSigningKey("", "", "", "", "", nil, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		handler.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Successfully issue client_credentials token with client_id and client_secret as authorization header", func(t *testing.T) {
		//GIVEN
		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)

		id, secret, tenantHeader := "id", "secret", "x-zid"
		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, "", "", key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)

		var response oauth2.TokenResponse
		err = json.Unmarshal(body, &response)

		require.NoError(t, err)
		require.NotEmpty(t, response.AccessToken)
	})

	t.Run("Failed issuing client_credentials token if the client_id or client_secret as part of authorization header does not match the expected one", func(t *testing.T) {
		//GIVEN
		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)

		id, secret, tenantHeader := "id", "secret", "x-zid"
		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		expectedSecret := "expectedSecret"
		expectedID := "expectedID"
		h := oauth2.NewHandlerWithSigningKey(expectedSecret, expectedID, tenantHeader, "", "", key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		bodyErr := gjson.GetBytes(body, "error")
		require.True(t, bodyErr.Exists())
		require.NotEmpty(t, bodyErr)
		require.Equal(t, bodyErr.String(), fmt.Sprintf("client_id or client_secret from authorization header doesn't match the expected one. Expected: %s and %s but we got: %s and %s", expectedID, expectedSecret, id, secret))
	})

	t.Run("Successfully issue client_credentials token with client_id and client_secret as part of the request body", func(t *testing.T) {
		//GIVEN
		id, secret, tenantHeader := "id", "secret", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.ClientSecretKey, secret)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, "", "", key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var response oauth2.TokenResponse
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)
		require.NotEmpty(t, response.AccessToken)
	})

	t.Run("Failed issuing client_credentials token if the client_id or client_secret as part of the request body does not match the expected one", func(t *testing.T) {
		//GIVEN
		id, secret, tenantHeader := "id", "secret", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.ClientSecretKey, secret)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		expectedSecret := "expectedSecret"
		expectedID := "expectedID"
		h := oauth2.NewHandlerWithSigningKey(expectedSecret, expectedID, tenantHeader, "", "", key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		bodyErr := gjson.GetBytes(body, "error")
		require.True(t, bodyErr.Exists())
		require.NotEmpty(t, bodyErr)
		require.Equal(t, bodyErr.String(), fmt.Sprintf("client_id or client_secret from request body doesn't match the expected one. Expected: %s and %s but we got: %s and %s", expectedID, expectedSecret, id, secret))
	})

	t.Run("Successfully issue user token", func(t *testing.T) {
		//GIVEN
		id, secret, username, password, tenantHeader := "id", "secret", "username", "password", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.PasswordGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.UserNameKey, username)
		data.Add(oauth2.PasswordKey, password)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, username, password, key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var response oauth2.TokenResponse
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)
		require.NotEmpty(t, response.AccessToken)
	})

	t.Run("Failed issuing user token if authorization header is missing", func(t *testing.T) {
		//GIVEN
		id, secret, username, password, tenantHeader := "id", "secret", "username", "password", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.PasswordGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.UserNameKey, username)
		data.Add(oauth2.PasswordKey, password)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, username, password, key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		bodyErr := gjson.GetBytes(body, "error")
		require.True(t, bodyErr.Exists())
		require.NotEmpty(t, bodyErr)
		require.Contains(t, bodyErr.String(), "missing authorization header")
	})

	t.Run("Failed issuing user token if client_id or client_secret does not match the expected one", func(t *testing.T) {
		//GIVEN
		id, secret, username, password, tenantHeader := "id", "secret", "username", "password", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.PasswordGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.UserNameKey, username)
		data.Add(oauth2.PasswordKey, password)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		expectedSecret := "expectedSecret"
		expectedID := "expectedID"
		h := oauth2.NewHandlerWithSigningKey(expectedSecret, expectedID, tenantHeader, username, password, key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		bodyErr := gjson.GetBytes(body, "error")
		require.True(t, bodyErr.Exists())
		require.NotEmpty(t, bodyErr)
		require.Equal(t, bodyErr.String(), fmt.Sprintf("client_id or client_secret doesn't match the expected one. Expected: %s and %s but we got: %s and %s", expectedID, expectedSecret, id, secret))
	})

	t.Run("Failed issuing user token if username or password does not match the expected one", func(t *testing.T) {
		//GIVEN
		id, secret, username, password, tenantHeader := "id", "secret", "username", "password", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.PasswordGrantType)
		data.Add(oauth2.ClientIDKey, id)
		data.Add(oauth2.UserNameKey, username)
		data.Add(oauth2.PasswordKey, password)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		expectedUsername := "expectedUsername"
		expectedPassword := "expectedPassword"
		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, expectedUsername, expectedPassword, key, map[string]oauth2.ClaimsGetterFunc{})
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		bodyErr := gjson.GetBytes(body, "error")
		require.True(t, bodyErr.Exists())
		require.NotEmpty(t, bodyErr)
		require.Equal(t, bodyErr.String(), fmt.Sprintf("username or password doesn't match the expected one. Expected: %s and %s but we got: %s and %s", expectedUsername, expectedPassword, username, password))
	})

	t.Run("Successfully generate client_credentials token with provided claims_key", func(t *testing.T) {
		//GIVEN
		expectedClaims := map[string]interface{}{
			"test-claim": "test-value",
			"x-zid":      "",
		}
		staticClaimsMapping := map[string]oauth2.ClaimsGetterFunc{
			"tenantFetcherClaims": func() map[string]interface{} {
				return expectedClaims
			},
		}
		id, secret, tenantHeader := "id", "secret", "x-zid"

		data := url.Values{}
		data.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)

		req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", bytes.NewBuffer([]byte(data.Encode())))
		encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))
		req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)

		q := req.URL.Query()
		q.Add(oauth2.ClaimsKey, "tenantFetcherClaims")
		req.URL.RawQuery = q.Encode()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		h := oauth2.NewHandlerWithSigningKey(secret, id, tenantHeader, "", "", key, staticClaimsMapping)
		r := httptest.NewRecorder()

		//WHEN
		h.Generate(r, req)
		resp := r.Result()

		//THEN
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var response oauth2.TokenResponse
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)
		require.NotEmpty(t, response.AccessToken)

		claims := map[string]interface{}{}
		_, err = jwt.ParseWithClaims(response.AccessToken, jwt.MapClaims(claims), func(token *jwt.Token) (interface{}, error) {
			return key.Public(), nil
		})
		require.NoError(t, err)
		require.Equal(t, expectedClaims, claims)
	})
}

func TestHandler_GenerateWithoutCredentialsWithSigningKeyForm(t *testing.T) {
	//GIVEN
	scopes, id, tenantHeader := "scopes", "id", "x-zid"

	form := url.Values{}
	form.Add(oauth2.GrantTypeFieldName, oauth2.CredentialsGrantType)
	form.Add(oauth2.ClientIDKey, id)
	form.Add(oauth2.ScopesFieldName, scopes)

	body := strings.NewReader(form.Encode())
	req, err := http.NewRequest(http.MethodPost, "http://target.com/oauth/token", body)
	require.NoError(t, err)
	req.Header.Set(tenantHeader, "tenant1")
	req.Header.Set(oauth2.ContentTypeHeader, oauth2.ContentTypeApplicationURLEncoded)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	h := oauth2.NewHandlerWithSigningKey("", id, tenantHeader, "", "", key, map[string]oauth2.ClaimsGetterFunc{})
	r := httptest.NewRecorder()

	//WHEN
	h.GenerateWithCredentialsFromReqBody(r, req)
	resp := r.Result()

	//THEN
	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var response oauth2.TokenResponse
	err = json.Unmarshal(respBody, &response)

	require.NoError(t, err)
	require.NotEmpty(t, response.AccessToken)

	claims := &oauth2.Claims{}

	_, err = jwt.ParseWithClaims(response.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return key.Public(), nil
	})

	require.NoError(t, err)
	require.Equal(t, "tenant1", claims.Tenant)
	require.Equal(t, id, claims.ClientId)
	require.Equal(t, scopes, claims.Scopes)
}
