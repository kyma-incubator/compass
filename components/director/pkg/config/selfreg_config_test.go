package config

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSelfRegConfig_MapInstanceConfigs(t *testing.T) {
	testCases := []struct {
		Name                     string
		Config                   SelfRegConfig
		OauthMode                oauth.AuthMode
		ExpectedRegionToInstance map[string]InstanceConfig
		ExpectedErr              error
	}{
		{
			Name: "Success for MTLS mode",
			Config: SelfRegConfig{
				SelfRegisterSecretPath:   "testdata/TestSelfRegConfig_MapInstanceConfigs_MTLS_Success.golden",
				OAuthMode:                oauth.Mtls,
				InstanceClientIDPath:     "clientId",
				InstanceClientSecretPath: "clientSecret",
				InstanceURLPath:          "url",
				InstanceTokenURLPath:     "tokenUrl",
				InstanceCertPath:         "clientCert",
				InstanceKeyPath:          "clientKey",
			},
			ExpectedRegionToInstance: map[string]InstanceConfig{
				"eu-1": {
					ClientID: "client_id",
					URL:      "url",
					TokenURL: "token-url",
					Cert:     "cert",
					Key:      "key",
				},
				"eu-2": {
					ClientID: "client_id_2",
					URL:      "url-2",
					TokenURL: "token-url-2",
					Cert:     "cert2",
					Key:      "key2",
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Success for Standard mode",
			Config: SelfRegConfig{
				SelfRegisterSecretPath:   "testdata/TestSelfRegConfig_MapInstanceConfigs_StandardMode_Success.golden",
				OAuthMode:                oauth.Standard,
				InstanceClientIDPath:     "clientId",
				InstanceClientSecretPath: "clientSecret",
				InstanceURLPath:          "url",
				InstanceTokenURLPath:     "tokenUrl",
				InstanceCertPath:         "clientCert",
				InstanceKeyPath:          "clientKey",
			},
			ExpectedRegionToInstance: map[string]InstanceConfig{
				"eu-1": {
					ClientID:     "client_id",
					ClientSecret: "client_secret",
					URL:          "url",
					TokenURL:     "token-url",
				},
				"eu-2": {
					ClientID:     "client_id_2",
					ClientSecret: "client_secret",
					URL:          "url-2",
					TokenURL:     "token-url-2",
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when Client ID and URLs are missing",
			Config: SelfRegConfig{
				SelfRegisterSecretPath:   "testdata/TestSelfRegConfig_MapInstanceConfigs_StandardMode_Missing_ClientID_URL.golden",
				OAuthMode:                oauth.Standard,
				InstanceClientIDPath:     "clientId",
				InstanceClientSecretPath: "clientSecret",
				InstanceURLPath:          "url",
				InstanceTokenURLPath:     "tokenUrl",
				InstanceCertPath:         "clientCert",
				InstanceKeyPath:          "clientKey",
			},
			ExpectedRegionToInstance: nil,
			ExpectedErr:              errors.Errorf("while validating instance for region: %q: Client ID is missing, Token URL is missing, URL is missing", "eu-2"),
		},
		{
			Name: "Returns error when Client Secret is missing in Standard flow",
			Config: SelfRegConfig{
				SelfRegisterSecretPath:   "testdata/TestSelfRegConfig_MapInstanceConfigs_StandardMode_Missing_ClientSecret.golden",
				OAuthMode:                oauth.Standard,
				InstanceClientIDPath:     "clientId",
				InstanceClientSecretPath: "clientSecret",
				InstanceURLPath:          "url",
				InstanceTokenURLPath:     "tokenUrl",
				InstanceCertPath:         "clientCert",
				InstanceKeyPath:          "clientKey",
			},
			ExpectedRegionToInstance: nil,
			ExpectedErr:              errors.Errorf("while validating instance for region: %q: Client Secret is missing", "eu-2"),
		},
		{
			Name: "Returns error when Certificate and Key is missing in MTLS flow",
			Config: SelfRegConfig{
				SelfRegisterSecretPath:   "testdata/TestSelfRegConfig_MapInstanceConfigs_MTLS_Missing_Cert_Key.golden",
				OAuthMode:                oauth.Mtls,
				InstanceClientIDPath:     "clientId",
				InstanceClientSecretPath: "clientSecret",
				InstanceURLPath:          "url",
				InstanceTokenURLPath:     "tokenUrl",
				InstanceCertPath:         "clientCert",
				InstanceKeyPath:          "clientKey",
			},
			ExpectedRegionToInstance: nil,
			ExpectedErr:              errors.Errorf("while validating instance for region: %q: Certificate is missing, Key is missing", "eu-2"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			err := testCase.Config.MapInstanceConfigs()

			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				require.True(t, len(testCase.Config.RegionToInstanceConfig) == 0)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedRegionToInstance, testCase.Config.RegionToInstanceConfig)
			}
		})
	}
}
