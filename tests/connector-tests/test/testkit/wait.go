package testkit

import (
	"errors"
	"time"
)

func WaitForFunction(interval time.Duration, timeout time.Duration, isReady func() bool) error {
	done := time.After(timeout)

	for {
		if isReady() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}
