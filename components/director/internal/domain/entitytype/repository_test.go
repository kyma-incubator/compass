package entitytype_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestPgRepository_GetForApplication(t *testing.T) {
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityType for Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{entityTypeID, appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()).
							AddRow(fixEntityTypeRow(entityTypeID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ExpectedModelEntity:       entityTypeModel,
		ExpectedDBEntity:          entityTypeEntity,
		MethodArgs:                []interface{}{tenantID, entityTypeID, appID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByApplicationID",
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityType",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{entityTypeID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()).
							AddRow(fixEntityTypeRow(entityTypeID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ExpectedModelEntity:       entityTypeModel,
		ExpectedDBEntity:          entityTypeEntity,
		MethodArgs:                []interface{}{tenantID, entityTypeID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityType Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE id = $1`),
				Args:     []driver.Value{entityTypeID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()).
							AddRow(fixEntityTypeRow(entityTypeID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ExpectedModelEntity:       entityTypeModel,
		ExpectedDBEntity:          entityTypeEntity,
		MethodArgs:                []interface{}{entityTypeID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	firstEntityTypeID := "111111111-1111-1111-1111-111111111111"
	firstEntityTypeModel := fixEntityTypeModel(firstEntityTypeID)
	firstEntityTypeEntity := fixEntityTypeEntity(firstEntityTypeID)
	secondEntityTypeID := "222222222-2222-2222-2222-222222222222"
	secondEntityTypeModel := fixEntityTypeModel(secondEntityTypeID)
	secondEntityTypeEntity := fixEntityTypeEntity(secondEntityTypeID)

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List EntityTypes for AppID and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE app_id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeColumns()).AddRow(fixEntityTypeRow(firstEntityTypeID)...).AddRow(fixEntityTypeRow(secondEntityTypeID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ExpectedModelEntities:     []interface{}{firstEntityTypeModel, secondEntityTypeModel},
		ExpectedDBEntities:        []interface{}{firstEntityTypeEntity, secondEntityTypeEntity},
		MethodArgs:                []interface{}{tenantID, appID, resource.Application},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List EntityTypes for AppTemplateVersionID ",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeColumns()).AddRow(fixEntityTypeRow(firstEntityTypeID)...).AddRow(fixEntityTypeRow(secondEntityTypeID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ExpectedModelEntities:     []interface{}{firstEntityTypeModel, secondEntityTypeModel},
		ExpectedDBEntities:        []interface{}{firstEntityTypeEntity, secondEntityTypeEntity},
		MethodArgs:                []interface{}{tenantID, appTemplateVersionID, resource.ApplicationTemplateVersion},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_ListByApplicationIDPage(t *testing.T) {
	pageSize := 1
	cursor := ""

	firstEntityTypeID := "firstEntityTypeID"
	firstEntityTypeModel := fixEntityTypeModel(firstEntityTypeID)
	firstEntityTypeEntity := fixEntityTypeEntity(firstEntityTypeID)

	suite := testdb.RepoListPageableTestSuite{
		Name: "List EntityTypes with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_tenant_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, last_update, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE (app_id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2)))`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeColumns()).AddRow(fixEntityTypeRow(firstEntityTypeID)...)}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.entity_types WHERE (app_id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2)))`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{firstEntityTypeModel},
				ExpectedDBEntities:    []interface{}{firstEntityTypeEntity},
				ExpectedPage: &model.EntityTypePage{
					Data: []*model.EntityType{firstEntityTypeModel},
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
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		MethodName:                "ListByApplicationIDPage",
		MethodArgs:                []interface{}{tenantID, appID, pageSize, cursor},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilEntityTypeModel *model.EntityType
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create EntityTypes",
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
				Query:       `^INSERT INTO public.entity_types \(.+\) VALUES \(.+\)$`,
				Args:        fixEntityTypeCreateArgs(entityTypeID, entityTypeModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ModelEntity:               entityTypeModel,
		DBEntity:                  entityTypeEntity,
		NilModelEntity:            nilEntityTypeModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	var nilEntityTypeModel *model.EntityType
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create EntityTypes Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.entity_types \(.+\) VALUES \(.+\)$`,
				Args:        fixEntityTypeCreateArgs(entityTypeID, entityTypeModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ModelEntity:               entityTypeModel,
		DBEntity:                  entityTypeEntity,
		NilModelEntity:            nilEntityTypeModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.entity_types SET ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, ord_id = ?, local_tenant_id = ?, correlation_ids = ?, level = ?, title = ?, short_description = ?, description = ?, system_instance_aware = ?, changelog_entries = ?, package_id = ?, visibility = ?, links = ?, part_of_products = ?, last_update = ?, policy_level = ?, custom_policy_level = ?, release_status = ?, sunset_date = ?, successors = ?, extensible = ?, tags = ?, labels = ?, documentation_labels = ?, resource_hash = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ? WHERE id = ? AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = ? AND owner = true))`)
	var nilEntityTypeModel *model.EntityType
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)
	entityTypeEntity.UpdatedAt = &fixedTimestamp
	entityTypeEntity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update EntityTypes",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateQuery,
				Args:          append(fixEntityTypeUpdateArgs(entityTypeID, entityTypeEntity), tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ModelEntity:               entityTypeModel,
		DBEntity:                  entityTypeEntity,
		NilModelEntity:            nilEntityTypeModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "EntityType Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.entity_types WHERE id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{entityTypeID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc: entitytype.NewRepository,
		MethodArgs:          []interface{}{tenantID, entityTypeID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "EntityType Delete Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.entity_types WHERE id = $1`),
				Args:          []driver.Value{entityTypeID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc: entitytype.NewRepository,
		MethodArgs:          []interface{}{entityTypeID},
		IsGlobal:            true,
		MethodName:          "DeleteGlobal",
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "EntityType Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.entity_types WHERE id = $1 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{entityTypeID, tenantID},
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
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc: entitytype.NewRepository,
		TargetID:            entityTypeID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, entityTypeID},
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	var nilEntityTypeModel *model.EntityType
	entityTypeModel := fixEntityTypeModel(entityTypeID)
	entityTypeEntity := fixEntityTypeEntity(entityTypeID)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update EntityType Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.entity_types SET ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, ord_id = ?, local_tenant_id = ?, correlation_ids = ?, level = ?, title = ?, short_description = ?, description = ?, system_instance_aware = ?, changelog_entries = ?, package_id = ?, visibility = ?, links = ?, part_of_products = ?, last_update = ?, policy_level = ?, custom_policy_level = ?, release_status = ?, sunset_date = ?, successors = ?, extensible = ?, tags = ?, labels = ?, documentation_labels = ?, resource_hash = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ? WHERE id = ?`),
				Args:          fixEntityTypeUpdateArgs(entityTypeID, entityTypeEntity),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeConverter{}
		},
		RepoConstructorFunc:       entitytype.NewRepository,
		ModelEntity:               entityTypeModel,
		DBEntity:                  entityTypeEntity,
		NilModelEntity:            nilEntityTypeModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		UpdateMethodName:          "UpdateGlobal",
	}

	suite.Run(t)
}
