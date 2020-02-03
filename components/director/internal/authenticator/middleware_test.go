package authenticator_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
)

const defaultTenant = "af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e"
const PublicJWKSURL = "file://testdata/jwks-public.json"
const PrivateJWKSURL = "file://testdata/jwks-private.json"
const PrivateJWKS2URL = "file://testdata/jwks-private2.json"
const fakeJWKSURL = "file://testdata/invalid.json"

func TestAuthenticator_SynchronizeJWKS(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, true)
		//when
		err := auth.SynchronizeJWKS()

		//then
		require.NoError(t, err)
	})

	t.Run("Error when can't fetch JWKS", func(t *testing.T) {
		//given
		authFake := authenticator.New(fakeJWKSURL, true)

		//when
		err := authFake.SynchronizeJWKS()

		//then
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while fetching JWKS from endpoint %s: failed to unmarshal JWK: invalid character '<' looking for beginning of value", fakeJWKSURL))
	})
}

func TestAuthenticator_Handler(t *testing.T) {
	//given
	scopes := "scope-a scope-b"

	privateJWKS, err := authenticator.FetchJWK(PrivateJWKSURL)
	require.NoError(t, err)

	t.Run("Success - token with signing method", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - token with no signing method when it's allowed", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, true)
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		require.NoError(t, err)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when tenant is empty", func(t *testing.T) {
		//given
		tnt := ""
		middleware := createMiddleware(t, true)
		handler := testHandler(t, tnt, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, tnt, scopes)
		require.NoError(t, err)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - token with no signing method when it's not allowed", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createNotSingedToken(t, defaultTenant, scopes)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)
		//then
		assert.Equal(t, "{\"errors\":[{\"message\":\"while parsing token: unexpected signing method: none\"}]}\n", rr.Body.String())
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run("Error - can't parse token", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add("Authorization", "Bearer fake-token")

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "{\"errors\":[{\"message\":\"while parsing token: token contains an invalid number of segments\"}]}\n", rr.Body.String())
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)

		privateJWKS2, err := authenticator.FetchJWK(PrivateJWKS2URL)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "{\"errors\":[{\"message\":\"while parsing token: crypto/rsa: verification error\"}]}\n", rr.Body.String())
	})
}

func createNotSingedToken(t *testing.T, tenant string, scopes string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, authenticator.Claims{
		Tenant:       tenant,
		Scopes:       scopes,
		ConsumerID:   "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType: consumer.Runtime,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return signedToken
}

func createTokenWithSigningMethod(t *testing.T, tnt string, scopes string, key jwk.Key) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, authenticator.Claims{
		Tenant:       tnt,
		Scopes:       scopes,
		ConsumerID:   "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType: consumer.Runtime,
	})

	materializedKey, err := key.Materialize()
	require.NoError(t, err)
	signedToken, err := token.SignedString(materializedKey)
	require.NoError(t, err)

	return signedToken
}

func createMiddleware(t *testing.T, allowJWTSigningNone bool) func(next http.Handler) http.Handler {
	auth := authenticator.New(PublicJWKSURL, allowJWTSigningNone)
	err := auth.SynchronizeJWKS()
	require.NoError(t, err)
	return auth.Handler()
}

func testHandler(t *testing.T, expectedTenant string, scopes string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantFromContext, err := tenant.LoadFromContext(r.Context())
		require.NoError(t, err)
		scopesFromContext, err := scope.LoadFromContext(r.Context())
		require.NoError(t, err)

		require.Equal(t, expectedTenant, tenantFromContext)
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
