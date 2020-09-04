package api_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	// given
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")

	selectQuery := `^SELECT (.+) FROM "public"."api_definitions" WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(apiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", apiDefEntity).Return(model.APIDefinition{ID: apiDefID, Tenant: tenantID}, nil).Once()
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelApiDef, err := pgRepository.GetByID(ctx, tenantID, apiDefID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, modelApiDef.ID)
		assert.Equal(t, tenantID, modelApiDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

}

func TestPgRepository_GetForBundle(t *testing.T) {
	// given
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")

	selectQuery := `^SELECT (.+) FROM "public"."api_definitions" WHERE tenant_id = \$1 AND id = \$2 AND bundle_id = \$3`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(apiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, apiDefID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", apiDefEntity).Return(model.APIDefinition{ID: apiDefID, Tenant: tenantID, BundleID: bundleID}, nil).Once()
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelApiDef, err := pgRepository.GetForBundle(ctx, tenantID, apiDefID, bundleID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, modelApiDef.ID)
		assert.Equal(t, tenantID, modelApiDef.Tenant)
		assert.Equal(t, bundleID, modelApiDef.BundleID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := api.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, apiDefID, bundleID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelApiDef, err := repo.GetForBundle(ctx, tenantID, apiDefID, bundleID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelApiDef)
		require.EqualError(t, err, `Internal Server Error: while getting object from table "public"."api_definitions": test error`)
	})
}

func TestPgRepository_ListForBundle(t *testing.T) {
	// GIVEN
	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityAPIDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityAPIDefinition(secondApiDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."api_definitions" 
		WHERE tenant_id = \$1 AND bundle_id = \$2
		ORDER BY id LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	rawCountQuery := `SELECT COUNT(*) FROM "public"."api_definitions" 
		WHERE tenant_id = $1 AND bundle_id = $2`
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{ID: firstApiDefID}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{ID: secondApiDefID}, nil)
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.ListForBundle(ctx, tenantID, bundleID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef.Data, 2)
		assert.Equal(t, firstApiDefID, modelAPIDef.Data[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDef.Data[1].ID)
		assert.Equal(t, "", modelAPIDef.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelAPIDef.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	apiDefModel := fixFullAPIDefinitionModel("placeholder")
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")
	insertQuery := `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixAPICreateArgs(apiDefID, &apiDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiDefModel).Return(apiDefEntity, nil).Once()
		pgRepository := api.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, &apiDefModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.APIDefinitionConverter{}
		pgRepository := api.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first := fixFullAPIDefinitionModel("first")
		second := fixFullAPIDefinitionModel("second")
		third := fixFullAPIDefinitionModel("third")
		items := []*model.APIDefinition{&first, &second, &third}

		convMock := &automock.APIDefinitionConverter{}
		for _, item := range items {
			convMock.On("ToEntity", *item).Return(fixFullEntityAPIDefinition(item.ID, item.Title), nil).Once()
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixAPICreateArgs(item.ID, item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := api.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, items)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."api_definitions" SET name = ?, description = ?, group_name = ?, 
		entry_point = ?, spec_data = ?, spec_format = ?, spec_type = ?, version = ?, 
		version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiModel := fixFullAPIDefinitionModel("update")
		entity := fixFullEntityAPIDefinition(apiDefID, "update")

		convMock := &automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.Title, entity.Description, entity.Group, entity.EntryPoint, entity.SpecData,
				entity.SpecFormat, entity.SpecType, entity.VersionValue, entity.VersionDepracated,
				entity.VersionDepracatedSince, entity.VersionForRemoval, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := api.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &apiModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		pgRepository := api.NewRepository(convMock)
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
	deleteQuery := `^DELETE FROM "public"."api_definitions" WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, apiDefID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, apiDefID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM "public"."api_definitions" WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, apiDefID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, apiDefID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
