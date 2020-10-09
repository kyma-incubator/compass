package oauth_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *oauth.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *oauth.Config {
				return oauth.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Missing SecretName should be invalid",
			ConfigProvider: func() *oauth.Config {
				config := oauth.DefaultConfig()
				config.SecretName = ""
				return config
			},
		},
		{
			Msg: "Missing SecretNamespace should be invalid",
			ConfigProvider: func() *oauth.Config {
				config := oauth.DefaultConfig()
				config.SecretNamespace = ""
				return config
			},
		},
		{
			Msg: "Missing TokenValue should be invalid when running locally",
			ConfigProvider: func() *oauth.Config {
				config := oauth.DefaultConfig()
				config.Local = true
				config.TokenValue = ""
				return config
			},
		},
		{
			Msg: "Negative WaitSecretTimeout should be invalid",
			ConfigProvider: func() *oauth.Config {
				config := oauth.DefaultConfig()
				config.WaitSecretTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative WaitSecretTimeout should be invalid",
			ConfigProvider: func() *oauth.Config {
				config := oauth.DefaultConfig()
				config.WaitSecretTimeout = 0
				return config
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			err := test.ConfigProvider().Validate()
			if test.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
