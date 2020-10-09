package http_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *http.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *http.Config {
				return http.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Negative Timeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.Timeout = -1
				return config
			},
		},
		{
			Msg: "Negative TLSHandshakeTimeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.TLSHandshakeTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative IdleConnTimeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.IdleConnTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative ResponseHeaderTimeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.ResponseHeaderTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative DialTimeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.DialTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative ExpectContinueTimeout should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.ExpectContinueTimeout = -1
				return config
			},
		},
		{
			Msg: "Negative MaxIdleConns should be invalid",
			ConfigProvider: func() *http.Config {
				config := http.DefaultConfig()
				config.MaxIdleConns = -1
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
