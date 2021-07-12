package healthz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/healthz/automock"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/stretchr/testify/require"
)

func TestNewDbIndicatorFunc(t *testing.T) {
	t.Run("should return error when pinger fails", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		pinger := &automock.Pinger{}
		defer pinger.AssertExpectations(t)
		pinger.On("PingContext", ctx).Return(errors.New("db error"))

		// WHEN
		dbIndFunc := healthz.NewDbIndicatorFunc(pinger)
		status := dbIndFunc(ctx)

		// THEN
		require.NotNil(t, status)
		require.Error(t, status.Error())
		require.Equal(t, "db error", status.Error().Error())
		require.Equal(t, "Error pinging database", status.Details())
	})

	t.Run("should return nil when pinger succeeds", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		pinger := &automock.Pinger{}
		defer pinger.AssertExpectations(t)
		pinger.On("PingContext", ctx).Return(nil)

		// WHEN
		dbIndFunc := healthz.NewDbIndicatorFunc(pinger)
		status := dbIndFunc(ctx)

		// THEN
		require.NotNil(t, status)
		require.NoError(t, status.Error())
		require.Equal(t, "", status.Details())
	})
}
