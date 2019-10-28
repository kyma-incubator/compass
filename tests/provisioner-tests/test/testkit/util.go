package testkit

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func WaitForFunction(interval, timeout time.Duration, isDone func() bool) error {
	done := time.After(timeout)

	for {
		if isDone() {
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

func RunParallelToMainFunction(timeout time.Duration, mainFunction func() error, parallelFunctions ...func() error) error {
	mainOut := make(chan error, 1)
	go func() {
		mainOut <- mainFunction()
	}()

	errOut := make(chan error, len(parallelFunctions))
	for _, fun := range parallelFunctions {
		go func(function func() error) {
			errOut <- function()
		}(fun)
	}

	funcErrors := make([]error, 0, len(parallelFunctions))

	for {
		select {
		case err := <-errOut:
			funcErrors = append(funcErrors, err)
		case err := <-mainOut:
			if err != nil {
				return errors.Errorf("Main function failed: %s", err.Error())
			}

			if len(funcErrors) < len(parallelFunctions) {
				return errors.Errorf("Not all parallel functions finished. Functions finished %d. Errors: %v", len(funcErrors), processErrors(funcErrors))
			}

			return processErrors(funcErrors)
		case <-time.After(timeout):
			return errors.Errorf("Timeout waiting for for parallel processes to finish. Functions finished %d. Errors: %v", len(funcErrors), processErrors(funcErrors))
		}
	}
}

func processErrors(errorsArray []error) error {
	errorMsg := ""

	for i, err := range errorsArray {
		if err != nil {
			errorMsg = fmt.Sprintf("%s -- Error %d not nil: %s.", errorMsg, i, err.Error())
		}
	}

	if errorMsg != "" {
		return errors.Errorf("Errors: %s", errorMsg)
	}

	return nil
}
