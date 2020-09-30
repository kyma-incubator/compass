package log_test

import (
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *log.Config
		ExpectValid    bool
	}{
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *log.Config {
				return log.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Should return error on invalid log level",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Level = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Log Format should be invalid",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = ""
				return config
			},
		},
		{
			Msg: "Should return error on unsupported log format",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Output should be invalid",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Output = ""
				return config
			},
		},
		{
			Msg: "Should return error on unsupported output",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Output = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Bootstrap Correlation ID should be invalid",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.BootstrapCorrelationID = ""
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

