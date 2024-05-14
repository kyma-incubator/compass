package formationtemplate_test

import (
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Template By ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE id = $1`),
				Args:     []driver.Value{testFormationTemplateID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntity.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ExpectedModelEntity:       &formationTemplateModel,
		ExpectedDBEntity:          &formationTemplateEntity,
		MethodArgs:                []interface{}{testFormationTemplateID},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_GetByNameAndTenant(t *testing.T) {
	suiteWithTenant := testdb.RepoGetTestSuite{
		Name:       "Get Formation Template By name and tenant when tenant is present",
		MethodName: "GetByNameAndTenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE tenant_id = $1 AND name = $2`),
				Args:     []driver.Value{testTenantID, formationTemplateName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntity.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ExpectedModelEntity:       &formationTemplateModel,
		ExpectedDBEntity:          &formationTemplateEntity,
		MethodArgs:                []interface{}{formationTemplateName, testTenantID},
		DisableConverterErrorTest: false,
		ExpectNotFoundError:       true,
		AfterNotFoundErrorSQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE name = $1 AND tenant_id IS NULL`),
				Args:     []driver.Value{formationTemplateName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntityNullTenant.TenantID)}
				},
			},
		},
		AfterNotFoundErrorExpectedDBEntity:    &formationTemplateEntityNullTenant,
		AfterNotFoundErrorExpectedModelEntity: &formationTemplateModelNullTenant,
	}

	suiteWithoutTenant := testdb.RepoGetTestSuite{
		Name:       "Get Formation Template By name and tenant when tenant is not present",
		MethodName: "GetByNameAndTenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE name = $1 AND tenant_id IS NULL`),
				Args:     []driver.Value{formationTemplateName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntityNullTenant.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ExpectedModelEntity:       &formationTemplateModelNullTenant,
		ExpectedDBEntity:          &formationTemplateEntityNullTenant,
		MethodArgs:                []interface{}{formationTemplateName, ""},
		DisableConverterErrorTest: false,
	}

	suiteWithoutTenant.Run(t)
	suiteWithTenant.Run(t)
}

