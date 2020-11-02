package log_test

import (
	"bytes"
	"context"
	"os"
	"sync"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestMultipleGoroutinesDefaultLog validates that no race conditions occur when two go routines log using the default log
func TestMultipleGoroutinesDefaultLog(t *testing.T) {
	_, err := log.Configure(context.TODO(), log.DefaultConfig())
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		log.D().Debug("message")
	}()
	go func() {
		defer wg.Done()
		log.D().Debug("message")
	}()
	wg.Wait()
}

// TestMultipleGoroutinesContextLog validates that no race conditions occur when two go routines log using the context log
func TestMultipleGoroutinesContextLog(t *testing.T) {
	ctx, err := log.Configure(context.Background(), log.DefaultConfig())
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		log.C(ctx).Debug("message")
	}()
	go func() {
		defer wg.Done()
		log.C(ctx).Debug("message")
	}()
	wg.Wait()
}

// TestMultipleGoroutinesMixedLog validates that no race conditions occur when two go routines log using both context and default log
func TestMultipleGoroutinesMixedLog(t *testing.T) {
	ctx, err := log.Configure(context.TODO(), log.DefaultConfig())
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		log.C(ctx).Debug("message")
	}()
	go func() {
		defer wg.Done()
		log.D().Debug("message")
	}()
	wg.Wait()
}

func TestConfigureReturnsErrorWhenConfigIsInvalid(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *log.Config
	}{
		{
			Msg: "Invalid log level",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Level = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Log Format",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = ""
				return config
			},
		},
		{
			Msg: "Unsupported log format",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Output",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Output = ""
				return config
			},
		},
		{
			Msg: "Unsupported output",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Output = "invalid"
				return config
			},
		},
		{
			Msg: "Missing Bootstrap Correlation ID",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.BootstrapCorrelationID = ""
				return config
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			previousConfig := log.Configuration()
			_, err := log.Configure(context.TODO(), test.ConfigProvider())
			require.Error(t, err)
			currentConfig := log.Configuration()
			require.Equal(t, previousConfig, currentConfig)
		})
	}
}

func TestDefaultLoggerConfiguration(t *testing.T) {
	config := log.DefaultConfig()
	config.Level = "trace"
	ctx, err := log.Configure(context.Background(), config)
	require.NoError(t, err)
	require.Equal(t, log.C(ctx).Level, log.D().Level)
}

func TestConfigureFormat(t *testing.T) {
	var tests = []struct {
		Msg            string
		ConfigProvider func() *log.Config
		ExpectedOutput []string
	}{
		{
			Msg: "text",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = "text"
				return config
			},
			ExpectedOutput: []string{"msg=Test"},
		},
		{
			Msg: "json",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = "json"
				return config
			},
			ExpectedOutput: []string{"\"msg\":\"Test\""},
		},
		{
			Msg: "structured",
			ConfigProvider: func() *log.Config {
				config := log.DefaultConfig()
				config.Format = "structured"
				return config
			},
			ExpectedOutput: []string{
				`"component_type":"application","correlation_id":"system-broker-bootstrap"`,
				`"msg":"Test","type":"log"`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			w := &bytes.Buffer{}
			ctx, err := log.Configure(context.TODO(), test.ConfigProvider())
			require.NoError(t, err)
			entry := log.LoggerFromContext(ctx)
			entry.Logger.SetOutput(w)
			defer entry.Logger.SetOutput(os.Stderr) // return default output
			entry.Info("Test")
			for _, out := range test.ExpectedOutput {
				require.Contains(t, w.String(), out)
			}
		})
	}
}

func TestRegisterFormatterAddsItIfNotExists(t *testing.T) {
	name := "formatter_name"
	err := log.RegisterFormatter(name, &logrus.TextFormatter{})
	require.NoError(t, err)

	config := log.DefaultConfig()
	config.Format = name
	_, err = log.Configure(context.TODO(), config)
	require.NoError(t, err)
}

func TestRegisterFormatterReturnsErrorForAlreadyExistingFormat(t *testing.T) {
	name := "text"
	err := log.RegisterFormatter(name, &logrus.TextFormatter{})
	require.Error(t, err)
}

func TestConfigureWillReconfigureDefaultLoggerEvenIfEntryAlreadyExistsInTheContext(t *testing.T) {
	config := log.DefaultConfig()
	ctx, err := log.Configure(context.TODO(), config)
	require.NoError(t, err)

	config.Level = "trace"
	ctx, err = log.Configure(ctx, config)
	require.NoError(t, err)

	require.Equal(t, log.C(ctx).Level, log.D().Level)
}