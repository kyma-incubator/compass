package api_test

import (
	"context"
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	entity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder")
	apiDefModel, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder")

	suite := testdb.RepoGetTestSuite{
		Name: "Get API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, hierarchy, supported_use_cases, documentation_labels FROM "public"."api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2))`),
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

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	entity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder")
	apiDefModel, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Global API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, hierarchy, supported_use_cases, documentation_labels FROM "public"."api_definitions" WHERE id = $1`),
				Args:     []driver.Value{apiDefID},
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
		MethodName:                "GetByIDGlobal",
		MethodArgs:                []interface{}{apiDefID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	entity1App := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder")
	apiDefModel1App, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder")
	entity2App := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder2")
	apiDefModel2App, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder2")
	entity1AppTemplateVersion := fixFullEntityAPIDefinitionWithAppTemplateVersionID(apiDefID, "placeholder")
	apiDefModel1AppTemplateVersion, _, _ := fixFullAPIDefinitionModelWithAppTemplateVersionID("placeholder")
	entity2AppTemplateVersion := fixFullEntityAPIDefinitionWithAppTemplateVersionID(apiDefID, "placeholder2")
	apiDefModel2AppTemplateVersion, _, _ := fixFullAPIDefinitionModelWithAppTemplateVersionID("placeholder2")

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List APIs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, hierarchy, supported_use_cases, documentation_labels FROM "public"."api_definitions" WHERE app_id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2)) FOR UPDATE`),
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
		ExpectedModelEntities:     []interface{}{&apiDefModel1App, &apiDefModel2App},
		ExpectedDBEntities:        []interface{}{&entity1App, &entity2App},
		MethodArgs:                []interface{}{tenantID, resource.Application, appID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List APIs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, hierarchy, supported_use_cases, documentation_labels FROM "public"."api_definitions" WHERE app_template_version_id = $1
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAPIDefinitionColumns()).AddRow(fixAPIDefinitionRowForAppTemplateVersion(apiDefID, "placeholder")...).AddRow(fixAPIDefinitionRowForAppTemplateVersion(apiDefID, "placeholder2")...)}
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
		ExpectedModelEntities:     []interface{}{&apiDefModel1AppTemplateVersion, &apiDefModel2AppTemplateVersion},
		ExpectedDBEntities:        []interface{}{&entity1AppTemplateVersion, &entity2AppTemplateVersion},
		MethodArgs:                []interface{}{tenantID, resource.ApplicationTemplateVersion, appTemplateVersionID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_ListAllForBundle(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageBundleID := "emptyPageBundleID"

	onePageBundleID := "onePageBundleID"
	firstAPIDefID := "firstAPIDefID"
	firstAPIDef, _, _ := fixFullAPIDefinitionModel(firstAPIDefID, "placeholder")
	firstAPIDef.ApplicationID = str.Ptr(appID)

	firstEntity := fixFullEntityAPIDefinitionWithAppID(firstAPIDefID, "placeholder")
	firstBundleRef := fixModelBundleReference(onePageBundleID, firstAPIDefID)

	multiplePagesBundleID := "multiplePagesBundleID"

	secondAPIDefID := "secondAPIDefID"
	secondAPIDef, _, _ := fixFullAPIDefinitionModel(secondAPIDefID, "placeholder")
	secondAPIDef.ApplicationID = str.Ptr(appID)
	secondEntity := fixFullEntityAPIDefinitionWithAppID(secondAPIDefID, "placeholder")
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
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, api_protocol, tags, countries, links, api_resource_links, release_status, sunset_date, changelog_entries, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, target_urls, extensible, successors, resource_hash, hierarchy, supported_use_cases, documentation_labels FROM "public"."api_definitions" WHERE id IN ($1, $2) AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $3))`),
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
	apiDefModel, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder")
	apiDefEntity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder")

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

func TestPgRepository_CreateGlobal(t *testing.T) {
	var nilAPIDefModel *model.APIDefinition
	apiDefModel, _, _ := fixFullAPIDefinitionModelWithAppTemplateVersionID("placeholder")
	apiDefEntity := fixFullEntityAPIDefinitionWithAppTemplateVersionID(apiDefID, "placeholder")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`,
				Args:        fixAPICreateArgsForAppTemplateVersion(apiDefID, &apiDefModel),
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
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first, _, _ := fixFullAPIDefinitionModelWithAppID("first")
		second, _, _ := fixFullAPIDefinitionModelWithAppID("second")
		third, _, _ := fixFullAPIDefinitionModelWithAppID("third")
		items := []*model.APIDefinition{&first, &second, &third}

		convMock := &automock.APIDefinitionConverter{}
		for _, item := range items {
			ent := fixFullEntityAPIDefinitionWithAppID(item.ID, item.Name)
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
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?, local_tenant_id = ?,
		short_description = ?, system_instance_aware = ?, policy_level = ?, custom_policy_level = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?,
		updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?,
		target_urls = ?, extensible = ?, successors = ?, resource_hash = ?, hierarchy = ?, supported_use_cases = ?, documentation_labels = ?
		WHERE id = ? AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilAPIDefModel *model.APIDefinition
	apiModel, _, _ := fixFullAPIDefinitionModelWithAppID("update")
	entity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update API",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.Group,
					entity.OrdID, entity.LocalTenantID, entity.ShortDescription, entity.SystemInstanceAware, entity.PolicyLevel, entity.CustomPolicyLevel, entity.APIProtocol, entity.Tags, entity.Countries,
					entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated,
					entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt,
					entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription,
					entity.TargetURLs, entity.Extensible, entity.Successors, entity.ResourceHash, entity.Hierarchy, entity.SupportedUseCases, entity.DocumentationLabels, entity.ID, tenantID},
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

func TestPgRepository_UpdateGlobal(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?, local_tenant_id = ?,
		short_description = ?, system_instance_aware = ?, policy_level = ?, custom_policy_level = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?,
		updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?,
		target_urls = ?, extensible = ?, successors = ?, resource_hash = ?, hierarchy = ?, supported_use_cases = ?, documentation_labels = ?
		WHERE id = ?`)

	var nilAPIDefModel *model.APIDefinition
	apiModel, _, _ := fixFullAPIDefinitionModelWithAppID("update")
	entity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Global API Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.Group,
					entity.OrdID, entity.LocalTenantID, entity.ShortDescription, entity.SystemInstanceAware, entity.PolicyLevel, entity.CustomPolicyLevel, entity.APIProtocol, entity.Tags, entity.Countries,
					entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated,
					entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt,
					entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription,
					entity.TargetURLs, entity.Extensible, entity.Successors, entity.ResourceHash, entity.Hierarchy, entity.SupportedUseCases, entity.DocumentationLabels, entity.ID},
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
		DisableConverterErrorTest: true,
		UpdateMethodName:          "UpdateGlobal",
		IsGlobal:                  true,
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
