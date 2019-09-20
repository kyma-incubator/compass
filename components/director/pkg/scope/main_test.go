package scope_test

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain is used to verify that there are no unexpected goroutines running at the end of a test
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
