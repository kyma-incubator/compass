package authenticator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/automock"
	"github.com/stretchr/testify/mock"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/stretchr/testify/require"
)

const (
	AuthorizationHeaderKey = "Authorization"
	ClientIDHeaderKey      = "client_user"

	defaultTenant   = "af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e"
	PrivateJWKSURL  = "file://testdata/jwks-private.json"
	PrivateJWKS2URL = "file://testdata/jwks-private2.json"
	PrivateJWKS3URL = "file://testdata/jwks-private3.json"
	PublicJWKSURL   = "file://testdata/jwks-public.json"
	PublicJWKS2URL  = "file://testdata/jwks-public2.json"
	PublicJWKS3URL  = "file://testdata/jwks-public3.json"
	fakeJWKSURL     = "file://testdata/invalid.json"
)

type mockRoundTripper struct{}

func (rt *mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusFound,
		Header: map[string][]string{
			"Location": {"somewhere.else.gone"},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString("")),
	}, nil
}

var httpClientWithoutRedirectsWithMockTransport = &http.Client{
	Transport: &mockRoundTripper{},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var httpClientWithoutRedirects = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func TestAuthenticator_SynchronizeJWKS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, true, ClientIDHeaderKey, claimsValidatorMock())
		// WHEN
		err := auth.SynchronizeJWKS(context.TODO())

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when can't fetch JWKS", func(t *testing.T) {
		// GIVEN
		authFake := authenticator.New(httpClientWithoutRedirects, fakeJWKSURL, true, ClientIDHeaderKey, nil)

		// WHEN
		err := authFake.SynchronizeJWKS(context.TODO())

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while fetching JWKS from endpoint %s: failed to unmarshal JWK set: invalid character '<' looking for beginning of value", fakeJWKSURL))
	})
}

func TestAuthenticator_Handler(t *testing.T) {
	// GIVEN
	scopes := "scope-a scope-b"

	privateJWKS, err := authenticator.FetchJWK(context.TODO(), PrivateJWKSURL)
	require.NoError(t, err)

	privateJWKS2, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS2URL)
	require.NoError(t, err)

	privateJWKS3, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS3URL)
	require.NoError(t, err)

	t.Run("http client configured without redirects", func(t *testing.T) {
		// WHEN
		jwks, err := authenticator.FetchJWK(context.TODO(), "http://idonotexist.gone", jwk.WithHTTPClient(httpClientWithoutRedirectsWithMockTransport))

		// THEN
		require.Nil(t, jwks)
		require.Contains(t, err.Error(), "failed to fetch remote JWK (status = 302)") // the existing error is due to FetchJWK function logic and checking if response status code is different thann 200, in this case it's 302. But redirect was not performed
	})

	t.Run("Success - token with signing method", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)
		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - token with no signing method when it's allowed", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, true, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		require.NoError(t, err)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - with client user provided", func(t *testing.T) {
		clientUser := "foo"
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerWithClientUser(t, defaultTenant, clientUser, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)
		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
		req.Header.Add(ClientIDHeaderKey, clientUser)

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when tenant is empty", func(t *testing.T) {
		// GIVEN
		tnt := ""
		middleware := createMiddleware(t, true, claimsValidatorMock())
		handler := testHandler(t, tnt, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, tnt, scopes)
		require.NoError(t, err)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - retry parsing token with synchronizing JWKS", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint(PublicJWKS2URL)

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the first key", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKS3URL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS3.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the second key", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKS3URL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS3.Get(1)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - retry parsing token with failing synchronizing JWKS", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint("invalid.url.scheme")

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Internal Server Error: while synchronizing JWKS during parsing token: while fetching JWKS from endpoint invalid.url.scheme: invalid url scheme ", apperrors.InternalError)
		assertGraphqlResponse(t, expected, response)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - token with no signing method when it's not allowed", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=unexpected signing method: none]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - can't parse token", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add(AuthorizationHeaderKey, "Bearer fake-token")

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=token contains an invalid number of segments]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - invalid header and bearer token", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add("invalidHeader", "Bearer fake-token")

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=invalid bearer token]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - token without signing key", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, nil, false)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=while getting the key ID: Internal Server Error: unable to find the key ID in the token]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - after successful parsing claims are not valid", func(t *testing.T) {
		// GIVEN
		v := &automock.ClaimsValidator{}
		v.On("Validate", mock.Anything, mock.Anything).Return(apperrors.NewTenantNotFoundError("externalTenantName"))
		middleware := createMiddleware(t, false, v)
		handler := testHandler(t, "", scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, "", scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Tenant not found [externalTenant=externalTenantName]", apperrors.TenantNotFound)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - after successful parsing claims there are no scopes", func(t *testing.T) {
		// GIVEN
		requiredScopes := []string{"wanted-scope"}
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claims.NewScopesValidator(requiredScopes))
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)
		middleware := auth.Handler()
		handler := testHandler(t, "", scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, "", scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse(fmt.Sprintf("Unauthorized [reason=Not all required scopes %q were found in claim with scopes %q]", requiredScopes, scopes), apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		// GIVEN
		middleware := createMiddleware(t, false, claimsValidatorMock())
		handler := testHandler(t, defaultTenant, scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		oldKey, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		newKey, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		oldKeyID := oldKey.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, newKey, &oldKeyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=crypto/rsa: verification error]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})
}

