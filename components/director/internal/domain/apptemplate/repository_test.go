package apptemplate_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(appTemplateEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.app_templates ( id, name, description, application_input, placeholders, access_level ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixAppTemplateCreateArgs(*appTemplateEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Create(ctx, *appTemplateModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(nil, testError).Once()

		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Create(context.TODO(), *appTemplateModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(appTemplateEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.app_templates ( id, name, description, application_input, placeholders, access_level ) VALUES ( ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixAppTemplateCreateArgs(*appTemplateEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Create(ctx, *appTemplateModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateEntity).Return(appTemplateModel, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplate.Entity{*appTemplateEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		result, err := appTemplateRepo.Get(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, appTemplateModel, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.Get(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateEntity).Return(nil, testError).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplate.Entity{*appTemplateEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.Get(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_GetByName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateEntity).Return(appTemplateModel, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplate.Entity{*appTemplateEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE name = $1`)).
			WithArgs(testName).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		result, err := appTemplateRepo.GetByName(ctx, testName)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, appTemplateModel, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE name = $1`)).
			WithArgs(testName).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.GetByName(ctx, testName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateEntity).Return(nil, testError).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows([]apptemplate.Entity{*appTemplateEntity})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates WHERE name = $1`)).
			WithArgs(testName).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.GetByName(ctx, testName)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(nil)

		// WHEN
		result, err := appTemplateRepo.Exists(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(nil)

		// WHEN
		result, err := appTemplateRepo.Exists(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.False(t, result)
	})
}

func TestRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateModels := []*model.ApplicationTemplate{
			fixModelAppTemplate("id1", "name1"),
			fixModelAppTemplate("id2", "name2"),
			fixModelAppTemplate("id3", "name3"),
		}

		appTemplateEntities := []apptemplate.Entity{
			*fixEntityAppTemplate(t, "id1", "name1"),
			*fixEntityAppTemplate(t, "id2", "name2"),
			*fixEntityAppTemplate(t, "id3", "name3"),
		}

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		for i := range appTemplateEntities {
			mockConverter.On("FromEntity", &appTemplateEntities[i]).Return(appTemplateModels[i], nil).Once()
		}
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows(appTemplateEntities)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.app_templates`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		result, err := appTemplateRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, appTemplateModels, result.Data)
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateEntities := []apptemplate.Entity{
			*fixEntityAppTemplate(t, "id1", "name1"),
		}

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		mockConverter.On("FromEntity", &appTemplateEntities[0]).Return(nil, testError).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRows(appTemplateEntities)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.app_templates`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, application_input, placeholders, access_level FROM public.app_templates ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		_, err := appTemplateRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(appTemplateEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.app_templates SET name = ?, description = ?, application_input = ?, placeholders = ?, access_level = ? WHERE id = ?`)).
			WithArgs(appTemplateEntity.Name, appTemplateEntity.Description, appTemplateEntity.ApplicationInputJSON, appTemplateEntity.PlaceholdersJSON, appTemplateEntity.AccessLevel, appTemplateEntity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Update(ctx, *appTemplateModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when updating", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)
		appTemplateEntity := fixEntityAppTemplate(t, testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(appTemplateEntity, nil).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.app_templates SET name = ?, description = ?, application_input = ?, placeholders = ?, access_level = ? WHERE id = ?`)).
			WithArgs(appTemplateEntity.Name, appTemplateEntity.Description, appTemplateEntity.ApplicationInputJSON, appTemplateEntity.PlaceholdersJSON, appTemplateEntity.AccessLevel, appTemplateEntity.ID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Update(ctx, *appTemplateModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})

	t.Run("Error when converting", func(t *testing.T) {
		// GIVEN
		appTemplateModel := fixModelAppTemplate(testID, testName)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", appTemplateModel).Return(nil, testError).Once()

		appTemplateRepo := apptemplate.NewRepository(mockConverter)

		// WHEN
		err := appTemplateRepo.Update(context.TODO(), *appTemplateModel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(nil)

		// WHEN
		err := appTemplateRepo.Delete(ctx, testID)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when deleting", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.app_templates WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		appTemplateRepo := apptemplate.NewRepository(nil)

		// WHEN
		err := appTemplateRepo.Delete(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}
