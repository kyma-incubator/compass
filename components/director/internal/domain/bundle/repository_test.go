package bundle_test

import (
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	name := "foo"
	desc := "bar"

	var nilBundleMode *model.Bundle
	bndlModel := fixBundleModel(name, desc)
	bndlEntity := fixEntityBundleWithAppID(bundleID, name, desc)

	defAuth, err := json.Marshal(bndlModel.DefaultInstanceAuth)
	require.NoError(t, err)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Bundle",
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
				Query:       `^INSERT INTO public.bundles \(.+\) VALUES \(.+\)$`,
				Args:        fixBundleCreateArgsForApp(string(defAuth), *bndlModel.InstanceAuthRequestInputSchema, bndlModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndlModel,
		DBEntity:            bndlEntity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	name := "foo"

	var nilBundleMode *model.Bundle
	bndlModel := fixBundleModelWithIDAndAppTemplateVersionID(bundleID, name, desc)
	bndlEntity := fixEntityBundleWithAppTemplateVersionID(bundleID, name, desc)

	defAuth, err := json.Marshal(bndlModel.DefaultInstanceAuth)
	require.NoError(t, err)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Bundle Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.bundles \(.+\) VALUES \(.+\)$`,
				Args:        fixBundleCreateArgsForAppTemplateVersion(string(defAuth), *bndlModel.InstanceAuthRequestInputSchema, bndlModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndlModel,
		DBEntity:            bndlEntity,
		NilModelEntity:      nilBundleMode,
		IsGlobal:            true,
		MethodName:          "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.bundles SET name = ?, description = ?, version = ?, instance_auth_request_json_schema = ?, default_instance_auth = ?, ord_id = ?, local_tenant_id = ?, short_description = ?, links = ?, labels = ?, credential_exchange_strategies = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, tags = ?, resource_hash = ?, documentation_labels = ? WHERE id = ? AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilBundleMode *model.Bundle
	bndl := fixBundleModel("foo", "update")
	entity := fixEntityBundleWithAppID(bundleID, "foo", "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateQuery,
				Args:          []driver.Value{entity.Name, entity.Description, entity.Version, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, entity.OrdID, entity.LocalTenantID, entity.ShortDescription, entity.Links, entity.Labels, entity.CredentialExchangeStrategies, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.Tags, entity.ResourceHash, entity.DocumentationLabels, entity.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndl,
		DBEntity:            entity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.bundles SET name = ?, description = ?, version = ?, instance_auth_request_json_schema = ?, default_instance_auth = ?, ord_id = ?, local_tenant_id = ?, short_description = ?, links = ?, labels = ?, credential_exchange_strategies = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, tags = ?, resource_hash = ?, documentation_labels = ? WHERE id = ?`)

	var nilBundleMode *model.Bundle
	bndl := fixBundleModel("foo", "update")
	entity := fixEntityBundleWithAppID(bundleID, "foo", "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Bundle Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateQuery,
				Args:          []driver.Value{entity.Name, entity.Description, entity.Version, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, entity.OrdID, entity.LocalTenantID, entity.ShortDescription, entity.Links, entity.Labels, entity.CredentialExchangeStrategies, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.Tags, entity.ResourceHash, entity.DocumentationLabels, entity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndl,
		DBEntity:            entity,
		NilModelEntity:      nilBundleMode,
		IsGlobal:            true,
		UpdateMethodName:    "UpdateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Bundle Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{bundleID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Bundle Delete Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.bundles WHERE id = $1`),
				Args:          []driver.Value{bundleID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		MethodArgs:          []interface{}{bundleID},
		IsGlobal:            true,
		MethodName:          "DeleteGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Bundle Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
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
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		TargetID:            bundleID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	bndlEntity := fixEntityBundleWithAppID(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRowWithAppID(bundleID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	bndlEntity := fixEntityBundleWithAppTemplateVersionID(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE id = $1`),
				Args:     []driver.Value{bundleID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRowWithAppTemplateVersionID(bundleID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{bundleID},
		MethodName:          "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_GetForApplication(t *testing.T) {
	bndlEntity := fixEntityBundleWithAppID(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle For Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{bundleID, appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRowWithAppID(bundleID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID, appID},
		MethodName:          "GetForApplication",
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageAppID := "emptyPageAppID"

	onePageAppID := "onePageAppID"
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstBundleEntity := fixEntityBundleWithAppID(firstBundleID, "foo", "bar")
	firstBundleEntity.ApplicationID = repo.NewValidNullableString(onePageAppID)
	firstBndlModel := fixBundleModelWithIDAndAppID(firstBundleID, "foo", desc)
	firstBndlModel.ApplicationID = &onePageAppID

	multiplePagesAppID := "multiplePagesAppID"

	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondBundleEntity := fixEntityBundleWithAppID(secondBundleID, "foo", "bar")
	secondBundleEntity.ApplicationID = repo.NewValidNullableString(multiplePagesAppID)
	secondBndlModel := fixBundleModelWithIDAndAppID(secondBundleID, "foo", desc)
	secondBndlModel.ApplicationID = &multiplePagesAppID

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Bundles for multiple Applications with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $1)) AND app_id = $2 ORDER BY app_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $5)) AND app_id = $6 ORDER BY app_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $9)) AND app_id = $10 ORDER BY app_id ASC, id ASC LIMIT $11 OFFSET $12)`),

				Args:     []driver.Value{tenantID, emptyPageAppID, pageSize, 0, tenantID, onePageAppID, pageSize, 0, tenantID, multiplePagesAppID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).AddRow(fixBundleRowWithCustomAppID(firstBundleID, onePageAppID)...).AddRow(fixBundleRowWithCustomAppID(secondBundleID, multiplePagesAppID)...)}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT app_id AS id, COUNT(*) AS total_count FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $1)) GROUP BY app_id ORDER BY app_id ASC`),
				Args:     []driver.Value{tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(emptyPageAppID, 0).AddRow(onePageAppID, 1).AddRow(multiplePagesAppID, 2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.BundlePage{
					Data: nil,
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 0,
				},
			},
			{
				ExpectedModelEntities: []interface{}{firstBndlModel},
				ExpectedDBEntities:    []interface{}{firstBundleEntity},
				ExpectedPage: &model.BundlePage{
					Data: []*model.Bundle{firstBndlModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{secondBndlModel},
				ExpectedDBEntities:    []interface{}{secondBundleEntity},
				ExpectedPage: &model.BundlePage{
					Data: []*model.Bundle{secondBndlModel},
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
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		MethodName:          "ListByApplicationIDs",
		MethodArgs:          []interface{}{tenantID, []string{emptyPageAppID, onePageAppID, multiplePagesAppID}, pageSize, cursor},
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceIDNoPaging(t *testing.T) {
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstAppBundleEntity := fixEntityBundleWithAppID(firstBundleID, "foo", "bar")
	firstAppTemplateVersionBundleEntity := fixEntityBundleWithAppTemplateVersionID(firstBundleID, "foo", "bar")
	firstAppBndlModel := fixBundleModelWithIDAndAppID(firstBundleID, "foo", desc)
	firstAppTemplateVersionBndlModel := fixBundleModelWithIDAndAppTemplateVersionID(firstBundleID, "foo", desc)
	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondAppBundleEntity := fixEntityBundleWithAppID(secondBundleID, "foo", "bar")
	secondAppTemplateVersionBundleEntity := fixEntityBundleWithAppTemplateVersionID(secondBundleID, "foo", "bar")
	secondAppBndlModel := fixBundleModelWithIDAndAppID(secondBundleID, "foo", desc)
	secondAppTemplateVersionBndlModel := fixBundleModelWithIDAndAppTemplateVersionID(secondBundleID, "foo", desc)

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List Bundles No Paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE app_id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).
						AddRow(fixBundleRowWithAppID(firstBundleID)...).AddRow(fixBundleRowWithAppID(secondBundleID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundle.NewRepository,
		ExpectedModelEntities: []interface{}{firstAppBndlModel, secondAppBndlModel},
		ExpectedDBEntities:    []interface{}{firstAppBundleEntity, secondAppBundleEntity},
		MethodArgs:            []interface{}{tenantID, appID, resource.Application},
		MethodName:            "ListByResourceIDNoPaging",
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Bundles No Paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM public.bundles WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).
						AddRow(fixBundleRowWithAppTemplateVersionID(firstBundleID)...).AddRow(fixBundleRowWithAppTemplateVersionID(secondBundleID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundle.NewRepository,
		ExpectedModelEntities: []interface{}{firstAppTemplateVersionBndlModel, secondAppTemplateVersionBndlModel},
		ExpectedDBEntities:    []interface{}{firstAppTemplateVersionBundleEntity, secondAppTemplateVersionBundleEntity},
		MethodArgs:            []interface{}{tenantID, appTemplateVersionID, resource.ApplicationTemplateVersion},
		MethodName:            "ListByResourceIDNoPaging",
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_ListByDestination(t *testing.T) {
	bndlEntity := fixEntityBundleWithAppID(bundleID, "foo", "bar")
	modelBundle := fixBundleModel("foo", "bar")

	destinationWithSystemName := model.DestinationInput{
		XSystemBaseURL:    "http://localhost",
		XSystemTenantName: "system_name",
		XCorrelationID:    "correlation_id",
	}

	suiteBySystemName := testdb.RepoListTestSuite{
		Name: "List Bundles By Destination with system name",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema, 
					default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies,
					ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM 
					public.bundles WHERE app_id IN (
						SELECT id
						FROM public.applications
						WHERE id IN (
							SELECT id
							FROM tenant_applications
							WHERE tenant_id=(SELECT parent FROM business_tenant_mappings WHERE id = $1 )
						)
						AND name = $2 AND base_url = $3
				) AND correlation_ids ?| array[$4]`),
				Args: []driver.Value{tenantID, destinationWithSystemName.XSystemTenantName,
					destinationWithSystemName.XSystemBaseURL, destinationWithSystemName.XCorrelationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).
						AddRow(fixBundleRowWithAppID(bundleID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundle.NewRepository,
		ExpectedModelEntities: []interface{}{modelBundle},
		ExpectedDBEntities:    []interface{}{bndlEntity},
		MethodArgs:            []interface{}{tenantID, destinationWithSystemName},
		MethodName:            "ListByDestination",
	}

	suiteBySystemName.Run(t)

	destinationWithSystemID := model.DestinationInput{
		XSystemType:     "system_type",
		XSystemTenantID: "system_id",
		XCorrelationID:  "correlation_id",
	}

	suiteBySystemID := testdb.RepoListTestSuite{
		Name: "List Bundles By Destination with system ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, name, description, version, instance_auth_request_json_schema,
					default_instance_auth, ord_id, local_tenant_id, short_description, links, labels, credential_exchange_strategies,
					ready, created_at, updated_at, deleted_at, error, correlation_ids, tags, resource_hash, documentation_labels FROM
					public.bundles WHERE app_id IN (
						SELECT DISTINCT pa.id as id
						FROM public.applications pa JOIN labels l ON pa.id=l.app_id
						WHERE pa.id IN (
							SELECT id
							FROM tenant_applications
							WHERE tenant_id=(SELECT parent FROM business_tenant_mappings WHERE id = $1 )
						)
						AND l.key='applicationType'
						AND l.value ?| array[$2]
						AND pa.local_tenant_id = $3
				) AND correlation_ids ?| array[$4]`),
				Args: []driver.Value{tenantID, destinationWithSystemID.XSystemType,
					destinationWithSystemID.XSystemTenantID, destinationWithSystemID.XCorrelationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).
						AddRow(fixBundleRowWithAppID(bundleID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundle.NewRepository,
		ExpectedModelEntities: []interface{}{modelBundle},
		ExpectedDBEntities:    []interface{}{bndlEntity},
		MethodArgs:            []interface{}{tenantID, destinationWithSystemID},
		MethodName:            "ListByDestination",
	}

	suiteBySystemID.Run(t)
}
