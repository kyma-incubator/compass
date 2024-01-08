package dataproduct_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct"
	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestPgRepository_ListByResourceID(t *testing.T) {
	firstDataProductID := "111111111-1111-1111-1111-111111111111"
	firstDataProductModel := fixDataProductModel(firstDataProductID)
	firstDataProductEntity := fixDataProductEntity(firstDataProductID, appID)
	secondDataProductID := "222222222-2222-2222-2222-222222222222"
	secondDataProductModel := fixDataProductModel(secondDataProductID)
	secondDataProductEntity := fixDataProductEntity(secondDataProductID, appID)

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List Data Products by AppID and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type, category, entity_types, input_ports, output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels, documentation_labels, policy_level, custom_policy_level, system_instance_aware, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.data_products WHERE app_id = $1 AND (id IN (SELECT id FROM data_products_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixDataProductColumns()).AddRow(fixDataProductRow(firstDataProductID, appID)...).AddRow(fixDataProductRow(secondDataProductID, appID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixDataProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ExpectedModelEntities:     []interface{}{firstDataProductModel, secondDataProductModel},
		ExpectedDBEntities:        []interface{}{firstDataProductEntity, secondDataProductEntity},
		MethodArgs:                []interface{}{tenantID, resource.Application, appID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Data Products by AppTemplateVersionID ",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type, category, entity_types, input_ports, output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels, documentation_labels, policy_level, custom_policy_level, system_instance_aware, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.data_products WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:  []driver.Value{appTemplateVersionID}, IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixDataProductColumns()).AddRow(fixDataProductRow(firstDataProductID, appID)...).AddRow(fixDataProductRow(secondDataProductID, appID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixDataProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ExpectedModelEntities:     []interface{}{firstDataProductModel, secondDataProductModel},
		ExpectedDBEntities:        []interface{}{firstDataProductEntity, secondDataProductEntity},
		MethodArgs:                []interface{}{tenantID, resource.ApplicationTemplateVersion, appTemplateVersionID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilDataProductModel *model.DataProduct
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Data Product",
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
				Query:       `^INSERT INTO public.data_products \(.+\) VALUES \(.+\)$`,
				Args:        fixDataProductCreateArgs(dataProductID, dataProductModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ModelEntity:               dataProductModel,
		DBEntity:                  dataProductEntity,
		NilModelEntity:            nilDataProductModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	var nilDataProductModel *model.DataProduct
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Data Product Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.data_products \(.+\) VALUES \(.+\)$`,
				Args:        fixDataProductCreateArgs(dataProductID, dataProductModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ModelEntity:               dataProductModel,
		DBEntity:                  dataProductEntity,
		NilModelEntity:            nilDataProductModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Data Product by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type, category, entity_types, input_ports, output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels, documentation_labels, policy_level, custom_policy_level, system_instance_aware, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.data_products WHERE id = $1 AND (id IN (SELECT id FROM data_products_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{dataProductID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixDataProductColumns()).
							AddRow(fixDataProductRow(dataProductID, appID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixDataProductColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ExpectedModelEntity:       dataProductModel,
		ExpectedDBEntity:          dataProductEntity,
		MethodArgs:                []interface{}{tenantID, dataProductID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Data Product Global by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, title, short_description, description, package_id, last_update, visibility, release_status, disabled, deprecation_date, sunset_date, successors, changelog_entries, type, category, entity_types, input_ports, output_ports, responsible, data_product_links, links, industry, line_of_business, tags, labels, documentation_labels, policy_level, custom_policy_level, system_instance_aware, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash FROM public.data_products WHERE id = $1`),
				Args:     []driver.Value{dataProductID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixDataProductColumns()).
							AddRow(fixDataProductRow(dataProductID, appID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixDataProductColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ExpectedModelEntity:       dataProductModel,
		ExpectedDBEntity:          dataProductEntity,
		MethodArgs:                []interface{}{dataProductID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilDataProductModel *model.DataProduct
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)
	dataProductEntity.UpdatedAt = &fixedTimestamp
	dataProductEntity.DeletedAt = &fixedTimestamp

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Data Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.data_products SET ord_id = ?, local_tenant_id = ?, correlation_ids = ?, title = ?, short_description = ?, description = ?, package_id = ?, last_update = ?, visibility = ?, release_status = ?, disabled = ?, deprecation_date = ?, sunset_date = ?, successors = ?, changelog_entries = ?, type = ?, category = ?, entity_types = ?, input_ports = ?, output_ports = ?, responsible = ?, data_product_links = ?, links = ?, industry = ?, line_of_business = ?, tags = ?, labels = ?, documentation_labels = ?, policy_level = ?, custom_policy_level = ?, system_instance_aware = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, resource_hash = ? WHERE id = ? AND (id IN (SELECT id FROM data_products_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixDataProductUpdateArgs(dataProductEntity), tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ModelEntity:               dataProductModel,
		DBEntity:                  dataProductEntity,
		NilModelEntity:            nilDataProductModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	var nilDataProductModel *model.DataProduct
	dataProductModel := fixDataProductModel(dataProductID)
	dataProductEntity := fixDataProductEntity(dataProductID, appID)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Data Product Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.data_products SET ord_id = ?, local_tenant_id = ?, correlation_ids = ?, title = ?, short_description = ?, description = ?, package_id = ?, last_update = ?, visibility = ?, release_status = ?, disabled = ?, deprecation_date = ?, sunset_date = ?, successors = ?, changelog_entries = ?, type = ?, category = ?, entity_types = ?, input_ports = ?, output_ports = ?, responsible = ?, data_product_links = ?, links = ?, industry = ?, line_of_business = ?, tags = ?, labels = ?, documentation_labels = ?, policy_level = ?, custom_policy_level = ?, system_instance_aware = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, resource_hash = ? WHERE id = ?`),
				Args:          fixDataProductUpdateArgs(dataProductEntity),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc:       dataproduct.NewRepository,
		ModelEntity:               dataProductModel,
		DBEntity:                  dataProductEntity,
		NilModelEntity:            nilDataProductModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		UpdateMethodName:          "UpdateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Data Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.data_products WHERE id = $1 AND (id IN (SELECT id FROM data_products_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{dataProductID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc: dataproduct.NewRepository,
		MethodArgs:          []interface{}{tenantID, dataProductID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Data Product Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.data_products WHERE id = $1`),
				Args:          []driver.Value{dataProductID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.DataProductConverter{}
		},
		RepoConstructorFunc: dataproduct.NewRepository,
		MethodArgs:          []interface{}{dataProductID},
		IsGlobal:            true,
		MethodName:          "DeleteGlobal",
	}

	suite.Run(t)
}
