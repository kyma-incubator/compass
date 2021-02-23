package authenticator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/form3tech-oss/jwt-go"
	directorAuth "github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/authenticator"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	PublicJWKSURL              = "file://../../../director/internal/authenticator/testdata/jwks-public.json"
	PrivateJWKSURL             = "file://../../../director/internal/authenticator/testdata/jwks-private.json"
	PrivateJWKS2URL            = "file://../../../director/internal/authenticator/testdata/jwks-private2.json"
	PublicJWKS2URL             = "file://../../../director/internal/authenticator/testdata/jwks-public2.json"
	fakeJWKSURL                = "file://../../../director/internal/authenticator/testdata/invalid.json"
	ZoneId                     = "private-zone"
	HandlerEndpoint            = "tenants/v1/callback/test-tenant"
	SubscriptionCallbacksScope = "Callback"
)

var (
	trustedPrefixes   = []string{"scopeA."}
	untrustedPrefixes = []string{"fakeScope"}
	fakeScopes        = []string{fmt.Sprintf("%s%s", untrustedPrefixes[0], SubscriptionCallbacksScope)}
	scopes            = []string{fmt.Sprintf("%s%s", trustedPrefixes[0], SubscriptionCallbacksScope)}
)

type Tenant struct {
	TenantId string
}

func TestMiddleware_SynchronizeJWKS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, ZoneId, SubscriptionCallbacksScope, trustedPrefixes, true)

		//when
		err := auth.SynchronizeJWKS(context.TODO())

		//then
		require.NoError(t, err)
	})

	t.Run("Error when can't fetch JWKS", func(t *testing.T) {
		//given
		authFake := authenticator.New(fakeJWKSURL, ZoneId, SubscriptionCallbacksScope, trustedPrefixes, true)

		//when
		err := authFake.SynchronizeJWKS(context.TODO())

		//then
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while fetching JWKS from endpoint %s: failed to unmarshal JWK: invalid character '<' looking for beginning of value", fakeJWKSURL))
	})
}

func TestMiddleware_Handler(t *testing.T) {
	//given
	privateJWKS, err := directorAuth.FetchJWK(context.TODO(), PrivateJWKSURL)
	require.NoError(t, err)

	t.Run("Success - token with signing method", func(t *testing.T) {
		//given
		middleware := createMiddleware(t)
		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		token := createTokenWithSigningMethod(t, scopes, ZoneId, privateJWKS.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success - retry parsing token with synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, ZoneId, SubscriptionCallbacksScope, trustedPrefixes, true)
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint(PublicJWKS2URL)

		middleware := auth.Handler()

		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		privateJWKS2, err := directorAuth.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		token := createTokenWithSigningMethod(t, scopes, ZoneId, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, "OK", rr.Body.String())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Error - retry parsing token with failing synchronizing JWKS", func(t *testing.T) {
		//given
		auth := authenticator.New(PublicJWKSURL, ZoneId, SubscriptionCallbacksScope, trustedPrefixes, true)
		err := auth.SynchronizeJWKS(context.TODO())
		require.NoError(t, err)

		auth.SetJWKSEndpoint("invalid.url.scheme")

		middleware := auth.Handler()

		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		privateJWKS2, err := directorAuth.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		token := createTokenWithSigningMethod(t, scopes, ZoneId, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		message, ok := response["message"]
		require.True(t, ok)

		expected := "Internal Server Error: while synchronizing JWKs during parsing token: while fetching JWKS from endpoint invalid.url.scheme: invalid url scheme "
		assert.Equal(t, expected, message)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - can't parse token", func(t *testing.T) {
		//given
		middleware := createMiddleware(t)
		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		req.Header.Add("Authorization", "Bearer fake-token")

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		message, ok := response["message"]
		require.True(t, ok)

		expected := "Unauthorized [reason=token contains an invalid number of segments]"
		assert.Equal(t, expected, message)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - token signed with different key", func(t *testing.T) {
		//given
		middleware := createMiddleware(t)
		handler := testHandler(t)

		privateJWKS2, err := directorAuth.FetchJWK(context.TODO(), PrivateJWKS2URL)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		token := createTokenWithSigningMethod(t, scopes, ZoneId, privateJWKS2.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		message, ok := response["message"]
		require.True(t, ok)

		expected := "Unauthorized [reason=crypto/rsa: verification error]"
		assert.Equal(t, expected, message)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run(" Error - no bearer token sent", func(t *testing.T) {
		//given
		middleware := createMiddleware(t)
		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - untrusted zone id provided in token", func(t *testing.T) {
		//given
		untrustedZoneId := "fakeZone"
		middleware := createMiddleware(t)
		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		token := createTokenWithSigningMethod(t, scopes, untrustedZoneId, privateJWKS.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		message, ok := response["message"]
		require.True(t, ok)

		expected := fmt.Sprintf(`Zone id "%s" from user token is not trusted`, untrustedZoneId)
		assert.Equal(t, expected, message)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Error - invalid scopes provided in token", func(t *testing.T) {
		//given
		middleware := createMiddleware(t)
		handler := testHandler(t)
		rr := httptest.NewRecorder()
		req := emptyRequest(t)

		token := createTokenWithSigningMethod(t, fakeScopes, ZoneId, privateJWKS.Keys[0])
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		//when
		middleware(handler).ServeHTTP(rr, req)

		//then
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		message, ok := response["message"]
		require.True(t, ok)

		expected := fmt.Sprintf(`Scope "%s" is not trusted`, fakeScopes)
		assert.Equal(t, expected, message)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func createTokenWithSigningMethod(t *testing.T, scopes []string, zone string, key jwk.Key) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, authenticator.Claims{
		Scopes: scopes,
		ZID:    zone,
	})

	materializedKey, err := key.Materialize()
	require.NoError(t, err)

	signedToken, err := token.SignedString(materializedKey)
	require.NoError(t, err)

	return signedToken
}

func createMiddleware(t *testing.T) func(next http.Handler) http.Handler {
	auth := authenticator.New(PublicJWKSURL, ZoneId, SubscriptionCallbacksScope, trustedPrefixes, true)
	err := auth.SynchronizeJWKS(context.TODO())
	require.NoError(t, err)

	return auth.Handler()
}

func testHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	}
}

func emptyRequest(t *testing.T) *http.Request {
	providedTenant := &Tenant{
		TenantId: "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72",
	}
	byteTenant, err := json.Marshal(providedTenant)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", HandlerEndpoint, bytes.NewBuffer(byteTenant))
	require.NoError(t, err)

	return req
}
