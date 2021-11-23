package certloader

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertificateCache(t *testing.T) {
	certBytes, keyBytes := generateTestCertAndKey(t, testCN)
	cache := NewCertificateCache(secretName)
	invalidCert := "-----BEGIN CERTIFICATE-----\naZOCUHlJ1wKwnYiLnOofB1xyIUZhVLaJy7Ob\n-----END CERTIFICATE-----\n"
	invalidKey := "-----BEGIN RSA PRIVATE KEY-----\n7qFmWkbkOAM9CUPx5RwSRt45oxlQjvDniZALWqbYxgO5f8cYZsEAyOU1n2DXgiei\n-----END RSA PRIVATE KEY-----\n"

	testCases := []struct {
		Name             string
		CacheData        map[string][]byte
		ExpectedErrorMsg string
	}{
		{
			Name:      "Successfully get certificate from cache",
			CacheData: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
		},
		{
			Name:             "Error when secret data in the cache is empty",
			CacheData:        map[string][]byte{},
			ExpectedErrorMsg: "There is no certificate data in the cache",
		},
		{
			Name:             "Error when certificate is invalid",
			CacheData:        map[string][]byte{secretCertKey: []byte("invalid"), secretKeyKey: []byte("invalid")},
			ExpectedErrorMsg: "Error while decoding certificate pem block",
		},
		{
			Name:      "Error when parsing certificate",
			CacheData: map[string][]byte{secretCertKey: []byte(invalidCert), secretKeyKey: []byte("invalid")},
		},
		{
			Name:             "Error when private key is invalid",
			CacheData:        map[string][]byte{secretCertKey: certBytes, secretKeyKey: []byte("invalid")},
			ExpectedErrorMsg: "Error while decoding private key pem block",
		},
		{
			Name:      "Error when parsing private key",
			CacheData: map[string][]byte{secretCertKey: certBytes, secretKeyKey: []byte(invalidKey)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			cache.Put(testCase.CacheData)
			tlsCert, err := cache.Get()

			if err != nil {
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				require.Nil(t, tlsCert)
			} else {
				require.NotNil(t, tlsCert)
			}
		})
	}

	t.Run("Error when certificate data not found in the cache", func(t *testing.T) {
		nonExistingCacheKey := "non-existing-key"
		cache := NewCertificateCache(nonExistingCacheKey)
		tlsCert, err := cache.Get()
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("Client certificate data not found in the cache for key: %s", nonExistingCacheKey))
		require.Nil(t, tlsCert)
	})
}