func TestAuthenticator_NSAdapterHandler(t *testing.T) {
	// GIVEN
	scopes := "scope-a scope-b"

	privateJWKS, err := authenticator.FetchJWK(context.TODO(), PrivateJWKSURL)
	require.NoError(t, err)

	privateJWKS2, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS2URL)
	require.NoError(t, err)

	privateJWKS3, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS3URL)
	require.NoError(t, err)

	t.Run("Success - token with signing method", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)
		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - token with no signing method when it's allowed", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, true, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		require.NoError(t, err)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - retry parsing token with synchronizing JWKS", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint(PublicJWKS2URL)

		middleware := auth.NSAdapterHandler()

		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the first key", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKS3URL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		middleware := auth.Handler()

		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS3.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the second key", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKS3URL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		middleware := auth.Handler()

		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS3.Get(1)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - retry parsing token with failing synchronizing JWKS", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claimsValidatorMock())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint("invalid.url.scheme")

		middleware := auth.NSAdapterHandler()

		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - token with no signing method when it's not allowed", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - can't parse token", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add(AuthorizationHeaderKey, "Bearer fake-token")

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - invalid header and bearer token", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add("invalidHeader", "Bearer fake-token")

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - token without signing key", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, nil, false)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - after successful parsing claims are not valid", func(t *testing.T) {
		// GIVEN
		v := &automock.ClaimsValidator{}
		v.On("Validate", mock.Anything, mock.Anything).Return(apperrors.NewTenantNotFoundError("externalTenantName"))
		middleware := createNSAdapterMiddleware(t, false, v)
		handler := testHandlerNSAdapterMiddleware(t)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, "", scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - after successful parsing claims there is no tenant", func(t *testing.T) {
		// GIVEN
		auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, false, ClientIDHeaderKey, claims.NewClaimsValidator())
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)
		middleware := auth.NSAdapterHandler()
		handler := testHandlerNSAdapterMiddleware(t)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, "", "", key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		// GIVEN
		middleware := createNSAdapterMiddleware(t, false, claimsValidatorMock())
		handler := testHandlerNSAdapterMiddleware(t)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		oldKey, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		newKey, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		oldKeyID := oldKey.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, newKey, &oldKeyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		// WHEN
		middleware(handler).ServeHTTP(rr, req)

		// THEN
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response httputil.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixErrorResponse("missing or invalid authorization token", http.StatusUnauthorized)
		assert.Equal(t, expected, response)
	})
}

func createNotSingedToken(t *testing.T, tenant string, scopes string) string {
	tokenClaims := struct {
		Tenant       string                `json:"tenant"`
		Scopes       string                `json:"scopes"`
		ConsumerID   string                `json:"consumerID"`
		ConsumerType consumer.ConsumerType `json:"consumerType"`
		OnBehalfOf   string                `json:"onBehalfOf"`
		jwt.StandardClaims
	}{
		Scopes:       scopes,
		ConsumerID:   "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType: consumer.Runtime,
	}

	tenantJSON, err := json.Marshal(map[string]string{"consumerTenant": tenant, "externalTenant": ""})
	require.NoError(t, err)
	tokenClaims.Tenant = string(tenantJSON)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, tokenClaims)

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return signedToken
}

