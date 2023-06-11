package apptemplateversion_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(appTemplateVersionEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.app_template_versions ( id, version, title, correlation_ids, release_date, created_at, app_template_id ) VALUES ( ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixAppTemplateVersionCreateArgs(*appTemplateVersionEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateVersionRepo.Create(ctx, *appTemplateVersionModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(nil, testError).Once()

		appTemplateRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Create(context.TODO(), *appTemplateVersionModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(appTemplateVersionEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.app_template_versions ( id, version, title, correlation_ids, release_date, created_at, app_template_id ) VALUES ( ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixAppTemplateVersionCreateArgs(*appTemplateVersionEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Create(ctx, *appTemplateVersionModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_GetByAppTemplateIDAndVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateVersionEntity).Return(appTemplateVersionModel, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplateversion.Entity{*appTemplateVersionEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, version, title, correlation_ids, release_date, created_at, app_template_id FROM public.app_template_versions WHERE app_template_id = $1 AND version = $2`)).
			WithArgs(appTemplateID, testVersion).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		result, err := appTemplateVersionRepo.GetByAppTemplateIDAndVersion(ctx, appTemplateID, testVersion)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, appTemplateVersionModel, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, version, title, correlation_ids, release_date, created_at, app_template_id FROM public.app_template_versions WHERE app_template_id = $1 AND version = $2`)).
			WithArgs(appTemplateID, testVersion).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateVersionRepo.GetByAppTemplateIDAndVersion(ctx, appTemplateID, testVersion)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateVersionEntity).Return(nil, testError).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplateversion.Entity{*appTemplateVersionEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, version, title, correlation_ids, release_date, created_at, app_template_id FROM public.app_template_versions WHERE app_template_id = $1 AND version = $2`)).
			WithArgs(appTemplateID, testVersion).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateVersionRepo.GetByAppTemplateIDAndVersion(ctx, appTemplateID, testVersion)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_ListByAppTemplateID(t *testing.T) {
	appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
	appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)
	suite := testdb.RepoListTestSuite{
		Name:       "List Application Template Versions by App Template ID",
		MethodName: "ListByAppTemplateID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, version, title, correlation_ids, release_date, created_at, app_template_id FROM public.app_template_versions WHERE app_template_id = $1`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(appTemplateVersionEntity.ID, appTemplateVersionEntity.Version, appTemplateVersionEntity.Title, appTemplateVersionEntity.CorrelationIDs, appTemplateVersionEntity.ReleaseDate, appTemplateVersionEntity.CreatedAt, appTemplateVersionEntity.ApplicationTemplateID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
				Args: []driver.Value{appTemplateID},
			},
		},
		ExpectedModelEntities: []interface{}{appTemplateVersionModel},
		ExpectedDBEntities:    []interface{}{appTemplateVersionEntity},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       apptemplateversion.NewRepository,
		MethodArgs:                []interface{}{appTemplateID},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.app_template_versions WHERE id = $1`)).
			WithArgs(appTemplateVersionID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(nil)

		// WHEN
		result, err := appTemplateVersionRepo.Exists(ctx, appTemplateVersionID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.app_template_versions WHERE id = $1`)).
			WithArgs(appTemplateVersionID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(nil)

		// WHEN
		result, err := appTemplateVersionRepo.Exists(ctx, appTemplateVersionID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(appTemplateVersionEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.app_template_versions SET title = ?,  correlation_ids = ? WHERE id = ?`)).
			WithArgs(appTemplateVersionEntity.Title, appTemplateVersionEntity.CorrelationIDs, appTemplateVersionID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateVersionRepo.Update(ctx, *appTemplateVersionModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when updating", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		appTemplateVersionEntity := fixEntityApplicationTemplateVersion(t, appTemplateVersionID)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(appTemplateVersionEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.app_template_versions SET title = ?,  correlation_ids = ? WHERE id = ?`)).
			WithArgs(appTemplateVersionEntity.Title, appTemplateVersionEntity.CorrelationIDs, appTemplateVersionID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateVersionRepo.Update(ctx, *appTemplateVersionModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateVersionModel := fixModelApplicationTemplateVersion(appTemplateVersionID)
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateVersionModel).Return(nil, testError).Once()

		appTemplateVersionRepo := apptemplateversion.NewRepository(mockConverter)

		// WHEN
		err := appTemplateVersionRepo.Update(context.TODO(), *appTemplateVersionModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}
