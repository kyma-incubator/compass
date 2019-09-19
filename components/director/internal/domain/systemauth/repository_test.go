package systemauth_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	sysAuthID := "foo"
	objID := "bar"

	modelAuth := fixModelAuth()

	insertQuery := `^INSERT INTO public.system_auths \(.+\) VALUES \(.+\)$`

	t.Run("Success creating auth for Runtime", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuth := fixModelSystemAuth("foo", model.RuntimeReference, objID, modelAuth)
		entSysAuth := fixEntity(sysAuthID, model.RuntimeReference, objID, true)

		dbMock.ExpectExec(insertQuery).
			WithArgs(fixSystemAuthCreateArgs(entSysAuth)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		convMock := automock.Converter{}
		convMock.On("ToEntity", *modelSysAuth).Return(entSysAuth, nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		err := pgRepository.Create(ctx, *modelSysAuth)

		//THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success creating auth for Application", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuth := fixModelSystemAuth("foo", model.ApplicationReference, objID, modelAuth)
		entSysAuth := fixEntity(sysAuthID, model.ApplicationReference, objID, true)

		dbMock.ExpectExec(insertQuery).
			WithArgs(fixSystemAuthCreateArgs(entSysAuth)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		convMock := automock.Converter{}
		convMock.On("ToEntity", *modelSysAuth).Return(entSysAuth, nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		err := pgRepository.Create(ctx, *modelSysAuth)

		//THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success creating auth for Integration System", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuth := fixModelSystemAuth("foo", model.IntegrationSystemReference, objID, modelAuth)
		entSysAuth := fixEntity(sysAuthID, model.IntegrationSystemReference, objID, true)

		dbMock.ExpectExec(insertQuery).
			WithArgs(fixSystemAuthCreateArgs(entSysAuth)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		convMock := automock.Converter{}
		convMock.On("ToEntity", *modelSysAuth).Return(entSysAuth, nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		err := pgRepository.Create(ctx, *modelSysAuth)

		//THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Error converting", func(t *testing.T) {
		ctx := context.TODO()

		modelSysAuth := fixModelSystemAuth("foo", model.IntegrationSystemReference, objID, modelAuth)

		convMock := automock.Converter{}
		convMock.On("ToEntity", *modelSysAuth).Return(systemauth.Entity{}, testErr).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		err := pgRepository.Create(ctx, *modelSysAuth)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
	})

	t.Run("Error creating", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuth := fixModelSystemAuth("foo", model.RuntimeReference, objID, modelAuth)
		entSysAuth := fixEntity(sysAuthID, model.RuntimeReference, objID, true)

		dbMock.ExpectExec(insertQuery).
			WithArgs(fixSystemAuthCreateArgs(entSysAuth)...).
			WillReturnError(testErr)

		convMock := automock.Converter{}
		convMock.On("ToEntity", *modelSysAuth).Return(entSysAuth, nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		// WHEN
		err := pgRepository.Create(ctx, *modelSysAuth)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())

		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListForObject(t *testing.T) {
	//GIVEN
	objID := "bar"

	modelAuth := fixModelAuth()

	t.Run("Success listing auths for Runtime", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuths := []*model.SystemAuth{
			fixModelSystemAuth("foo", model.RuntimeReference, objID, modelAuth),
			fixModelSystemAuth("bar", model.RuntimeReference, objID, modelAuth),
		}
		entSysAuths := []systemauth.Entity{
			fixEntity("foo", model.RuntimeReference, objID, true),
			fixEntity("bar", model.RuntimeReference, objID, true),
		}

		query := `SELECT id, tenant_id, app_id, runtime_id, integration_system_id, value FROM public.system_auths WHERE tenant_id=$1 AND runtime_id = 'bar'`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				{
					id:       modelSysAuths[0].ID,
					tenant:   testTenant,
					appID:    modelSysAuths[0].AppID,
					rtmID:    modelSysAuths[0].RuntimeID,
					intSysID: modelSysAuths[0].IntegrationSystemID,
				},
				{
					id:       modelSysAuths[1].ID,
					tenant:   testTenant,
					appID:    modelSysAuths[1].AppID,
					rtmID:    modelSysAuths[1].RuntimeID,
					intSysID: modelSysAuths[1].IntegrationSystemID,
				},
			}))

		convMock := automock.Converter{}
		convMock.On("FromEntity", entSysAuths[0]).Return(*modelSysAuths[0], nil).Once()
		convMock.On("FromEntity", entSysAuths[1]).Return(*modelSysAuths[1], nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListForObject(ctx, testTenant, model.RuntimeReference, objID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success listing auths for Application", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuths := []*model.SystemAuth{
			fixModelSystemAuth("foo", model.ApplicationReference, objID, modelAuth),
			fixModelSystemAuth("bar", model.ApplicationReference, objID, modelAuth),
		}
		entSysAuths := []systemauth.Entity{
			fixEntity("foo", model.ApplicationReference, objID, true),
			fixEntity("bar", model.ApplicationReference, objID, true),
		}

		query := `SELECT id, tenant_id, app_id, runtime_id, integration_system_id, value FROM public.system_auths WHERE tenant_id=$1 AND app_id = 'bar'`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				{
					id:       modelSysAuths[0].ID,
					tenant:   testTenant,
					appID:    modelSysAuths[0].AppID,
					rtmID:    modelSysAuths[0].RuntimeID,
					intSysID: modelSysAuths[0].IntegrationSystemID,
				},
				{
					id:       modelSysAuths[1].ID,
					tenant:   testTenant,
					appID:    modelSysAuths[1].AppID,
					rtmID:    modelSysAuths[1].RuntimeID,
					intSysID: modelSysAuths[1].IntegrationSystemID,
				},
			}))

		convMock := automock.Converter{}
		convMock.On("FromEntity", entSysAuths[0]).Return(*modelSysAuths[0], nil).Once()
		convMock.On("FromEntity", entSysAuths[1]).Return(*modelSysAuths[1], nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListForObject(ctx, testTenant, model.ApplicationReference, objID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success listing auths for Integration System", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuths := []*model.SystemAuth{
			fixModelSystemAuth("foo", model.IntegrationSystemReference, objID, modelAuth),
			fixModelSystemAuth("bar", model.IntegrationSystemReference, objID, modelAuth),
		}
		entSysAuths := []systemauth.Entity{
			fixEntity("foo", model.IntegrationSystemReference, objID, true),
			fixEntity("bar", model.IntegrationSystemReference, objID, true),
		}

		query := `SELECT id, tenant_id, app_id, runtime_id, integration_system_id, value FROM public.system_auths WHERE tenant_id=$1 AND integration_system_id = 'bar'`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(model.IntegrationSystemTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				{
					id:       modelSysAuths[0].ID,
					tenant:   model.IntegrationSystemTenant,
					appID:    modelSysAuths[0].AppID,
					rtmID:    modelSysAuths[0].RuntimeID,
					intSysID: modelSysAuths[0].IntegrationSystemID,
				},
				{
					id:       modelSysAuths[1].ID,
					tenant:   model.IntegrationSystemTenant,
					appID:    modelSysAuths[1].AppID,
					rtmID:    modelSysAuths[1].RuntimeID,
					intSysID: modelSysAuths[1].IntegrationSystemID,
				},
			}))

		convMock := automock.Converter{}
		convMock.On("FromEntity", entSysAuths[0]).Return(*modelSysAuths[0], nil).Once()
		convMock.On("FromEntity", entSysAuths[1]).Return(*modelSysAuths[1], nil).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListForObject(ctx, model.IntegrationSystemTenant, model.IntegrationSystemReference, objID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Error listing auths for unsupported reference object type", func(t *testing.T) {
		pgRepository := systemauth.NewRepository(nil)
		errorMsg := "unsupported reference object type"

		//WHEN
		result, err := pgRepository.ListForObject(context.TODO(), testTenant, "unsupported", objID)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), errorMsg)
		require.Nil(t, result)
	})

	t.Run("Error listing auths", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := `SELECT id, tenant_id, app_id, runtime_id, integration_system_id, value FROM public.system_auths WHERE tenant_id=$1 AND integration_system_id = 'bar'`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(model.IntegrationSystemTenant).
			WillReturnError(testErr)

		pgRepository := systemauth.NewRepository(nil)

		//WHEN
		result, err := pgRepository.ListForObject(ctx, model.IntegrationSystemTenant, model.IntegrationSystemReference, objID)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error converting auth", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		modelSysAuths := []*model.SystemAuth{
			fixModelSystemAuth("foo", model.IntegrationSystemReference, objID, modelAuth),
			fixModelSystemAuth("bar", model.IntegrationSystemReference, objID, modelAuth),
		}
		entSysAuths := []systemauth.Entity{
			fixEntity("foo", model.IntegrationSystemReference, objID, true),
			fixEntity("bar", model.IntegrationSystemReference, objID, true),
		}

		query := `SELECT id, tenant_id, app_id, runtime_id, integration_system_id, value FROM public.system_auths WHERE tenant_id=$1 AND integration_system_id = 'bar'`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(model.IntegrationSystemTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				{
					id:       modelSysAuths[0].ID,
					tenant:   model.IntegrationSystemTenant,
					appID:    modelSysAuths[0].AppID,
					rtmID:    modelSysAuths[0].RuntimeID,
					intSysID: modelSysAuths[0].IntegrationSystemID,
				},
				{
					id:       modelSysAuths[1].ID,
					tenant:   model.IntegrationSystemTenant,
					appID:    modelSysAuths[1].AppID,
					rtmID:    modelSysAuths[1].RuntimeID,
					intSysID: modelSysAuths[1].IntegrationSystemID,
				},
			}))

		convMock := automock.Converter{}
		convMock.On("FromEntity", entSysAuths[0]).Return(model.SystemAuth{}, testErr).Once()
		pgRepository := systemauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListForObject(ctx, model.IntegrationSystemTenant, model.IntegrationSystemReference, objID)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	// GIVEN
	sysAuthID := "foo"

	t.Run("Success deleting auth for Runtime", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := `DELETE FROM public.system_auths WHERE tenant_id = $1 AND id = $2`
		dbMock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(testTenant, sysAuthID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		repo := systemauth.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, testTenant, sysAuthID, model.RuntimeReference)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success deleting auth for Application", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := `DELETE FROM public.system_auths WHERE tenant_id = $1 AND id = $2`
		dbMock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(testTenant, sysAuthID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		repo := systemauth.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, testTenant, sysAuthID, model.ApplicationReference)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success deleting auth for Integration System", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := `DELETE FROM public.system_auths WHERE tenant_id = $1 AND id = $2`
		dbMock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(model.IntegrationSystemTenant, sysAuthID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		repo := systemauth.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, model.IntegrationSystemTenant, sysAuthID, model.IntegrationSystemReference)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when deleting", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := `DELETE FROM public.system_auths WHERE tenant_id = $1 AND id = $2`
		dbMock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(testTenant, sysAuthID).
			WillReturnError(testErr)

		repo := systemauth.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, testTenant, sysAuthID, model.RuntimeReference)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
}