func createTokenWithSigningMethod(t *testing.T, tenant string, scopes string, key jwk.Key, keyID *string, isSigningKeyAvailable bool) string {
	tokenClaims := struct {
		Tenant       string                `json:"tenant"`
		Scopes       string                `json:"scopes"`
		ConsumerID   string                `json:"consumerID"`
		ConsumerType consumer.ConsumerType `json:"consumerType"`
		OnBehalfOf   string                `json:"onBehalfOf"`
		jwt.StandardClaims
	}{
		Scopes:       scopes,
		ConsumerID:   "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType: consumer.Runtime,
	}

	tenantJSON, err := json.Marshal(map[string]string{"consumerTenant": tenant, "externalTenant": "externalTenantName"})
	require.NoError(t, err)
	tokenClaims.Tenant = string(tenantJSON)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)

	if isSigningKeyAvailable {
		token.Header[authenticator.JwksKeyIDKey] = keyID
	}

	var rawKey interface{}
	err = key.Raw(&rawKey)
	require.NoError(t, err)

	signedToken, err := token.SignedString(rawKey)
	require.NoError(t, err)

	return signedToken
}

func createMiddleware(t *testing.T, allowJWTSigningNone bool, claimsValidatorMock authenticator.ClaimsValidator) func(next http.Handler) http.Handler {
	auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, allowJWTSigningNone, ClientIDHeaderKey, claimsValidatorMock)
	err := auth.SynchronizeJWKS(context.TODO())
	require.NoError(t, err)
	return auth.Handler()
}

func createNSAdapterMiddleware(t *testing.T, allowJWTSigningNone bool, claimsValidatorMock authenticator.ClaimsValidator) func(next http.Handler) http.Handler {
	auth := authenticator.New(httpClientWithoutRedirects, PublicJWKSURL, allowJWTSigningNone, ClientIDHeaderKey, claimsValidatorMock)
	err := auth.SynchronizeJWKS(context.TODO())
	require.NoError(t, err)
	return auth.NSAdapterHandler()
}

func testHandlerNSAdapterMiddleware(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}

func testHandler(t *testing.T, expectedTenant string, scopes string) http.HandlerFunc {
	return testHandlerWithClientUser(t, expectedTenant, "", scopes)
}

func testHandlerWithClientUser(t *testing.T, expectedTenant, expectedClientUser, scopes string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantFromContext, err := tenant.LoadFromContext(r.Context())
		if !apperrors.IsTenantRequired(err) {
			require.NoError(t, err)
		}
		clientUserFromContext, err := client.LoadFromContext(r.Context())
		if expectedClientUser == "" {
			require.Error(t, err)
		}
		scopesFromContext, err := scope.LoadFromContext(r.Context())
		require.NoError(t, err)

		require.Equal(t, expectedTenant, tenantFromContext)
		require.Equal(t, expectedClientUser, clientUserFromContext)
		scopesArray := strings.Split(scopes, " ")
		require.ElementsMatch(t, scopesArray, scopesFromContext)

		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}

func fixEmptyRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	return req
}

func fixGraphqlResponse(msg string, errorType apperrors.ErrorType) graphql.Response {
	return graphql.Response{
		Data: []byte("null"),
		Errors: []*gqlerror.Error{{
			Message:    msg,
			Extensions: map[string]interface{}{"error_code": errorType, "error": errorType.String()}}}}
}

func fixErrorResponse(msg string, statusCode int) httputil.ErrorResponse {
	return httputil.ErrorResponse{
		Error: httputil.Error{
			Code:    statusCode,
			Message: msg,
		},
	}
}

func assertGraphqlResponse(t *testing.T, expected, actual graphql.Response) {
	require.Len(t, expected.Errors, 1)
	require.Len(t, actual.Errors, 1)
	assert.Equal(t, expected.Errors[0].Extensions["error"], actual.Errors[0].Extensions["error"])

	errType, ok := expected.Errors[0].Extensions["error_code"].(apperrors.ErrorType)
	require.True(t, ok)
	actualErrCode := int(errType)

	errCode, ok := actual.Errors[0].Extensions["error_code"].(float64)
	require.True(t, ok)
	expectedErrCode := int(errCode)
	assert.Equal(t, expectedErrCode, actualErrCode)
	assert.Equal(t, expected.Errors[0].Message, actual.Errors[0].Message)
}

func claimsValidatorMock() authenticator.ClaimsValidator {
	v := &automock.ClaimsValidator{}
	v.On("Validate", mock.Anything, mock.Anything).Return(nil)
	return v
}
