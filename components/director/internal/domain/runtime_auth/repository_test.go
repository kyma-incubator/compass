package runtime_auth_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Get(t *testing.T) {
	// GIVEN
	apiID := "foo"
	rtmID := "bar"
	rtmAuthID := "baz"

	modelRtmAuth := fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&rtmAuthID, rtmID, apiID, true)

	testErr := errors.New("test error")

	stmt := `SELECT id, tenant_id, runtime_id, api_def_id, value FROM public.runtime_auths WHERE tenant_id = $1 AND runtime_id = $2 AND api_def_id = $3`

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(*modelRtmAuth, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: rtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.Get(ctx, testTenant, apiID, rtmID)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, modelRtmAuth, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)

	})

	t.Run("Error from DB", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(nil)

		// WHEN
		result, err := repo.Get(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(model.RuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: rtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.Get(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetOrDefault(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	rtmID := "foo"
	apiID := "bar"
	rtmAuthID := "baz"

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ra.id, $2 AS api_def_id, COALESCE(ra.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value FROM (SELECT * FROM runtimes WHERE id = $3) AS r LEFT OUTER JOIN (SELECT * FROM runtime_auths WHERE api_def_id = $2 AND runtime_id = $3 AND tenant_id = $1) AS ra ON ra.runtime_id = r.id`

	modelRtmAuth := fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&rtmAuthID, rtmID, apiID, true)

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(*modelRtmAuth, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: rtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.GetOrDefault(ctx, testTenant, apiID, rtmID)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, modelRtmAuth, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when extracting persistance from context", func(t *testing.T) {
		repo := runtime_auth.NewRepository(nil)

		// WHEN
		result, err := repo.GetOrDefault(context.TODO(), testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to fetch database from context")
		assert.Nil(t, result)
	})

	t.Run("Error when querying db", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).
			WithArgs(testTenant, apiID, rtmID).WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(nil)

		// WHEN
		result, err := repo.GetOrDefault(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(model.RuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: rtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.GetOrDefault(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListForAllRuntimes(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	apiID := "bar"

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ra.id, $2 AS api_def_id, coalesce(ra.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value FROM (SELECT * FROM runtime_auths WHERE api_def_id = $2 AND tenant_id = $1) AS ra RIGHT OUTER JOIN runtimes AS r ON ra.runtime_id = r.id WHERE r.tenant_id = $1`

	modelRtmAuths := []model.RuntimeAuth{
		*fixModelRuntimeAuth(strings.Ptr("ra1"), "r1", apiID, fixModelAuth()),
		*fixModelRuntimeAuth(strings.Ptr("ra2"), "r2", apiID, fixModelAuth()),
		*fixModelRuntimeAuth(strings.Ptr("ra3"), "r3", apiID, fixModelAuth()),
	}
	ents := []runtime_auth.Entity{
		fixEntity(strings.Ptr("ra1"), "r1", apiID, true),
		fixEntity(strings.Ptr("ra2"), "r2", apiID, true),
		fixEntity(strings.Ptr("ra3"), "r3", apiID, true),
	}

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ents[0]).Return(modelRtmAuths[0], nil).Once()
		conv.On("FromEntity", ents[1]).Return(modelRtmAuths[1], nil).Once()
		conv.On("FromEntity", ents[2]).Return(modelRtmAuths[2], nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{
			{id: "ra1", rtmID: "r1", apiID: apiID},
			{id: "ra2", rtmID: "r2", apiID: apiID},
			{id: "ra3", rtmID: "r3", apiID: apiID},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.ListForAllRuntimes(ctx, testTenant, apiID)

		// THEN
		require.NoError(t, err)
		assert.ElementsMatch(t, modelRtmAuths, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when extracting persistance from context", func(t *testing.T) {
		repo := runtime_auth.NewRepository(nil)

		// WHEN
		result, err := repo.ListForAllRuntimes(context.TODO(), testTenant, apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to fetch database from context")
		assert.Nil(t, result)
	})

	t.Run("Error when querying db", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID).WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(nil)

		// WHEN
		result, err := repo.ListForAllRuntimes(ctx, testTenant, apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ents[0]).Return(model.RuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: "ra1", rtmID: "r1", apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		result, err := repo.ListForAllRuntimes(ctx, testTenant, apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})
}

func TestPgRepository_Upsert(t *testing.T) {
	// GIVEN
	rtmID := "foo"
	apiID := "bar"
	rtmAuthID := "baz"

	stmt := `INSERT INTO public.runtime_auths ( id, tenant_id, runtime_id, api_def_id, value ) VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, runtime_id, api_def_id ) DO UPDATE SET value=EXCLUDED.value`

	modelRtmAuth := fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&rtmAuthID, rtmID, apiID, true)

	testErr := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelRtmAuth).Return(ent, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(modelRtmAuth.ID, modelRtmAuth.TenantID, modelRtmAuth.RuntimeID, modelRtmAuth.APIDefID, testMarshalledSchema).
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(ctx, *modelRtmAuth)

		// THEN
		require.NoError(t, err)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error from DB", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelRtmAuth).Return(ent, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(modelRtmAuth.ID, modelRtmAuth.TenantID, modelRtmAuth.RuntimeID, modelRtmAuth.APIDefID, testMarshalledSchema).
			WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(ctx, *modelRtmAuth)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		dbMock.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error when converting runtime auth", func(t *testing.T) {
		testErr := errors.New("test error")
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelRtmAuth).Return(runtime_auth.Entity{}, testErr).Once()

		repo := runtime_auth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(context.TODO(), *modelRtmAuth)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		conv.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	// GIVEN
	apiID := "foo"
	rtmID := "bar"

	testErr := errors.New("test error")

	stmt := `DELETE FROM public.runtime_auths WHERE tenant_id = $1 AND api_def_id = $2 AND runtime_id = $3`

	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(nil)

		// WHEN
		err := repo.Delete(ctx, testTenant, apiID, rtmID)

		// THEN
		require.NoError(t, err)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := runtime_auth.NewRepository(nil)

		// WHEN
		err := repo.Delete(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		dbMock.AssertExpectations(t)
	})
}
