package healthz_test

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/kyma-incubator/compass/components/director/internal/healthz/automock"
	"github.com/stretchr/testify/require"
)

var (
	cfg = healthz.ReadyConfig{
		SchemaMigrationVersion: "XXXXXXXXXXXXXX",
	}
)

func TestNewReadinessHandler(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		ctx := context.TODO()
		pinger := &automock.Pinger{}
		pinger.On("PingContext", ctx).Once().Return(nil)
		defer pinger.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctx).Once().Return("XXXXXXXXXXXXXX", nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(ctx, pinger, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
	})

	t.Run("success when cached result", func(t *testing.T) {
		ctx := context.TODO()
		pinger := &automock.Pinger{}
		pinger.On("PingContext", ctx).Twice().Return(nil)
		defer pinger.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctx).Once().Return("XXXXXXXXXXXXXX", nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(ctx, pinger, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
	})

	t.Run("fail when ping fails", func(t *testing.T) {
		ctx := context.TODO()
		pinger := &automock.Pinger{}
		pinger.On("PingContext", ctx).Once().Return(errors.New("Ping failure"))
		defer pinger.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctx).Once().Return("XXXXXXXXXXXXXX", nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(ctx, pinger, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail when schema compatibility check fails", func(t *testing.T) {
		ctx := context.TODO()
		pinger := &automock.Pinger{}

		repository := &automock.Repository{}
		repository.On("GetVersion", ctx).Once().Return("YYYYYYYYYYYYY", nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(ctx, pinger, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail when error is received while getting schema version from database", func(t *testing.T) {
		ctx := context.TODO()
		pinger := &automock.Pinger{}

		repository := &automock.Repository{}
		repository.On("GetVersion", ctx).Once().Return("", errors.New("db error"))
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(ctx, pinger, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})
}

func AssertHandlerStatusCodeForReadiness(t *testing.T, r *healthz.Ready, expectedCode int, expectedBody string) {
	req, err := http.NewRequest("GET", "/readyz", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthz.NewReadinessHandler(r))
	// WHEN
	handler.ServeHTTP(rr, req)
	// THEN
	require.Equal(t, expectedCode, rr.Code)
	require.Equal(t, expectedBody, rr.Body.String())
}
