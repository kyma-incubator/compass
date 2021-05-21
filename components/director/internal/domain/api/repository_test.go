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
		convMock.On("FromEntity", apiDefEntity).Return(model.APIDefinition{Tenant: tenantID, BaseEntity: &model.BaseEntity{ID: apiDefID}}, nil).Once()
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

	selectQuery := fmt.Sprintf(`SELECT (.+) FROM "public"."api_definitions"
		WHERE tenant_id = \$1 AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$2 AND bundle_id = \$3 AND api_def_id IS NOT NULL\) 
		ORDER BY id LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	countQuery := `SELECT COUNT\(\*\) FROM "public"."api_definitions"
		WHERE tenant_id = \$1 AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$2 AND bundle_id = \$3 AND api_def_id IS NOT NULL\)`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, tenantID, bundleID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, tenantID, bundleID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}}, nil)
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

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityAPIDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityAPIDefinition(secondApiDefID, "placeholder")

	selectQuery := `^SELECT (.+) FROM "public"."api_definitions" 
		WHERE tenant_id = \$1 AND app_id = \$2`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}}, nil)
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef, totalCount)
		assert.Equal(t, firstApiDefID, modelAPIDef[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDef[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	apiDefModel, _, _ := fixFullAPIDefinitionModel("placeholder")
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")
	insertQuery := `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixAPICreateArgs(apiDefID, &apiDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiDefModel).Return(&apiDefEntity, nil).Once()
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
		first, _, _ := fixFullAPIDefinitionModel("first")
		second, _, _ := fixFullAPIDefinitionModel("second")
		third, _, _ := fixFullAPIDefinitionModel("third")
		items := []*model.APIDefinition{&first, &second, &third}

		convMock := &automock.APIDefinitionConverter{}
		for _, item := range items {
			ent := fixFullEntityAPIDefinition(item.ID, item.Name)
			convMock.On("ToEntity", *item).Return(&ent, nil).Once()
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
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, successor = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?, target_urls = ?, extensible = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiModel, _, _ := fixFullAPIDefinitionModel("update")
		entity := fixFullEntityAPIDefinition(apiDefID, "update")
		entity.UpdatedAt = &fixedTimestamp
		entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

		convMock := &automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiModel).Return(&entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.PackageID, entity.Name, entity.Description, entity.Group,
				entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.ApiProtocol, entity.Tags, entity.Countries,
				entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.Successor, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
				entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription, entity.TargetURLs, entity.Extensible, tenantID, entity.ID).
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

func TestPgRepository_DeleteAllByBundleID(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `DELETE FROM "public"."api_definitions"
		WHERE tenant_id = \$1 AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE tenant_id = \$2 AND bundle_id = \$3 AND api_def_id IS NOT NULL\)`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, tenantID, bundleID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	err := pgRepository.DeleteAllByBundleID(ctx, tenantID, bundleID)
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
