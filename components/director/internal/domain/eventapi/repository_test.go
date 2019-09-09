package eventapi_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	// given
	eventAPIDefEntity := fixFullEventAPIDef(eventAPIID, "placeholder")
	selectQuery := `^SELECT (.+) FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2$`
	testError := errors.New("test error")

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventAPIDefinitionColumns()).
			AddRow(fixEventAPIDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventAPIDefinition{ID: eventAPIID, Tenant: tenantID}, nil).Once()
		pgRepository := eventapi.NewRepository(convMock)
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
		rows := sqlmock.NewRows(fixEventAPIDefinitionColumns()).
			AddRow(fixEventAPIDefinitionRow(eventAPIID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventAPIID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventAPIDefEntity).Return(model.EventAPIDefinition{}, testError).Once()
		pgRepository := eventapi.NewRepository(convMock)
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
		pgRepository := eventapi.NewRepository(nil)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, eventAPIID)
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
	firstEventAPIDefEntity := fixFullEventAPIDef(firstEventAPIDefID, "placeholder")
	secondEventAPIDefID := "222222222-2222-2222-2222-222222222222"
	secondEventAPIDefEntity := fixFullEventAPIDef(secondEventAPIDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."event_api_definitions" 
		WHERE tenant_id=\$1 AND app_id = '%s' 
		ORDER BY id LIMIT %d OFFSET %d`, appID, ExpectedLimit, ExpectedOffset)

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "public"."event_api_definitions" 
		WHERE tenant_id=$1 AND app_id = '%s'`, appID)
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventAPIDefinitionColumns()).
			AddRow(fixEventAPIDefinitionRow(firstEventAPIDefID, "placeholder")...).
			AddRow(fixEventAPIDefinitionRow(secondEventAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventAPIDefEntity).Return(model.EventAPIDefinition{ID: firstEventAPIDefID}, nil)
		convMock.On("FromEntity", secondEventAPIDefEntity).Return(model.EventAPIDefinition{ID: secondEventAPIDefID}, nil)
		pgRepository := eventapi.NewRepository(convMock)
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
		rows := sqlmock.NewRows(fixEventAPIDefinitionColumns()).
			AddRow(fixEventAPIDefinitionRow(firstEventAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(1))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventAPIDefEntity).Return(model.EventAPIDefinition{}, testErr).Once()
		pgRepository := eventapi.NewRepository(convMock)
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
		pgRepository := eventapi.NewRepository(nil)
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
	eventAPIDefModel := fixFullModelEventAPIDefinition(eventAPIID, "placeholder")
	eventAPIDefEntity := fixFullEventAPIDef(eventAPIID, "placeholder")
	insertQuery := `^INSERT INTO "public"."event_api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventAPICreateArgs(eventAPIID, eventAPIDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIDefModel).Return(eventAPIDefEntity, nil).Once()
		pgRepository := eventapi.NewRepository(&convMock)
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
		convMock.On("ToEntity", eventAPIDefModel).Return(eventapi.Entity{}, errors.New("test error"))
		pgRepository := eventapi.NewRepository(&convMock)
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
		pgRepository := eventapi.NewRepository(&convMock)
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
			WithArgs(fixEventAPICreateArgs(eventAPIID, eventAPIDefModel)...).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := eventapi.NewRepository(&convMock)
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
		first := fixFullModelEventAPIDefinition(eventAPIID, "first")
		second := fixFullModelEventAPIDefinition(eventAPIID, "second")
		third := fixFullModelEventAPIDefinition(eventAPIID, "third")
		items := []*model.EventAPIDefinition{&first, &second, &third}

		convMock := &automock.EventAPIDefinitionConverter{}
		for _, item := range items {
			convMock.On("ToEntity", *item).Return(fixFullEventAPIDef(item.ID, item.Name), nil).Once()
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixEventAPICreateArgs(item.ID, *item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := eventapi.NewRepository(convMock)
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
		eventAPIModel := fixFullModelEventAPIDefinition(eventAPIID, "api")
		require.NotNil(t, eventAPIModel)
		items := []*model.EventAPIDefinition{&eventAPIModel}

		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(eventapi.Entity{}, errors.New("test error"))
		pgRepository := eventapi.NewRepository(&convMock)
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
		item := fixFullModelEventAPIDefinition(eventAPIID, "first")

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", item).Return(fixFullEventAPIDef(item.ID, item.Name), nil).Once()
		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventAPICreateArgs(item.ID, item)...).
			WillReturnError(testErr)
		pgRepository := eventapi.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, []*model.EventAPIDefinition{&item})
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
		eventAPIModel := fixFullModelEventAPIDefinition(eventAPIID, "update")
		entity := fixFullEventAPIDef(eventAPIID, "update")

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.Name, entity.Description, entity.GroupName, entity.SpecData, entity.SpecFormat,
				entity.SpecType, entity.VersionValue, entity.VersionDepracated, entity.VersionDepracatedSince,
				entity.VersionForRemoval, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := eventapi.NewRepository(convMock)
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
		eventAPIModel := model.EventAPIDefinition{}
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventAPIModel).Return(eventapi.Entity{}, errors.New("test error")).Once()
		pgRepository := eventapi.NewRepository(convMock)
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
		pgRepository := eventapi.NewRepository(convMock)
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
	pgRepository := eventapi.NewRepository(convMock)
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
	pgRepository := eventapi.NewRepository(convMock)
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
	pgRepository := eventapi.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, eventAPIID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
