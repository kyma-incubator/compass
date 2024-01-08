package tests

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHealthAPI(t *testing.T) {
	// GIVEN
	client := &http.Client{
		Timeout: time.Second * 2,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, conf.HealthUrl, nil)
	require.NoError(t, err)

	// WHEN
	err = retry.Do(
		func() error {
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return errors.New(fmt.Sprintf("Health api returned non 200 response: %d", resp.StatusCode))
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.OnRetry(func(n uint, err error) {
			logrus.WithField("component", "TestHealthAPI").Warnf("OnRetry: attempts: %d, error: %v", n, err)
		}),
		retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "connection refused") ||
				strings.Contains(err.Error(), "connection reset by peer")
		}))

	//THEN
	require.NoError(t, err)
}
