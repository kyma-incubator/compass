package retry

import (
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var defaultConfig = Config{
	Attempts: 3,
	Delay:    100 * time.Millisecond,
}

var mockedErr = errors.New("error")
var mockedResp = &http.Response{StatusCode: http.StatusOK}

func TestExecuteSuccessfulWhenNoErrors(t *testing.T) {
	t.Run("should finish successfully when there are no errors when executing the http request", func(t *testing.T) {
		retryExecutor := NewHTTPExecutor(&defaultConfig)

		invocations := 0
		resp, err := retryExecutor.Execute(func() (*http.Response, error) {
			defer func() {
				invocations++
			}()

			return mockedResp, nil
		})

		require.Equal(t, mockedResp, resp)
		require.NoError(t, err)
		require.Equal(t, 1, invocations)
	})
}

func TestExecuteFailsInitiallyButSucceedsOnLastTryWithNoErrors(t *testing.T) {
	t.Run("should finish successfully on the last retry attempt", func(t *testing.T) {
		retryExecutor := NewHTTPExecutor(&defaultConfig)

		invocations := 0
		resp, err := retryExecutor.Execute(func() (*http.Response, error) {
			defer func() {
				invocations++
			}()

			if invocations != (int(defaultConfig.Attempts) - 1) {
				return nil, mockedErr
			}

			return mockedResp, nil
		})

		require.Equal(t, mockedResp, resp)
		require.NoError(t, err)
		require.Equal(t, int(defaultConfig.Attempts), invocations)
	})
}

func TestExecuteFailsWhenAllRetryAttemptsFinishWithFailureDueToError(t *testing.T) {
	t.Run("should finish with failure when all retry attempts fail", func(t *testing.T) {
		retryExecutor := NewHTTPExecutor(&defaultConfig)

		invocations := 0
		resp, err := retryExecutor.Execute(func() (*http.Response, error) {
			defer func() {
				invocations++
			}()
			return nil, mockedErr
		})

		require.Nil(t, resp)
		require.Error(t, err)
		require.Equal(t, int(defaultConfig.Attempts), invocations)
	})
}

func TestExecuteFailsWhenAllRetryAttemptsFinishWithFailureDueToUnexpectedStatusCode(t *testing.T) {
	t.Run("should finish with failure when all retry attempts fail", func(t *testing.T) {
		retryExecutor := NewHTTPExecutor(&defaultConfig)

		invocations := 0
		mockedErrResp := http.Response{StatusCode: http.StatusInternalServerError}
		resp, err := retryExecutor.Execute(func() (*http.Response, error) {
			defer func() {
				invocations++
			}()
			return &mockedErrResp, nil
		})

		require.Equal(t, &mockedErrResp, resp)
		require.Error(t, err)
		require.Equal(t, int(defaultConfig.Attempts), invocations)
	})
}
