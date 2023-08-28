package healthz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	persistautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/pkg/errors"

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
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(false)
		transactioner.On("PingContext", ctx).Once().Return(nil)
		defer transactioner.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("XXXXXXXXXXXXXX", false, nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
	})

	t.Run("success when cached result", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(false)
		transactioner.On("PingContext", ctx).Twice().Return(nil)
		defer transactioner.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("XXXXXXXXXXXXXX", false, nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
		AssertHandlerStatusCodeForReadiness(t, ready, 200, "")
	})

	t.Run("fail when ping fails", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)
		transactioner.On("PingContext", ctx).Once().Return(errors.New("Ping failure"))
		defer transactioner.AssertExpectations(t)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("XXXXXXXXXXXXXX", false, nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail when schema compatibility check fails", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("YYYYYYYYYYYYY", false, nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail when schema dirty flag is set", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("XXXXXXXXXXXXXX", true, nil)
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail when error is received while getting schema version from database", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(nil)

		ctxWithTransaction := persistence.SaveToContext(ctx, tx)
		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(tx, nil)
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)

		repository := &automock.Repository{}
		repository.On("GetVersion", ctxWithTransaction).Once().Return("", false, errors.New("db error"))
		defer repository.AssertExpectations(t)

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail while opening transaction", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}

		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(nil, errors.New("error while opening transaction"))
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)

		repository := &automock.Repository{}

		ready := healthz.NewReady(transactioner, cfg, repository)

		// THEN
		AssertHandlerStatusCodeForReadiness(t, ready, 500, "")
	})

	t.Run("fail while committing", func(t *testing.T) {
		ctx := context.Background()
		tx := &persistautomock.PersistenceTx{}
		tx.On("Commit").Once().Return(errors.New("commit error"))

		transactioner := &persistautomock.Transactioner{}
		transactioner.On("Begin").Once().Return(nil, errors.New("error while opening transaction"))
		transactioner.On("RollbackUnlessCommitted", ctx, tx).Once().Return(true)

		repository := &automock.Repository{}

		ready := healthz.NewReady(transactioner, cfg, repository)

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
