package authenticator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
	auths "github.com/kyma-incubator/compass/components/director/pkg/authenticator"
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

type testClaims struct {
	Tenant         string                `json:"tenant"`
	ExternalTenant string                `json:"externalTenant"`
	Scopes         string                `json:"scopes"`
	ConsumerID     string                `json:"consumerID"`
	ConsumerType   consumer.ConsumerType `json:"consumerType"`
	Flow           oathkeeper.AuthFlow   `json:"flow"`
	ZID            string                `json:"zid"`
	jwt.StandardClaims
}

func TestAuthenticator_SynchronizeJWKS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//given
		auth := authenticator.New(true, ClientIDHeaderKey, nil, PublicJWKSURL)
		//when
		err := auth.SynchronizeJWKS(context.TODO())

		//then
		require.NoError(t, err)
	})

	t.Run("Error when can't fetch JWKS", func(t *testing.T) {
		//given
		authFake := authenticator.New(true, ClientIDHeaderKey, nil, fakeJWKSURL)

		//when
		err := authFake.SynchronizeJWKS(context.TODO())

		//then
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while fetching JWKS from endpoint %s: failed to unmarshal JWK set: invalid character '<' looking for beginning of value", fakeJWKSURL))
	})
}

func TestAuthenticator_Handler(t *testing.T) {
	//given
	scopes := "scope-a scope-b"

	privateJWKS, err := auths.FetchJWK(context.TODO(), PrivateJWKSURL)
	require.NoError(t, err)

	privateJWKS2, err := auths.FetchJWK(context.TODO(), PrivateJWKS2URL)
	require.NoError(t, err)

	privateJWKS3, err := auths.FetchJWK(context.TODO(), PrivateJWKS3URL)
	require.NoError(t, err)

	t.Run("Success - token with signing method", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)
		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

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
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

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
		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
		req.Header.Add(ClientIDHeaderKey, clientUser)

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
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - retry parsing token with synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(false, ClientIDHeaderKey, claims.NewOathkeeperClaimsParser(), PublicJWKSURL)
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

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the first key", func(t *testing.T) {
		//given
		auth := authenticator.New(false, ClientIDHeaderKey, claims.NewOathkeeperClaimsParser(), PublicJWKS3URL)
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

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - when we have more than one JWKS and use the second key", func(t *testing.T) {
		//given
		auth := authenticator.New(false, ClientIDHeaderKey, claims.NewOathkeeperClaimsParser(), PublicJWKS3URL)
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

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - retry parsing token with failing synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(false, ClientIDHeaderKey, claims.NewOathkeeperClaimsParser(), PublicJWKSURL)
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

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Internal Server Error: while synchronizing JWKS during parsing token: while fetching JWKS from endpoint invalid.url.scheme: invalid url scheme ", apperrors.InternalError)
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
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

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

		req.Header.Add(AuthorizationHeaderKey, "Bearer fake-token")

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

	t.Run("Error - invalid header and bearer token", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)
		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		req.Header.Add("invalidHeader", "Bearer fake-token")

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=invalid bearer token]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - token without signing key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, defaultTenant, scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS2.Get(0)
		assert.True(t, ok)

		token := createTokenWithSigningMethod(t, defaultTenant, scopes, key, nil, false)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Unauthorized [reason=while getting the key ID: Internal Server Error: unable to find the key ID in the token]", apperrors.Unauthorized)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - after successful parsing claims there is no tenant", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
		handler := testHandler(t, "", scopes)

		rr := httptest.NewRecorder()
		req := fixEmptyRequest(t)

		key, ok := privateJWKS.Get(0)
		assert.True(t, ok)

		keyID := key.KeyID()
		token := createTokenWithSigningMethod(t, "", scopes, key, &keyID, true)
		req.Header.Add(AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response graphql.Response
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		expected := fixGraphqlResponse("Tenant not found [externalTenant=externalTenantName]", apperrors.TenantNotFound)
		assertGraphqlResponse(t, expected, response)
	})

	t.Run("Error - Token signed with different key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t, false)
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
	token := jwt.NewWithClaims(jwt.SigningMethodNone, testClaims{
		Tenant:       tenant,
		Scopes:       scopes,
		ConsumerID:   "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType: consumer.Runtime,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return signedToken
}

func createTokenWithSigningMethod(t *testing.T, tnt string, scopes string, key jwk.Key, keyID *string, isSigningKeyAvailable bool) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, testClaims{
		Tenant:         tnt,
		ExternalTenant: "externalTenantName",
		Scopes:         scopes,
		ConsumerID:     "1e176e48-e258-4091-a584-feb1bf708b7e",
		ConsumerType:   consumer.Runtime,
	})

	if isSigningKeyAvailable {
		token.Header[authenticator.JwksKeyIDKey] = keyID
	}

	var rawKey interface{}
	err := key.Raw(&rawKey)
	require.NoError(t, err)

	signedToken, err := token.SignedString(rawKey)
	require.NoError(t, err)

	return signedToken
}

func createMiddleware(t *testing.T, allowJWTSigningNone bool) func(next http.Handler) http.Handler {
	auth := authenticator.New(allowJWTSigningNone, ClientIDHeaderKey, claims.NewOathkeeperClaimsParser(), PublicJWKSURL)
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
