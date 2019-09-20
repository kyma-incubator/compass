package scope_test

import (
	"go.uber.org/goleak"
	"testing"
)

// TestMain is used to verify that there are no unexpected goroutines running at the end of a test
// TODO
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
