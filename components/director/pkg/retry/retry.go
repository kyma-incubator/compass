package retry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

// ExecutableHTTPFunc defines a generic HTTP function to be executed
type ExecutableHTTPFunc func() (*http.Response, error)

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
	}, retry.Attempts(he.attempts), retry.Delay(he.delay))

	return resp, err
}
