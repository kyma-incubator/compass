package retry

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/pkg/errors"
)

// ExecutableHTTPFunc defines a generic HTTP function to be executed
type ExecutableHTTPFunc func() (*http.Response, error)

// Config configuration for the HTTPExecutor
type Config struct {
	Attempts uint          `envconfig:"APP_HTTP_RETRY_ATTEMPTS"`
	Delay    time.Duration `envconfig:"APP_HTTP_RETRY_DELAY"`
}

// HTTPExecutor is capable of executing HTTP requests with a leveraged retry mechanism for more resilience
type HTTPExecutor struct {
	attempts              uint
	delay                 time.Duration
	acceptableStatusCodes []int
}

// NewHTTPExecutor constructs an HTTPExecutor based on the provided Config
func NewHTTPExecutor(config *Config) *HTTPExecutor {
	return &HTTPExecutor{
		attempts:              config.Attempts,
		delay:                 config.Delay,
		acceptableStatusCodes: []int{http.StatusOK},
	}
}

// WithAcceptableStatusCodes allows overriding the default acceptableStatusCodes values
func (he *HTTPExecutor) WithAcceptableStatusCodes(statusCodes []int) {
	he.acceptableStatusCodes = statusCodes
}

// Execute wraps the provided ExecutableHTTPFunc with a retry mechanism and executes it
func (he *HTTPExecutor) Execute(doRequest ExecutableHTTPFunc) (*http.Response, error) {
	var resp *http.Response
	var err error
	err = retry.Do(func() error {
		resp, err = doRequest()
		if err != nil {
			return err
		}

		for _, code := range he.acceptableStatusCodes {
			if resp.StatusCode == code {
				return nil
			}
		}

		return errors.New(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	},
		retry.Attempts(he.attempts),
		retry.Delay(he.delay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "connection refused") ||
				strings.Contains(err.Error(), "connection reset by peer")
		}))

	return resp, err
}
