package server_test

import (
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *server.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *server.Config {
				return server.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Missing Port should be invalid",
			ConfigProvider: func() *server.Config {
				config := server.DefaultConfig()
				config.Port = 0
				return config
			},
		},
		{
			Msg: "Missing RequestTimeout should be invalid",
			ConfigProvider: func() *server.Config {
				config := server.DefaultConfig()
				config.RequestTimeout = 0
				return config
			},
		},
		{
			Msg: "Missing ShutdownTimeout should be invalid",
			ConfigProvider: func() *server.Config {
				config := server.DefaultConfig()
				config.ShutdownTimeout = 0
				return config
			},
		},
		{
			Msg: "Missing SelfURL should be invalid",
			ConfigProvider: func() *server.Config {
				config := server.DefaultConfig()
				config.SelfURL = ""
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
