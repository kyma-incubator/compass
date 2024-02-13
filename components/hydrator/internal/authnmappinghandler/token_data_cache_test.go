package authnmappinghandler_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler"
	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler/automock"
	"github.com/stretchr/testify/require"
)

func TestTokenDataCache(t *testing.T) {
	tokenDataMock := &automock.TokenData{}
	token := "testToken"
	issuerURL := "testIssuerURL"

	t.Run("success when looking for missing data", func(t *testing.T) {
		testedTokenDataCache := authnmappinghandler.NewTokenDataCache(time.Minute)
		exist, _ := testedTokenDataCache.GetTokenData(token, issuerURL)
		require.False(t, exist)
	})

	t.Run("success when adding data", func(t *testing.T) {
		testedTokenDataCache := authnmappinghandler.NewTokenDataCache(time.Minute)
		testedTokenDataCache.PutTokenData(token, issuerURL, tokenDataMock)
		exist, _ := testedTokenDataCache.GetTokenData(token, issuerURL)
		require.True(t, exist)
	})

	t.Run("success when clear old data from cache", func(t *testing.T) {
		testedTokenDataCache := authnmappinghandler.NewTokenDataCache(time.Millisecond)
		testedTokenDataCache.PutTokenData(token, issuerURL, tokenDataMock)
		time.Sleep(time.Millisecond * 2)
		testedTokenDataCache.Cleanup()
		exist, _ := testedTokenDataCache.GetTokenData(token, issuerURL)
		require.False(t, exist)
	})

	t.Run("success when old data in cache is not expired", func(t *testing.T) {
		testedTokenDataCache := authnmappinghandler.NewTokenDataCache(time.Minute)
		testedTokenDataCache.PutTokenData(token, issuerURL, tokenDataMock)
		time.Sleep(time.Millisecond * 2)
		testedTokenDataCache.Cleanup()
		exist, _ := testedTokenDataCache.GetTokenData(token, issuerURL)
		require.True(t, exist)
	})

}
