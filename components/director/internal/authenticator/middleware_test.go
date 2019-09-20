package authenticator_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
)

const tnt = "2a1502ba-aded-11e9-a2a3-2a2ae2dbcce4"
const PublicJWKSURL = "file://testdata/jwks-public.json"
const PrivateJWKSURL = "file://testdata/jwks-private.json"
const PrivateJWKS2URL = "file://testdata/jwks-private2.json"
const fakeJWKSURL = "https://example.com"

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
	scopes := []string{"scope-a", "scope-b"}

	privateJWKS, err := authenticator.FetchJWK(PrivateJWKSURL)
	require.NoError(t, err)

	t.Run("Success - token with signing method", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, scopes)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)

		token, err := createTokenWithSigningMethod(t, tnt, scopes, privateJWKS.Keys[0])
		require.NoError(t, err)
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
		handler := testHandler(t, scopes)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)

		token, err := createToken(tnt, scopes)
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
		handler := testHandler(t, scopes)

		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)

		token, err := createToken(tnt, scopes)
		require.NoError(t, err)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)
		//then
		assert.Equal(t, "while parsing token: Unexpected signing method: none\n", rr.Body.String())
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run("Error - tenant header is not provided", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, scopes)

		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, rr.Code, 400)
		assert.Equal(t, "No tenant header\n", rr.Body.String())
	})

	t.Run("Error - can't get token from header", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, scopes)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "Invalid bearer token\n", rr.Body.String())
		assert.Equal(t, 400, rr.Code)
	})

	t.Run("Error - can't parse token", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, scopes)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)
		req.Header.Add("Authorization", "Bearer fake-token")

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "while parsing token: token contains an invalid number of segments\n", rr.Body.String())
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, scopes)

		privateJWKS2, err := authenticator.FetchJWK(PrivateJWKS2URL)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		req.Header.Add("tenant", tnt)

		token, err := createTokenWithSigningMethod(t, tnt, scopes, privateJWKS2.Keys[0])
		require.NoError(t, err)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "while parsing token: crypto/rsa: verification error\n", rr.Body.String())
	})

}

type jwtTokenClaims struct {
	Scopes []string `json:"scopes"`
	Tenant string   `json:"tenant"`
	jwt.StandardClaims
}

func createToken(t string, scopes []string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwtTokenClaims{
		Tenant: t,
		Scopes: scopes,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", errors.Wrap(err, "while signing token")
	}

	return signedToken, nil
}

func createTokenWithSigningMethod(t *testing.T, tnt string, scopes []string, key jwk.Key) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtTokenClaims{
		Tenant: tnt,
		Scopes: scopes,
	})

	materializedKey, err := key.Materialize()
	require.NoError(t, err)
	signedToken, err := token.SignedString(materializedKey)
	if err != nil {
		return "", errors.Wrap(err, "while signing token")
	}

	return signedToken, nil
}

func createMiddleware(t *testing.T, allowJWTSigningNone bool) func(next http.Handler) http.Handler {
	auth := authenticator.New(PublicJWKSURL, allowJWTSigningNone)
	publicJWKS, err := authenticator.FetchJWK(PublicJWKSURL)
	require.NoError(t, err)
	auth.Jwks = publicJWKS
	return auth.Handler()
}

func testHandler(t *testing.T, scopes []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantFromContext, err := tenant.LoadFromContext(r.Context())
		require.NoError(t, err)
		scopesFromContext, err := scope.LoadFromContext(r.Context())
		require.NoError(t, err)

		require.Equal(t, tnt, tenantFromContext)
		require.Equal(t, scopes, scopesFromContext)

		_, err = w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}
