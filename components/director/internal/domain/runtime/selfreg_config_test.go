package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelfRegConfig_MapInstanceConfigs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		selfRegConfig := SelfRegConfig{
			InstanceClientIDPath:     "clientId",
			InstanceClientSecretPath: "clientSecret",
			InstanceURLPath:          "url",
			InstanceTokenURLPath:     "tokenUrl",
			InstanceCertPath:         "clientCert",
			InstanceKeyPath:          "clientKey",
			InstanceConfigs: `{"eu-1":{"clientId":"client_id","clientSecret":"client_secret","url":"url","tokenUrl":"token-url","clientCert":"cert","clientKey":"key"},
							  "eu-2":{"clientId":"client_id_2","clientSecret":"client_secret_2","url":"url-2","tokenUrl":"token-url-2","clientCert":"cert2","clientKey":"key2"}}`,
		}
		expectedRegionToInstanceConfig := map[string]InstanceConfig{
			"eu-1": {
				ClientID:     "client_id",
				ClientSecret: "client_secret",
				URL:          "url",
				TokenURL:     "token-url",
				Cert:         "cert",
				Key:          "key",
			},
			"eu-2": {
				ClientID:     "client_id_2",
				ClientSecret: "client_secret_2",
				URL:          "url-2",
				TokenURL:     "token-url-2",
				Cert:         "cert2",
				Key:          "key2",
			},
		}

		selfRegConfig.MapInstanceConfigs()

		require.Equal(t, expectedRegionToInstanceConfig["eu-1"], selfRegConfig.RegionToInstanceConfig["eu-1"])
		require.Equal(t, expectedRegionToInstanceConfig["eu-2"], selfRegConfig.RegionToInstanceConfig["eu-2"])
	})
}
