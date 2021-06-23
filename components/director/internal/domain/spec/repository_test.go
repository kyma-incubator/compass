package spec_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetByID(t *testing.T) {
	// given
	selectQuery := `^SELECT (.+) FROM public.specifications WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("Success For API", func(t *testing.T) {
		specEntity := fixAPISpecEntity()
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixAPISpecRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, specID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", specEntity).Return(*fixModelAPISpec(), nil).Once()
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.GetByID(ctx, tenant, specID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, specID, modelSpec.ID)
		assert.Equal(t, tenant, modelSpec.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Success For Event", func(t *testing.T) {
		specEntity := fixEventSpecEntity()
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixEventSpecRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, specID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", specEntity).Return(*fixModelEventSpec(), nil).Once()
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.GetByID(ctx, tenant, specID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, specID, modelSpec.ID)
		assert.Equal(t, tenant, modelSpec.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

}

func TestRepository_Create(t *testing.T) {
	//GIVEN
	insertQuery := `^INSERT INTO public.specifications \(.+\) VALUES \(.+\)$`

	t.Run("Success for API", func(t *testing.T) {
		specModel := fixModelAPISpec()
		specEntity := fixAPISpecEntity()
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixAPISpecCreateArgs(specModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.Converter{}
		convMock.On("ToEntity", *specModel).Return(specEntity, nil).Once()
		pgRepository := spec.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, specModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success for Event", func(t *testing.T) {
		specModel := fixModelEventSpec()
		specEntity := fixEventSpecEntity()
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventSpecCreateArgs(specModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.Converter{}
		convMock.On("ToEntity", *specModel).Return(specEntity, nil).Once()
		pgRepository := spec.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, specModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.Converter{}
		pgRepository := spec.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item can not be empty")
		convMock.AssertExpectations(t)
	})
}

func TestRepository_ListByReferenceObjectID(t *testing.T) {
	// GIVEN

	t.Run("Success for API", func(t *testing.T) {
		firstSpecID := "111111111-1111-1111-1111-111111111111"
		firstSpecEntity := fixAPISpecEntityWithID(firstSpecID)
		secondSpecID := "222222222-2222-2222-2222-222222222222"
		secondApiDefEntity := fixAPISpecEntityWithID(secondSpecID)

		selectQuery := `^SELECT (.+) FROM public.specifications 
		WHERE tenant_id = \$1 AND api_def_id = \$2
		ORDER BY created_at`

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixAPISpecRowWithID(firstSpecID)...).
			AddRow(fixAPISpecRowWithID(secondSpecID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, apiID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstSpecEntity).Return(*fixModelAPISpecWithID(firstSpecID), nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(*fixModelAPISpecWithID(secondSpecID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectID(ctx, tenant, model.APISpecReference, apiID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Success for Event", func(t *testing.T) {
		firstSpecID := "111111111-1111-1111-1111-111111111111"
		firstSpecEntity := fixEventSpecEntityWithID(firstSpecID)
		secondSpecID := "222222222-2222-2222-2222-222222222222"
		secondApiDefEntity := fixEventSpecEntityWithID(secondSpecID)

		selectQuery := `^SELECT (.+) FROM public.specifications 
		WHERE tenant_id = \$1 AND event_def_id = \$2
		ORDER BY created_at`

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixSpecColumns()).
			AddRow(fixEventSpecRowWithID(firstSpecID)...).
			AddRow(fixEventSpecRowWithID(secondSpecID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenant, eventID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		convMock.On("FromEntity", firstSpecEntity).Return(*fixModelEventSpecWithID(firstSpecID), nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(*fixModelEventSpecWithID(secondSpecID), nil)
		pgRepository := spec.NewRepository(convMock)
		// WHEN
		modelSpec, err := pgRepository.ListByReferenceObjectID(ctx, tenant, model.EventSpecReference, eventID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelSpec, 2)
		assert.Equal(t, firstSpecID, modelSpec[0].ID)
		assert.Equal(t, secondSpecID, modelSpec[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM public.specifications WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenant, specID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.Converter{}
	pgRepository := spec.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenant, specID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestRepository_DeleteByReferenceObjectID(t *testing.T) {
	t.Run("Success for API", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		deleteQuery := `^DELETE FROM public.specifications WHERE tenant_id = \$1 AND api_def_id = \$2$`

		sqlMock.ExpectExec(deleteQuery).WithArgs(tenant, apiID).WillReturnResult(sqlmock.NewResult(-1, 1))
		convMock := &automock.Converter{}
		pgRepository := spec.NewRepository(convMock)
		//WHEN
		err := pgRepository.DeleteByReferenceObjectID(ctx, tenant, model.APISpecReference, apiID)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Success for Event", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		deleteQuery := `^DELETE FROM public.specifications WHERE tenant_id = \$1 AND event_def_id = \$2$`

		sqlMock.ExpectExec(deleteQuery).WithArgs(tenant, eventID).WillReturnResult(sqlmock.NewResult(-1, 1))
		convMock := &automock.Converter{}
		pgRepository := spec.NewRepository(convMock)
		//WHEN
		err := pgRepository.DeleteByReferenceObjectID(ctx, tenant, model.EventSpecReference, eventID)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

func TestRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.specifications SET spec_data = ?, api_spec_format = ?, api_spec_type = ?,
		event_spec_format = ?, event_spec_type = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("Success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		specModel := fixModelAPISpec()
		entity := fixAPISpecEntity()

		convMock := &automock.Converter{}
		convMock.On("ToEntity", *specModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.SpecData, entity.APISpecFormat, entity.APISpecType, entity.EventSpecFormat, entity.EventSpecType, tenant, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := spec.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, specModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("Returns error when item is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.Converter{}
		pgRepository := spec.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.specifications WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenant, specID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.Converter{}
	pgRepository := spec.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenant, specID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
