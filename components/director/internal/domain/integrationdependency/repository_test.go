package integrationdependency_test

import (
	"context"
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestPgRepository_GetByID(t *testing.T) {
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Integration Dependency",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id = $1 AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{integrationDependencyID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixIntegrationDependenciesColumns()).
							AddRow(fixIntegrationDependenciesRow(integrationDependencyID, appID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixIntegrationDependenciesColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ExpectedModelEntity:       integrationDependencyModel,
		ExpectedDBEntity:          integrationDependencyEntity,
		MethodArgs:                []interface{}{tenantID, integrationDependencyID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Integration Dependency Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id = $1`),
				Args:     []driver.Value{integrationDependencyID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixIntegrationDependenciesColumns()).
							AddRow(fixIntegrationDependenciesRow(integrationDependencyID, appID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixIntegrationDependenciesColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ExpectedModelEntity:       integrationDependencyModel,
		ExpectedDBEntity:          integrationDependencyEntity,
		MethodArgs:                []interface{}{integrationDependencyID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	firstIntegrationDependencyID := "111111111-1111-1111-1111-111111111111"
	firstIntegrationDependencyModel := fixIntegrationDependencyModel(firstIntegrationDependencyID)
	firstIntegrationDependencyEntity := fixIntegrationDependencyEntity(firstIntegrationDependencyID, appID)
	secondIntegrationDependencyID := "222222222-2222-2222-2222-222222222222"
	secondIntegrationDependencyModel := fixIntegrationDependencyModel(secondIntegrationDependencyID)
	secondIntegrationDependencyEntity := fixIntegrationDependencyEntity(secondIntegrationDependencyID, appID)

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List Integration Dependencies for AppID and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE app_id = $1 AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntegrationDependencyID, appID)...).AddRow(fixIntegrationDependenciesRow(secondIntegrationDependencyID, appID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ExpectedModelEntities:     []interface{}{firstIntegrationDependencyModel, secondIntegrationDependencyModel},
		ExpectedDBEntities:        []interface{}{firstIntegrationDependencyEntity, secondIntegrationDependencyEntity},
		MethodArgs:                []interface{}{tenantID, resource.Application, appID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Integration Dependencies for AppTemplateVersionID ",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntegrationDependencyID, appID)...).AddRow(fixIntegrationDependenciesRow(secondIntegrationDependencyID, appID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ExpectedModelEntities:     []interface{}{firstIntegrationDependencyModel, secondIntegrationDependencyModel},
		ExpectedDBEntities:        []interface{}{firstIntegrationDependencyEntity, secondIntegrationDependencyEntity},
		MethodArgs:                []interface{}{tenantID, resource.ApplicationTemplateVersion, appTemplateVersionID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForPackage := testdb.RepoListTestSuite{
		Name: "List Integration Dependencies for packageID and tenantID ",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE package_id = $1 AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{packageID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntegrationDependencyID, appID)...).AddRow(fixIntegrationDependenciesRow(secondIntegrationDependencyID, appID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ExpectedModelEntities:     []interface{}{firstIntegrationDependencyModel, secondIntegrationDependencyModel},
		ExpectedDBEntities:        []interface{}{firstIntegrationDependencyEntity, secondIntegrationDependencyEntity},
		MethodArgs:                []interface{}{tenantID, resource.Package, packageID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
	suiteForPackage.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilIntegrationDependencyModel *model.IntegrationDependency
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Integration Dependency",
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
				Query:       `^INSERT INTO public.integration_dependencies \(.+\) VALUES \(.+\)$`,
				Args:        fixIntegrationDependencyCreateArgs(integrationDependencyID, integrationDependencyModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ModelEntity:               integrationDependencyModel,
		DBEntity:                  integrationDependencyEntity,
		NilModelEntity:            nilIntegrationDependencyModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	var nilIntegrationDependencyModel *model.IntegrationDependency
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Integration Dependency Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.integration_dependencies \(.+\) VALUES \(.+\)$`,
				Args:        fixIntegrationDependencyCreateArgs(integrationDependencyID, integrationDependencyModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ModelEntity:               integrationDependencyModel,
		DBEntity:                  integrationDependencyEntity,
		NilModelEntity:            nilIntegrationDependencyModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilIntegrationDependencyModel *model.IntegrationDependency
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)
	integrationDependencyEntity.UpdatedAt = &fixedTimestamp
	integrationDependencyEntity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Integration Dependency",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.integration_dependencies SET ord_id = ?, local_tenant_id = ?, correlation_ids = ?, title = ?, short_description = ?, description  = ?, package_id = ?, last_update = ?, visibility = ?,  release_status = ?, sunset_date = ?, successors = ?, mandatory = ?, related_integration_dependencies = ?, links = ?, tags = ?, labels = ?, documentation_labels = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, resource_hash = ? WHERE id = ? AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixIntegrationDependencyUpdateArgs(integrationDependencyEntity), tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ModelEntity:               integrationDependencyModel,
		DBEntity:                  integrationDependencyEntity,
		NilModelEntity:            nilIntegrationDependencyModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	var nilIntegrationDependencyModel *model.IntegrationDependency
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID, appID)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Integration Dependency Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.integration_dependencies SET ord_id = ?, local_tenant_id = ?, correlation_ids = ?, title = ?, short_description = ?, description  = ?, package_id = ?, last_update = ?, visibility = ?,  release_status = ?, sunset_date = ?, successors = ?, mandatory = ?, related_integration_dependencies = ?, links = ?, tags = ?, labels = ?, documentation_labels = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, resource_hash = ? WHERE id = ?`),
				Args:          fixIntegrationDependencyUpdateArgs(integrationDependencyEntity),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		ModelEntity:               integrationDependencyModel,
		DBEntity:                  integrationDependencyEntity,
		NilModelEntity:            nilIntegrationDependencyModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		UpdateMethodName:          "UpdateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Integration Dependency Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.integration_dependencies WHERE id = $1 AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{integrationDependencyID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc: integrationdependency.NewRepository,
		MethodArgs:          []interface{}{tenantID, integrationDependencyID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Integration Dependency Delete Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.integration_dependencies WHERE id = $1`),
				Args:          []driver.Value{integrationDependencyID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc: integrationdependency.NewRepository,
		MethodArgs:          []interface{}{integrationDependencyID},
		IsGlobal:            true,
		MethodName:          "DeleteGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageAppID := "emptyPageAppID"

	onePageAppID := "onePageAppID"
	firstIntDepID := "111111111-1111-1111-1111-111111111111"
	firstIntDepEntity := fixIntegrationDependencyEntity(firstIntDepID, "foo")
	firstIntDepEntity.ApplicationID = repo.NewValidNullableString(onePageAppID)
	firstIntDepModel := fixIntegrationDependencyModel(firstIntDepID)
	firstIntDepModel.ApplicationID = &onePageAppID

	multiplePagesAppID := "multiplePagesAppID"

	secondIntDepID := "222222222-2222-2222-2222-222222222222"
	secondIntDepEntity := fixIntegrationDependencyEntity(secondIntDepID, "foo")
	secondIntDepEntity.ApplicationID = repo.NewValidNullableString(multiplePagesAppID)
	secondIntDepModel := fixIntegrationDependencyModel(secondIntDepID)
	secondIntDepModel.ApplicationID = &multiplePagesAppID

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Integration Dependencies for multiple Applications with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`(SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $1)) AND app_id = $2 ORDER BY app_id ASC, id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $5)) AND app_id = $6 ORDER BY app_id ASC, id ASC LIMIT $7 OFFSET $8) UNION (SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $9)) AND app_id = $10 ORDER BY app_id ASC, id ASC LIMIT $11 OFFSET $12`),
				Args:     []driver.Value{tenantID, emptyPageAppID, pageSize, 0, tenantID, onePageAppID, pageSize, 0, tenantID, multiplePagesAppID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntDepID, onePageAppID)...).AddRow(fixIntegrationDependenciesRow(secondIntDepID, multiplePagesAppID)...)}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT app_id AS id, COUNT(*) AS total_count FROM public.integration_dependencies WHERE id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $1)) GROUP BY app_id ORDER BY app_id ASC`),
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
				ExpectedPage: &model.IntegrationDependencyPage{
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
				ExpectedModelEntities: []interface{}{firstIntDepModel},
				ExpectedDBEntities:    []interface{}{firstIntDepEntity},
				ExpectedPage: &model.IntegrationDependencyPage{
					Data: []*model.IntegrationDependency{firstIntDepModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{secondIntDepModel},
				ExpectedDBEntities:    []interface{}{secondIntDepEntity},
				ExpectedPage: &model.IntegrationDependencyPage{
					Data: []*model.IntegrationDependency{secondIntDepModel},
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
			return &automock.IntegrationDependencyConverter{}
		},
		RepoConstructorFunc:       integrationdependency.NewRepository,
		MethodName:                "ListByApplicationIDs",
		MethodArgs:                []interface{}{tenantID, []string{emptyPageAppID, onePageAppID, multiplePagesAppID}, pageSize, cursor},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	t.Run("ListByApplicationIDs when there is missing visibility scope", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixIntegrationDependenciesColumns()).
			AddRow(fixIntegrationDependenciesRow(firstIntDepID, onePageAppID)...).
			AddRow(fixIntegrationDependenciesRow(secondIntDepID, multiplePagesAppID)...)
		query := regexp.QuoteMeta(`(SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IN (SELECT id FROM public.integration_dependencies WHERE visibility = $1) AND id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2)) AND app_id = $3 ORDER BY app_id ASC, id ASC LIMIT $4 OFFSET $5) UNION (SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IN (SELECT id FROM public.integration_dependencies WHERE visibility = $6) AND id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $7)) AND app_id = $8 ORDER BY app_id ASC, id ASC LIMIT $9 OFFSET $10) UNION (SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE id IN (SELECT id FROM public.integration_dependencies WHERE visibility = $11) AND id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $12)) AND app_id = $13 ORDER BY app_id ASC, id ASC LIMIT $14 OFFSET $15)`)
		sqlMock.ExpectQuery(query).
			WithArgs(publicVisibility, tenantID, emptyPageAppID, pageSize, 0, publicVisibility, tenantID, onePageAppID, pageSize, 0, publicVisibility, tenantID, multiplePagesAppID, pageSize, 0).
			WillReturnRows(rows)

		countQuery := regexp.QuoteMeta(`SELECT app_id AS id, COUNT(*) AS total_count FROM public.integration_dependencies WHERE id IN (SELECT id FROM public.integration_dependencies WHERE visibility = $1) AND id IS NOT NULL AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2)) GROUP BY app_id ORDER BY app_id ASC`)
		sqlMock.ExpectQuery(countQuery).
			WithArgs(publicVisibility, tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(emptyPageAppID, 0).
				AddRow(onePageAppID, 1).
				AddRow(multiplePagesAppID, 2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		ctx = scope.SaveToContext(ctx, []string{"test:test"})

		convMock := &automock.IntegrationDependencyConverter{}
		convMock.On("FromEntity", firstIntDepEntity).Return(&model.IntegrationDependency{
			ApplicationID: str.Ptr(onePageAppID),
		}, nil)
		convMock.On("FromEntity", secondIntDepEntity).Return(&model.IntegrationDependency{
			ApplicationID: str.Ptr(multiplePagesAppID),
		}, nil)
		pgRepository := integrationdependency.NewRepository(convMock)
		// WHEN
		intDepPage, err := pgRepository.ListByApplicationIDs(ctx, tenantID, []string{emptyPageAppID, onePageAppID, multiplePagesAppID}, pageSize, cursor)
		// THEN
		require.NoError(t, err)
		require.Len(t, intDepPage, 3)
		require.Len(t, intDepPage[0].Data, 0)
		require.Len(t, intDepPage[1].Data, 1)
		assert.Equal(t, *intDepPage[1].Data[0].ApplicationID, onePageAppID)
		require.Len(t, intDepPage[2].Data, 1)
		assert.Equal(t, *intDepPage[2].Data[0].ApplicationID, multiplePagesAppID)

		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
