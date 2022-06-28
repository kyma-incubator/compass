package retry

import (
	"github.com/avast/retry-go"
	"net/http"
	"time"
)

// ExecutableHTTPFunc defines a generic HTTP function to be executed
type ExecutableHTTPFunc func() (*http.Response, error)

type Config struct {
	Attempts uint          `envconfig:"APP_HTTP_RETRY_ATTEMPTS,default=3"`
	Delay    time.Duration `envconfig:"APP_HTTP_RETRY_DELAY,default=100ms"`
}

// HTTPExecutor is capable of executing HTTP requests with a leveraged retry mechanism for more resilience
type HTTPExecutor struct {
	attempts uint
	delay    time.Duration
}

func NewHTTPExecutor(config *Config) *HTTPExecutor {
	return &HTTPExecutor{
		attempts: config.Attempts,
		delay:    config.Delay,
	}
}

// Execute wraps the provided ExecutableHTTPFunc with a retry mechanism and executes it
func (rhe *HTTPExecutor) Execute(doRequest ExecutableHTTPFunc) (*http.Response, error) {
	var resp *http.Response
	var err error
	err = retry.Do(func() error {
		resp, err = doRequest()
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(rhe.attempts), retry.Delay(rhe.delay))

	return resp, err
}
