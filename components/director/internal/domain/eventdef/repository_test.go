package eventdef_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	// given
	eventAPIDefEntity := fixFullEventDef(eventAPIID, "placeholder")
	selectQuery := `^SELECT (.+) FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2$`
	testError := errors.New("test error")

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventDefinition{ID: eventAPIID, Tenant: tenantID}, nil).Once()
		pgRepository := eventdef.NewRepository(convMock)
		// WHEN
		modelEventAPIDef, err := pgRepository.GetByID(ctx, tenantID, eventAPIID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventAPIID, modelEventAPIDef.ID)
		assert.Equal(t, tenantID, modelEventAPIDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventDefinition{}, testError).Once()
		pgRepository := eventdef.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, eventAPIID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when get operation failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := eventdef.NewRepository(nil)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, eventAPIID)
		//THEN
		require.Error(t, err)
		assert.Error(t, err, testError)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_GetForApplication(t *testing.T) {
	// given
	eventAPIDefEntity := fixFullEventDef(eventAPIID, "placeholder")
	selectQuery := `^SELECT (.+) FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2 AND app_id = \$3`
	testError := errors.New("test error")

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventDefinition{ID: eventAPIID, Tenant: tenantID, ApplicationID: appID}, nil).Once()
		pgRepository := eventdef.NewRepository(convMock)
		// WHEN
		modelEventAPIDef, err := pgRepository.GetForApplication(ctx, tenantID, eventAPIID, appID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventAPIID, modelEventAPIDef.ID)
		assert.Equal(t, tenantID, modelEventAPIDef.Tenant)
		assert.Equal(t, appID, modelEventAPIDef.ApplicationID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := eventdef.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID, appID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		eventApiDef, err := repo.GetForApplication(ctx, tenantID, eventAPIID, appID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, eventApiDef)
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", testError.Error()))
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventDefinition{}, testError).Once()
		pgRepository := eventdef.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetForApplication(ctx, tenantID, eventAPIID, appID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when get operation failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID, appID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := eventdef.NewRepository(nil)
		// WHEN
		_, err := pgRepository.GetForApplication(ctx, tenantID, eventAPIID, appID)
		//THEN
		require.Error(t, err)
		assert.Error(t, err, testError)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	firstEventAPIDefID := "111111111-1111-1111-1111-111111111111"
	firstEventAPIDefEntity := fixFullEventDef(firstEventAPIDefID, "placeholder")
	secondEventAPIDefID := "222222222-2222-2222-2222-222222222222"
	secondEventAPIDefEntity := fixFullEventDef(secondEventAPIDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."event_api_definitions" 
		WHERE tenant_id=\$1 AND app_id = '%s' 
		ORDER BY id LIMIT %d OFFSET %d`, appID, ExpectedLimit, ExpectedOffset)

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "public"."event_api_definitions" 
		WHERE tenant_id=$1 AND app_id = '%s'`, appID)
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventAPIDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondEventAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventAPIDefEntity).Return(model.EventDefinition{ID: firstEventAPIDefID}, nil)
		convMock.On("FromEntity", secondEventAPIDefEntity).Return(model.EventDefinition{ID: secondEventAPIDefID}, nil)
		pgRepository := eventdef.NewRepository(convMock)
		// WHEN
		modelEventAPIDef, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventAPIDef.Data, 2)
		assert.Equal(t, firstEventAPIDefID, modelEventAPIDef.Data[0].ID)
		assert.Equal(t, secondEventAPIDefID, modelEventAPIDef.Data[1].ID)
		assert.Equal(t, "", modelEventAPIDef.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelEventAPIDef.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from entity to model failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(1))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventAPIDefEntity).Return(model.EventDefinition{}, testErr).Once()
		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		_, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when list operation failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := eventdef.NewRepository(nil)
		// WHEN
		_, err := pgRepository.ListByApplicationID(ctx, tenantID, appID, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		assert.Error(t, err, testErr)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	eventAPIDefModel := fixFullModelEventDefinition(eventAPIID, "placeholder")
	eventAPIDefEntity := fixFullEventDef(eventAPIID, "placeholder")
	insertQuery := `^INSERT INTO "public"."event_api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventCreateArgs(eventAPIID, eventAPIDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIDefModel).Return(eventAPIDefEntity, nil).Once()
		pgRepository := eventdef.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, &eventAPIDefModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIDefModel).Return(eventdef.Entity{}, errors.New("test error"))
		pgRepository := eventdef.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, &eventAPIDefModel)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EventAPIDefinitionConverter{}
		pgRepository := eventdef.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when create operation failed", func(t *testing.T) {
		testErr := errors.New("test error")
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIDefModel).Return(eventAPIDefEntity, nil).Once()
		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventCreateArgs(eventAPIID, eventAPIDefModel)...).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := eventdef.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, &eventAPIDefModel)
		//THEN
		require.Error(t, err)
		assert.Error(t, err, testErr)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."event_api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first := fixFullModelEventDefinition(eventAPIID, "first")
		second := fixFullModelEventDefinition(eventAPIID, "second")
		third := fixFullModelEventDefinition(eventAPIID, "third")
		items := []*model.EventDefinition{&first, &second, &third}

		convMock := &automock.EventAPIDefinitionConverter{}
		for _, item := range items {
			convMock.On("ToEntity", *item).Return(fixFullEventDef(item.ID, item.Name), nil).Once()
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixEventCreateArgs(item.ID, *item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, items)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		eventAPIModel := fixFullModelEventDefinition(eventAPIID, "api")
		require.NotNil(t, eventAPIModel)
		items := []*model.EventDefinition{&eventAPIModel}

		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(eventdef.Entity{}, errors.New("test error"))
		pgRepository := eventdef.NewRepository(&convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, items)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when create operation failed", func(t *testing.T) {
		testErr := errors.New("test error")

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		item := fixFullModelEventDefinition(eventAPIID, "first")

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", item).Return(fixFullEventDef(item.ID, item.Name), nil).Once()
		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventCreateArgs(item.ID, item)...).
			WillReturnError(testErr)
		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, []*model.EventDefinition{&item})
		//THEN
		require.Error(t, err)
		assert.Error(t, err, testErr)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."event_api_definitions" SET name = ?, description = ?, group_name = ?, 
		spec_data = ?, spec_format = ?, spec_type = ?, version_value = ?, version_deprecated = ?, 
		version_deprecated_since = ?, version_for_removal = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		eventAPIModel := fixFullModelEventDefinition(eventAPIID, "update")
		entity := fixFullEventDef(eventAPIID, "update")

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.Name, entity.Description, entity.GroupName, entity.SpecData, entity.SpecFormat,
				entity.SpecType, entity.VersionValue, entity.VersionDepracated, entity.VersionDepracatedSince,
				entity.VersionForRemoval, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &eventAPIModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when conversion from model to entity failed", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		eventAPIModel := model.EventDefinition{}
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(eventdef.Entity{}, errors.New("test error")).Once()
		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &eventAPIModel)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		pgRepository := eventdef.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, eventAPIID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := eventdef.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, eventAPIID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_DeleteAllByApplicationID(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND app_id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, appID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := eventdef.NewRepository(convMock)
	//WHEN
	err := pgRepository.DeleteAllByApplicationID(ctx, tenantID, appID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM "public"."event_api_definitions" WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, eventAPIID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := eventdef.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, eventAPIID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
