package integrationdependency_test

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"regexp"
	"testing"
)

func TestPgRepository_GetByID(t *testing.T) {
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)

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
							AddRow(fixIntegrationDependenciesRow(integrationDependencyID)...),
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
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)

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
							AddRow(fixIntegrationDependenciesRow(integrationDependencyID)...),
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
	firstIntegrationDependencyEntity := fixIntegrationDependencyEntity(firstIntegrationDependencyID)
	secondIntegrationDependencyID := "222222222-2222-2222-2222-222222222222"
	secondIntegrationDependencyModel := fixIntegrationDependencyModel(secondIntegrationDependencyID)
	secondIntegrationDependencyEntity := fixIntegrationDependencyEntity(secondIntegrationDependencyID)

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List Integration Dependencies for AppID and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, sunset_date, successors, mandatory, related_integration_dependencies, links, tags, labels, documentation_labels, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.integration_dependencies WHERE app_id = $1 AND (id IN (SELECT id FROM integration_dependencies_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntegrationDependencyID)...).AddRow(fixIntegrationDependenciesRow(secondIntegrationDependencyID)...)}
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
					return []*sqlmock.Rows{sqlmock.NewRows(fixIntegrationDependenciesColumns()).AddRow(fixIntegrationDependenciesRow(firstIntegrationDependencyID)...).AddRow(fixIntegrationDependenciesRow(secondIntegrationDependencyID)...)}
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

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilIntegrationDependencyModel *model.IntegrationDependency
	integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)

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
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)

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
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)
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
	integrationDependencyEntity := fixIntegrationDependencyEntity(integrationDependencyID)

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
