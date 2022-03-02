package runtimemapping_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/hydrator/internal/runtimemapping"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestJWKsFetch_GetKey(t *testing.T) {
	t.Run("should fetch and return valid key", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := runtimemapping.NewJWKsFetch()
		token := createToken()

		// WHEN
		key, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, key)
	})

	t.Run("should return error when token is nil", func(t *testing.T) {
		// GIVEN
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), nil)

		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("token cannot be nil").Error())
	})

	t.Run("should return error when unable to cast claims to MapClaims", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: while getting the discovery URL: Internal Server Error: unable to cast claims to the MapClaims")
	})

	t.Run("should return error when claims have no issuer claim", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{Claims: &jwt.MapClaims{}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: while getting the discovery URL: while getting the issuer from claims: Internal Server Error: no issuer claim found")
	})

	t.Run("should return error when claims have non-string issuer claim", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": byte(1)}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: while getting the discovery URL: while getting the issuer from claims: Internal Server Error: unable to cast the issuer to a string")
	})

	t.Run("should return error when claims have issuer claim in non-URL format", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": ":///cdef://"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: while getting the discovery URL: while parsing the issuer URL [issuer=:///cdef://]: parse \":///cdef://\": missing protocol scheme")
	})

	t.Run("should return error when discovery URL does not return proper response", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": "http://domain.local"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while getting the JWKs URI: while getting the configuration discovery: Get \"http://domain.local/.well-known/openid-configuration\": dial tcp: lookup domain.local")
	})

	t.Run("should return error when discovery URL does not return proper response", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := "A not valid json string"
			writeResponse(t, w, []byte(payload))
		})

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": "http://domain.local"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: while decoding the configuration discovery response: invalid character 'A' looking for beginning of value")
	})

	t.Run("should return error when discovery response contains invalid JWKs URL", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := map[string]interface{}{
				"jwks_uri": byte(0x88),
			}
			data, err := json.Marshal(payload)
			require.NoError(t, err)

			writeResponse(t, w, data)
		})

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": "http://domain.local"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the JWKs URI: Internal Server Error: unable to cast the JWKs URI to a string")
	})

	t.Run("should return error when unable to fetch the JWKs", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/.well-known/openid-configuration":
				payload := map[string]interface{}{
					"jwks_uri": "http://domain.local/keys",
				}
				data, err := json.Marshal(payload)
				require.NoError(t, err)

				writeResponse(t, w, data)
				return
			case "/keys":
				w.WriteHeader(404)
				return
			}
		})

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": "http://domain.local"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while fetching JWKs: failed to fetch remote JWK (status = 404)")
	})

	t.Run("should return error when unable to fetch the JWKs", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{Claims: &jwt.MapClaims{"iss": "http://domain.local"}}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the key ID: Internal Server Error: unable to find the key ID in the token")
	})

	t.Run("should return error when unable to fetch the JWKs", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{
			Claims: &jwt.MapClaims{"iss": "http://domain.local"},
			Header: map[string]interface{}{
				"kid": byte(0x88),
			},
		}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while getting the key ID: Internal Server Error: unable to cast the key ID to a string")
	})

	t.Run("should return error when unable to find a proper key", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		token := &jwt.Token{
			Claims: &jwt.MapClaims{"iss": "http://domain.local"},
			Header: map[string]interface{}{
				"kid": "555-666-777",
			},
			Method: jwt.GetSigningMethod("RS256"),
		}
		jwksFetch := runtimemapping.NewJWKsFetch()

		// WHEN
		_, err := jwksFetch.GetKey(context.TODO(), token)

		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to find a proper key").Error())
	})
}

func TestTokenVerifier_Verify(t *testing.T) {
	privateKeys := readJWK(t, "testdata/jwks-private.json")
	cachePeriod := 5 * time.Minute

	t.Run("should validate token using cache for keys", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := runtimemapping.NewJWKsFetch()
		jwksCache := runtimemapping.NewJWKsCache(jwksFetch, cachePeriod)
		tokenVerifier := runtimemapping.NewTokenVerifier(jwksCache)
		key, ok := privateKeys.Get(0)
		assert.True(t, ok)

		token := createSignedToken(t, key)

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(context.TODO(), logrus.NewEntry(logger))

		// WHEN
		claims, err := tokenVerifier.Verify(ctx, token)

		// THEN
		require.NoError(t, err)
		require.Equal(t, 1, jwksCache.GetSize())
		require.NotNil(t, claims)
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "Adding key 67bf0153-a6dc-4f06-9ce4-2f203b79adc8 to cache", hook.LastEntry().Message)
	})

	t.Run("should validate token using keys", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := runtimemapping.NewJWKsFetch()
		tokenVerifier := runtimemapping.NewTokenVerifier(jwksFetch)
		key, ok := privateKeys.Get(0)
		assert.True(t, ok)

		token := createSignedToken(t, key)

		// WHEN
		claims, err := tokenVerifier.Verify(context.TODO(), token)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, claims)
	})

	t.Run("should return error when token is empty", func(t *testing.T) {
		jwksFetch := runtimemapping.NewJWKsFetch()
		tokenVerifier := runtimemapping.NewTokenVerifier(jwksFetch)
		token := ""

		// WHEN
		_, err := tokenVerifier.Verify(context.TODO(), token)

		// THEN
		require.EqualError(t, err, apperrors.NewUnauthorizedError("token cannot be empty").Error())
	})

	t.Run("should return error when token is invalid", func(t *testing.T) {
		jwksFetch := runtimemapping.NewJWKsFetch()
		tokenVerifier := runtimemapping.NewTokenVerifier(jwksFetch)
		token := "invalid token"

		// WHEN
		_, err := tokenVerifier.Verify(context.TODO(), token)

		// THEN
		require.EqualError(t, err, "while veryfing the token: while parsing the token with claims: token contains an invalid number of segments")
	})
}

func setHTTPClient(c *http.Client) func() {
	defaultClient := *(http.DefaultClient)
	http.DefaultClient = c
	return func() {
		http.DefaultClient = &defaultClient
	}
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	srv := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, srv.Listener.Addr().String())
			},
		},
	}

	return cli, srv.Close
}

func mockValidJWKsHandler(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			payload := map[string]interface{}{
				"jwks_uri": "http://domain.local/keys",
			}
			data, err := json.Marshal(payload)
			require.NoError(t, err)
			writeResponse(t, w, data)
			return
		case "/keys":
			data, err := ioutil.ReadFile("testdata/jwks-public.json")
			require.NoError(t, err)
			writeResponse(t, w, data)
			return
		}
	}
}

func createToken() *jwt.Token {
	token := &jwt.Token{
		Header: map[string]interface{}{
			"alg": "RS256",
			"kid": "67bf0153-a6dc-4f06-9ce4-2f203b79adc8",
		},
		Method: jwt.GetSigningMethod("RS256"),
		Claims: &jwt.MapClaims{
			"iss": "http://domain.local",
		},
	}

	return token
}

func createSignedToken(t *testing.T, key jwk.Key) string {
	token := createToken()

	var rawKey interface{}
	err := key.Raw(&rawKey)

	require.NoError(t, err)
	signedToken, err := token.SignedString(rawKey)
	require.NoError(t, err)

	return signedToken
}

func readJWK(t *testing.T, path string) jwk.Set {
	jwksPrivateData, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	jwkSet, err := jwk.Parse(jwksPrivateData)
	require.NoError(t, err)

	return jwkSet
}

func writeResponse(t *testing.T, w http.ResponseWriter, d []byte) {
	_, err := w.Write(d)
	require.NoError(t, err)
}
