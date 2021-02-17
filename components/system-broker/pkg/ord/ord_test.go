package ord_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/ord"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *ord.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *ord.Config {
				return ord.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Missing ServiceURL should be invalid",
			ConfigProvider: func() *ord.Config {
				config := ord.DefaultConfig()
				config.ServiceURL = ""
				return config
			},
		},
		{
			Msg: "Missing StaticPath should be invalid",
			ConfigProvider: func() *ord.Config {
				config := ord.DefaultConfig()
				config.StaticPath = ""
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
