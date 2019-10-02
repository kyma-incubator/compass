package apiruntimeauth_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Get(t *testing.T) {
	// GIVEN
	apiID := "foo"
	rtmID := "bar"
	apiRtmAuthID := "baz"

	modelAPIRtmAuth := fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&apiRtmAuthID, rtmID, apiID, true)

	testErr := errors.New("test error")

	stmt := `SELECT id, tenant_id, runtime_id, api_def_id, value FROM public.api_runtime_auths WHERE tenant_id = $1 AND runtime_id = $2 AND api_def_id = $3`

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(*modelAPIRtmAuth, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: apiRtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		result, err := repo.Get(ctx, testTenant, apiID, rtmID)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, modelAPIRtmAuth, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)

	})

	t.Run("Error from DB", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(nil)

		// WHEN
		result, err := repo.Get(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting api runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(model.APIRuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: apiRtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, rtmID, apiID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

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
	apiRtmAuthID := "baz"

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ara.id, $2 AS api_def_id, COALESCE(ara.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value FROM (SELECT * FROM runtimes WHERE id = $3) AS r LEFT OUTER JOIN (SELECT * FROM api_runtime_auths WHERE api_def_id = $2 AND runtime_id = $3 AND tenant_id = $1) AS ara ON ara.runtime_id = r.id`

	modelAPIRtmAuth := fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&apiRtmAuthID, rtmID, apiID, true)

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(*modelAPIRtmAuth, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: apiRtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		result, err := repo.GetOrDefault(ctx, testTenant, apiID, rtmID)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, modelAPIRtmAuth, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when extracting persistance from context", func(t *testing.T) {
		repo := apiruntimeauth.NewRepository(nil)

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

		repo := apiruntimeauth.NewRepository(nil)

		// WHEN
		result, err := repo.GetOrDefault(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting api runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ent).Return(model.APIRuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: apiRtmAuthID, rtmID: rtmID, apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

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

	stmt := `SELECT r.id AS runtime_id, r.tenant_id, ara.id, $2 AS api_def_id, coalesce(ara.value, (SELECT default_auth FROM api_definitions WHERE api_definitions.id = $2)) AS value FROM (SELECT * FROM api_runtime_auths WHERE api_def_id = $2 AND tenant_id = $1) AS ara RIGHT OUTER JOIN runtimes AS r ON ara.runtime_id = r.id WHERE r.tenant_id = $1`

	modelAPIRtmAuths := []model.APIRuntimeAuth{
		*fixModelAPIRuntimeAuth(str.Ptr("ara1"), "r1", apiID, fixModelAuth()),
		*fixModelAPIRuntimeAuth(str.Ptr("ara2"), "r2", apiID, fixModelAuth()),
		*fixModelAPIRuntimeAuth(str.Ptr("ara3"), "r3", apiID, fixModelAuth()),
	}
	ents := []apiruntimeauth.Entity{
		fixEntity(str.Ptr("ara1"), "r1", apiID, true),
		fixEntity(str.Ptr("ara2"), "r2", apiID, true),
		fixEntity(str.Ptr("ara3"), "r3", apiID, true),
	}

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ents[0]).Return(modelAPIRtmAuths[0], nil).Once()
		conv.On("FromEntity", ents[1]).Return(modelAPIRtmAuths[1], nil).Once()
		conv.On("FromEntity", ents[2]).Return(modelAPIRtmAuths[2], nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{
			{id: "ara1", rtmID: "r1", apiID: apiID},
			{id: "ara2", rtmID: "r2", apiID: apiID},
			{id: "ara3", rtmID: "r3", apiID: apiID},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		result, err := repo.ListForAllRuntimes(ctx, testTenant, apiID)

		// THEN
		require.NoError(t, err)
		assert.ElementsMatch(t, modelAPIRtmAuths, result)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when extracting persistance from context", func(t *testing.T) {
		repo := apiruntimeauth.NewRepository(nil)

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

		repo := apiruntimeauth.NewRepository(nil)

		// WHEN
		result, err := repo.ListForAllRuntimes(ctx, testTenant, apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		assert.Nil(t, result)

		dbMock.AssertExpectations(t)
	})

	t.Run("Error when converting api runtime auth", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("FromEntity", ents[0]).Return(model.APIRuntimeAuth{}, testErr).Once()

		db, dbMock := testdb.MockDatabase(t)
		rows := fixSQLRows([]sqlRow{{id: "ara1", rtmID: "r1", apiID: apiID}})
		dbMock.ExpectQuery(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

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
	apiRtmAuthID := "baz"

	stmt := `INSERT INTO public.api_runtime_auths ( id, tenant_id, runtime_id, api_def_id, value ) VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, runtime_id, api_def_id ) DO UPDATE SET value=EXCLUDED.value`

	modelAPIRtmAuth := fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())
	ent := fixEntity(&apiRtmAuthID, rtmID, apiID, true)

	testErr := errors.New("test error")

	t.Run("Success", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelAPIRtmAuth).Return(ent, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(modelAPIRtmAuth.ID, modelAPIRtmAuth.TenantID, modelAPIRtmAuth.RuntimeID, modelAPIRtmAuth.APIDefID, testMarshalledSchema).
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(ctx, *modelAPIRtmAuth)

		// THEN
		require.NoError(t, err)

		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error from DB", func(t *testing.T) {
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelAPIRtmAuth).Return(ent, nil).Once()

		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(modelAPIRtmAuth.ID, modelAPIRtmAuth.TenantID, modelAPIRtmAuth.RuntimeID, modelAPIRtmAuth.APIDefID, testMarshalledSchema).
			WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(ctx, *modelAPIRtmAuth)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		dbMock.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error when converting api runtime auth", func(t *testing.T) {
		testErr := errors.New("test error")
		conv := &automock.Converter{}
		conv.On("ToEntity", *modelAPIRtmAuth).Return(apiruntimeauth.Entity{}, testErr).Once()

		repo := apiruntimeauth.NewRepository(conv)

		// WHEN
		err := repo.Upsert(context.TODO(), *modelAPIRtmAuth)

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

	stmt := `DELETE FROM public.api_runtime_auths WHERE tenant_id = $1 AND api_def_id = $2 AND runtime_id = $3`

	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectExec(regexp.QuoteMeta(stmt)).WithArgs(testTenant, apiID, rtmID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)

		repo := apiruntimeauth.NewRepository(nil)

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

		repo := apiruntimeauth.NewRepository(nil)

		// WHEN
		err := repo.Delete(ctx, testTenant, apiID, rtmID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		dbMock.AssertExpectations(t)
	})
}
