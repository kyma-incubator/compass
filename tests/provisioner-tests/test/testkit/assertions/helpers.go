package assertions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Standard require.NoError print only the top wrapper of error
func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if assert.NoError(t, err) {
		return
	}
	fullError := fmt.Sprintf("Received unexpected error: %s", err.Error())
	t.Fatal(fullError, msgAndArgs)
}

// Standard assert.NoError print only the top wrapper of error
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if assert.NoError(t, err) {
		return
	}
	fullError := fmt.Sprintf("Received unexpected error: %s", err.Error())
	t.Error(fullError, msgAndArgs)
}

func AssertNotNilAndEqualString(t *testing.T, expected string, actual *string) {
	if !assert.NotNil(t, actual) {
		assert.Equal(t, expected, *actual)
	}
}

func AssertNotNilAndEqualInt(t *testing.T, expected int, actual *int) {
	if !assert.NotNil(t, actual) {
		assert.Equal(t, expected, *actual)
	}
}

func AssertNotNilAndNotEmptyString(t *testing.T, str *string) {
	if !assert.NotNil(t, str) {
		assert.NotEmpty(t, *str)
	}
}