func TestRepository_Create(t *testing.T) {
	formationtemplate.Now = func() time.Time { return testTime }

	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation Template",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formation_templates \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntity.TenantID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		ModelEntity:               &formationTemplateModel,
		DBEntity:                  &formationTemplateEntity,
		NilModelEntity:            nilModelEntity,
		IsGlobal:                  true,
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists Formation Template By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.formation_templates WHERE id = $1`),
				Args:     []driver.Value{testFormationTemplateID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   testFormationTemplateID,
		IsGlobal:   true,
		MethodName: "ExistsGlobal",
		MethodArgs: []interface{}{testFormationTemplateID},
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suiteWithoutTenant := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation Template By ID when there is no tenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_templates WHERE id = $1 AND tenant_id IS NULL`),
				Args:          []driver.Value{testFormationTemplateID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodArgs: []interface{}{testFormationTemplateID, ""},
	}

	suiteWithTenant := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation Template By ID when there is tenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formation_templates WHERE tenant_id = $1 AND id = $2`),
				Args:          []driver.Value{testTenantID, testFormationTemplateID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		MethodArgs: []interface{}{testFormationTemplateID, testTenantID},
	}

	suiteWithTenant.Run(t)
	suiteWithoutTenant.Run(t)
}

func TestRepository_List(t *testing.T) {
	suiteWithEmptyTenantID := testdb.RepoListPageableTestSuite{
		Name:       "List Formation Templates with paging when there is no tenant",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE (tenant_id IS NULL AND id IN (SELECT "formation_template_id" FROM public.labels WHERE "formation_template_id" IS NOT NULL AND tenant_id IS NULL AND "key" = $1))  ORDER BY id LIMIT 3 OFFSET 0`),
				Args:     []driver.Value{testLabelKey},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntityNullTenant.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formation_templates`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{&formationTemplateModelNullTenant},
				ExpectedDBEntities:    []interface{}{&formationTemplateEntityNullTenant},
				ExpectedPage: &model.FormationTemplatePage{
					Data: []*model.FormationTemplate{&formationTemplateModelNullTenant},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		MethodArgs:                []interface{}{testLabelFilter, nil, "", 3, ""},
		DisableConverterErrorTest: false,
	}

	suiteWithTenantID := testdb.RepoListPageableTestSuite{
		Name:       "List Formation Templates with paging when tenant is passed",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE (tenant_id IS NULL OR tenant_id = $1) ORDER BY id LIMIT 3 OFFSET 0`),
				Args:     []driver.Value{testTenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntity.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formation_templates`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{&formationTemplateModel},
				ExpectedDBEntities:    []interface{}{&formationTemplateEntity},
				ExpectedPage: &model.FormationTemplatePage{
					Data: []*model.FormationTemplate{&formationTemplateModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		MethodArgs:                []interface{}{nil, nil, testTenantID, 3, ""},
		DisableConverterErrorTest: false,
	}

	suiteWithTenantIDandName := testdb.RepoListPageableTestSuite{
		Name:       "List Formation Templates with paging when tenant is passed",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, application_types, runtime_types, runtime_type_display_name, runtime_artifact_kind, leading_product_ids, supports_reset, discovery_consumers, created_at, updated_at, tenant_id FROM public.formation_templates WHERE ((tenant_id IS NULL OR tenant_id = $1) AND name = $2) ORDER BY id LIMIT 3 OFFSET 0`),
				Args:     []driver.Value{testTenantID, formationTemplateEntity.Name},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(formationTemplateEntity.ID, formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.CreatedAt, formationTemplateEntity.UpdatedAt, formationTemplateEntity.TenantID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formation_templates`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{&formationTemplateModel},
				ExpectedDBEntities:    []interface{}{&formationTemplateEntity},
				ExpectedPage: &model.FormationTemplatePage{
					Data: []*model.FormationTemplate{&formationTemplateModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formationtemplate.NewRepository,
		MethodArgs:                []interface{}{nil, &formationTemplateEntity.Name, testTenantID, 3, ""},
		DisableConverterErrorTest: false,
	}

	suiteWithEmptyTenantID.Run(t)
	suiteWithTenantIDandName.Run(t)
	suiteWithTenantID.Run(t)
}

func TestRepository_Update(t *testing.T) {
	formationtemplate.Now = func() time.Time { return testTime }

	updateStmtWithoutTenant := regexp.QuoteMeta(`UPDATE public.formation_templates SET name = ?, application_types = ?, runtime_types = ?, runtime_type_display_name = ?, runtime_artifact_kind = ?, leading_product_ids = ?, supports_reset = ?, discovery_consumers = ?, updated_at = ? WHERE id = ?`)
	suiteWithoutTenant := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Template By ID without tenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmtWithoutTenant,
				Args:          []driver.Value{formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.UpdatedAt, formationTemplateEntity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ModelEntity:    &formationTemplateModelNullTenant,
		DBEntity:       &formationTemplateEntityNullTenant,
		NilModelEntity: nilModelEntity,
		IsGlobal:       true,
	}

	updateStmtWithTenant := regexp.QuoteMeta(`UPDATE public.formation_templates SET name = ?, application_types = ?, runtime_types = ?, runtime_type_display_name = ?, runtime_artifact_kind = ?, leading_product_ids = ?, supports_reset = ?, discovery_consumers = ?, updated_at = ? WHERE id = ? AND tenant_id = ?`)
	suiteWithTenant := testdb.RepoUpdateTestSuite{
		Name: "Update Formation Template By ID with tenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmtWithTenant,
				Args:          []driver.Value{formationTemplateEntity.Name, formationTemplateEntity.ApplicationTypes, formationTemplateEntity.RuntimeTypes, formationTemplateEntity.RuntimeTypeDisplayName, formationTemplateEntity.RuntimeArtifactKind, formationTemplateEntity.LeadingProductIDs, formationTemplateEntity.SupportsReset, formationTemplateEntity.DiscoveryConsumers, formationTemplateEntity.UpdatedAt, formationTemplateEntity.ID, formationTemplateEntity.TenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: formationtemplate.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ModelEntity:    &formationTemplateModel,
		DBEntity:       &formationTemplateEntity,
		NilModelEntity: nilModelEntity,
	}

	suiteWithTenant.Run(t)
	suiteWithoutTenant.Run(t)
}
