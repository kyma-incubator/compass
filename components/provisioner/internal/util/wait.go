package util

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

func WaitForFunction(interval time.Duration, timeout time.Duration, isReady func() (bool, error)) error {
	done := time.After(timeout)

	for {
		b, err := isReady()
		if err != nil {
			return err
		}

		if b {
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

func Retry(interval time.Duration, count int, operation func() error) error {
	var err error
	for i := 0; i < count; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		logrus.Errorf("Error during updating operation status: %s", err.Error())
		time.Sleep(interval)
	}

	return err
}
