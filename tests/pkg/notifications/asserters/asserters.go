package asserters

import (
	"context"
	"testing"
)

type Asserter interface {
	AssertExpectations(t *testing.T, ctx context.Context)
}
