package retrier

import (
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
)

func DoOnCondition(componentName string, risky func() error, retryIf func(err error) bool) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", componentName).Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(retryIf))
}

func DoOnTemporaryConnectionProblems(componentName string, risky func() error) error {
	return DoOnCondition(componentName, risky, func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	})
}

func Do(componentName string, risky func() error) error {
	return DoOnCondition(componentName, risky, nil)
}
