package authenticator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
)

const defaultTenant = "af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e"
const PublicJWKSURL = "file://testdata/jwks-public.json"
const PrivateJWKSURL = "file://testdata/jwks-private.json"
const PrivateJWKS2URL = "file://testdata/jwks-private2.json"
const PublicJWKS2URL = "file://testdata/jwks-public2.json"
const fakeJWKSURL = "file://testdata/invalid.json"

func TestAuthenticator_SynchronizeJWKS(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, true)
		//when
		err := auth.SynchronizeJWKS(context.TODO())

		//then
		require.NoError(t, err)
	})

	t.Run("Error when can't fetch JWKS", func(t *testing.T) {
		//given
		authFake := authenticator.New(fakeJWKSURL, true)

		//when
		err := authFake.SynchronizeJWKS(context.TODO())

		//then
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while fetching JWKS from endpoint %s: failed to unmarshal JWK: invalid character '<' looking for beginning of value", fakeJWKSURL))
	})
}

func TestAuthenticator_Handler(t *testing.T) {
	//given
	scopes := "scope-a scope-b"

	privateJWKS, err := authenticator.FetchJWK(context.TODO(), PrivateJWKSURL)
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

	t.Run("Success - with client user provided", func(t *testing.T) {
		clientUser := "foo"
		//given
		middleware := createMiddleware(t, false)
		handler := testHandlerWithClientUser(t, defaultTenant, clientUser, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Add(authenticator.ClientUserHeader, clientUser)

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

	t.Run("Success - retry parsing token with synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, false)
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint(PublicJWKS2URL)

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		privateJWKS2, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - retry parsing token with failing synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, false)
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint("invalid.url.scheme")

		middleware := auth.Handler()

		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		privateJWKS2, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Internal Server Error: while synchronizing JWKs during parsing token: while fetching JWKS from endpoint invalid.url.scheme: invalid url scheme ", apperrors.InternalError)
		assertGraphqlResponse(t, expected, response)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
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
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=unexpected signing method: none]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
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
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=token contains an invalid number of segments]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)

		privateJWKS2, err := authenticator.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=crypto/rsa: verification error]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
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
	err := auth.SynchronizeJWKS(context.TODO())
	require.NoError(t, err)
	return auth.Handler()
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
