package signal_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleInterrupts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)
	term <- os.Interrupt
	require.Eventually(t, func() bool {
		return assert.Equal(t, ctx.Err(), context.Canceled)
	}, time.Second, 50*time.Millisecond)
}
