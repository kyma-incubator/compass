package runtimemapping

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/form3tech-oss/jwt-go"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

var cachePeriod time.Duration = time.Duration(5 * time.Minute)

func TestJWKCacheEntry_Expired(t *testing.T) {
	t.Run("should return true when entry is cached for more than cachePeriod", func(t *testing.T) {
		// WHEN
		entry := jwkCacheEntry{
			key:      "dummy",
			expireAt: time.Now().Add(-1 * cachePeriod),
		}

		// THEN
		require.True(t, entry.IsExpired())
	})

	t.Run("should return false when entry is cached for no longer than cachePeriod", func(t *testing.T) {
		// WHEN
		entry := jwkCacheEntry{
			key:      "dummy",
			expireAt: time.Now().Add(cachePeriod),
		}

		// THEN
		require.False(t, entry.IsExpired())
	})
}

func TestJWKsCache_GetKey(t *testing.T) {
	t.Run("should fetch, cache and return valid key", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)
		token := createToken()

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(context.Background(), logrus.NewEntry(logger))

		// WHEN
		key, err := jwksCache.GetKey(ctx, token)

		// THEN
		require.NoError(t, err)
		require.Equal(t, 1, len(jwksCache.cache))
		require.NotNil(t, key)
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "adding key 67bf0153-a6dc-4f06-9ce4-2f203b79adc8 to cache", hook.LastEntry().Message)
	})

	t.Run("should fetch, cache and return valid key, second call should return from cache", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(mockValidJWKsHandler(t))

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)
		token := createToken()

		logger, hook := logrustest.NewNullLogger()
		ctx := log.ContextWithLogger(context.Background(), logrus.NewEntry(logger))

		// WHEN
		_, err := jwksCache.GetKey(ctx, token)
		require.NoError(t, err)

		key, err := jwksCache.GetKey(ctx, token)

		// THEN
		require.NoError(t, err)
		require.Equal(t, 1, len(jwksCache.cache))
		require.NotNil(t, key)
		require.Equal(t, 2, len(hook.Entries))
		require.Equal(t, "adding key 67bf0153-a6dc-4f06-9ce4-2f203b79adc8 to cache", hook.Entries[0].Message)
		require.Equal(t, "using key 67bf0153-a6dc-4f06-9ce4-2f203b79adc8 from cache", hook.Entries[1].Message)
	})

	t.Run("should return error when token is nil", func(t *testing.T) {
		// GIVEN
		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)

		// WHEN
		_, err := jwksCache.GetKey(context.Background(), nil)

		// THEN
		require.EqualError(t, err, apperrors.NewUnauthorizedError("token cannot be nil").Error())
	})

	t.Run("should return error when unable to get token key ID", func(t *testing.T) {
		// GIVEN
		token := &jwt.Token{}
		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)

		// WHEN
		_, err := jwksCache.GetKey(context.Background(), token)

		// THEN
		require.EqualError(t, err, "while getting the key ID: Internal Server Error: unable to find the key ID in the token")
	})

	t.Run("should return error when unable to get key from remote server", func(t *testing.T) {
		// GIVEN
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		})

		httpClient, teardown := testingHTTPClient(handler)
		defer teardown()

		restoreHTTPClient := setHTTPClient(httpClient)
		defer restoreHTTPClient()

		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)
		token := createToken()

		// WHEN
		_, err := jwksCache.GetKey(context.Background(), token)

		// THEN
		require.EqualError(t, err, "while getting the key with ID [kid=67bf0153-a6dc-4f06-9ce4-2f203b79adc8]: while getting the JWKs URI: while decoding the configuration discovery response: EOF")
	})
}

func TestJWKsCache_Cleanup(t *testing.T) {
	t.Run("should cleanup expired cached keys", func(t *testing.T) {
		// GIVEN
		logger, hook := logrustest.NewNullLogger()
		jwksFetch := NewJWKsFetch()
		jwksCache := NewJWKsCache(jwksFetch, cachePeriod)

		// WHEN
		jwksCache.cache["123"] = jwkCacheEntry{
			expireAt: time.Now().Add(-1 * cachePeriod), //expired
			key:      "abc-key-value",
		}
		jwksCache.cache["234"] = jwkCacheEntry{
			expireAt: time.Now().Add(cachePeriod),
			key:      "def-key-value",
		}
		require.Equal(t, 2, len(jwksCache.cache))

		jwksCache.Cleanup(log.ContextWithLogger(context.Background(), logrus.NewEntry(logger)))

		// THEN
		require.Equal(t, 1, len(jwksCache.cache))
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, "removing key 123 from cache", hook.LastEntry().Message)
	})
}
