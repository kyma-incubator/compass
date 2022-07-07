package api_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_GetByID(t *testing.T) {
	entity := fixFullEntityAPIDefinition(apiDefID, "placeholder")
	apiDefModel, _, _ := fixFullAPIDefinitionModel("placeholder")

	suite := testdb.RepoGetTestSuite{
		Name: "Get API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, documentation_labels FROM "public"."api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{apiDefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns()).AddRow(fixAPIDefinitionRow(apiDefID, "placeholder")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ExpectedModelEntity:       &apiDefModel,
		ExpectedDBEntity:          &entity,
		MethodArgs:                []interface{}{tenantID, apiDefID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	entity1 := fixFullEntityAPIDefinition(apiDefID, "placeholder")
	apiDefModel1, _, _ := fixFullAPIDefinitionModel("placeholder")
	entity2 := fixFullEntityAPIDefinition(apiDefID, "placeholder2")
	apiDefModel2, _, _ := fixFullAPIDefinitionModel("placeholder2")

	suite := testdb.RepoListTestSuite{
		Name: "List APIs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, documentation_labels FROM "public"."api_definitions" WHERE app_id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns()).AddRow(fixAPIDefinitionRow(apiDefID, "placeholder")...).AddRow(fixAPIDefinitionRow(apiDefID, "placeholder2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ExpectedModelEntities:     []interface{}{&apiDefModel1, &apiDefModel2},
		ExpectedDBEntities:        []interface{}{&entity1, &entity2},
		MethodArgs:                []interface{}{tenantID, appID},
		MethodName:                "ListByApplicationID",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListAllForBundle(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageBundleID := "emptyPageBundleID"

	onePageBundleID := "onePageBundleID"
	firstAPIDefID := "firstAPIDefID"
	firstAPIDef, _, _ := fixFullAPIDefinitionModelWithID(firstAPIDefID, "placeholder")
	firstEntity := fixFullEntityAPIDefinition(firstAPIDefID, "placeholder")
	firstBundleRef := fixModelBundleReference(onePageBundleID, firstAPIDefID)

	multiplePagesBundleID := "multiplePagesBundleID"

	secondAPIDefID := "secondAPIDefID"
	secondAPIDef, _, _ := fixFullAPIDefinitionModelWithID(secondAPIDefID, "placeholder")
	secondEntity := fixFullEntityAPIDefinition(secondAPIDefID, "placeholder")
	secondBundleRef := fixModelBundleReference(multiplePagesBundleID, secondAPIDefID)

	totalCounts := map[string]int{
		emptyPageBundleID:     0,
		onePageBundleID:       1,
		multiplePagesBundleID: 2,
	}

	suite := testdb.RepoListPageableTestSuite{
		Name: "List APIs for multiple bundles with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, documentation_labels FROM "public"."api_definitions" WHERE id IN ($1, $2) AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{firstAPIDefID, secondAPIDefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns()).AddRow(fixAPIDefinitionRow(firstAPIDefID, "placeholder")...).AddRow(fixAPIDefinitionRow(secondAPIDefID, "placeholder")...)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.APIDefinitionPage{
					Data: []*model.APIDefinition{},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 0,
				},
			},
			{
				ExpectedModelEntities: []interface{}{&firstAPIDef},
				ExpectedDBEntities:    []interface{}{&firstEntity},
				ExpectedPage: &model.APIDefinitionPage{
					Data: []*model.APIDefinition{&firstAPIDef},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{&secondAPIDef},
				ExpectedDBEntities:    []interface{}{&secondEntity},
				ExpectedPage: &model.APIDefinitionPage{
					Data: []*model.APIDefinition{&secondAPIDef},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   pagination.EncodeNextOffsetCursor(0, pageSize),
						HasNextPage: true,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc: api.NewRepository,
		MethodName:          "ListByBundleIDs",
		MethodArgs: []interface{}{tenantID, []string{emptyPageBundleID, onePageBundleID, multiplePagesBundleID},
			[]*model.BundleReference{firstBundleRef, secondBundleRef}, totalCounts, pageSize, cursor},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	var nilAPIDefModel *model.APIDefinition
	apiDefModel, _, _ := fixFullAPIDefinitionModel("placeholder")
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`,
				Args:        fixAPICreateArgs(apiDefID, &apiDefModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ModelEntity:               &apiDefModel,
		DBEntity:                  &apiDefEntity,
		NilModelEntity:            nilAPIDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
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
			convMock.On("ToEntity", item).Return(&ent, nil).Once()
			sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3")).
				WithArgs(tenantID, appID, true).WillReturnRows(testdb.RowWhenObjectExist())
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixAPICreateArgs(item.ID, item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := api.NewRepository(convMock)
		// WHEN
		err := pgRepository.CreateMany(ctx, tenantID, items)
		// THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?,
		updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?,
		target_urls = ?, extensible = ?, successors = ?, resource_hash = ?, documentation_labels = ?
		WHERE id = ? AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilAPIDefModel *model.APIDefinition
	apiModel, _, _ := fixFullAPIDefinitionModel("update")
	entity := fixFullEntityAPIDefinition(apiDefID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.Group,
					entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.APIProtocol, entity.Tags, entity.Countries,
					entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated,
					entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt,
					entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription,
					entity.TargetURLs, entity.Extensible, entity.Successors, entity.ResourceHash, entity.DocumentationLabels, entity.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ModelEntity:               &apiModel,
		DBEntity:                  &entity,
		NilModelEntity:            nilAPIDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "API Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{apiDefID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc: api.NewRepository,
		MethodArgs:          []interface{}{tenantID, apiDefID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteAllByBundleID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "API Delete By BundleID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."api_definitions" WHERE id IN (SELECT api_def_id FROM public.bundle_references WHERE bundle_id = $1 AND api_def_id IS NOT NULL) AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{bundleID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc: api.NewRepository,
		MethodName:          "DeleteAllByBundleID",
		MethodArgs:          []interface{}{tenantID, bundleID},
		IsDeleteMany:        true,
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "API Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM "public"."api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{apiDefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc: api.NewRepository,
		TargetID:            apiDefID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, apiDefID},
	}

	suite.Run(t)
}
