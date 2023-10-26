package capability_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability"
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestPgRepository_ListByResourceID(t *testing.T) {
	entity1App := fixFullEntityCapabilityWithAppID(capabilityID, "name")
	capabilityModel1App, _ := fixFullCapabilityModelWithAppID("name")
	entity2App := fixFullEntityCapabilityWithAppID(capabilityID, "name2")
	capabilityModel2App, _ := fixFullCapabilityModelWithAppID("name2")
	entity1AppTemplateVersion := fixFullEntityCapabilityWithAppTemplateVersionID(capabilityID, "name")
	capabilityModel1AppTemplateVersion, _ := fixFullCapabilityModelWithAppTemplateVersionID("name")
	entity2AppTemplateVersion := fixFullEntityCapabilityWithAppTemplateVersionID(capabilityID, "name2")
	capabilityModel2AppTemplateVersion, _ := fixFullCapabilityModelWithAppTemplateVersionID("name2")

	suiteForApplication := testdb.RepoListTestSuite{
		Name: "List Capabilities",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, ord_id, type, custom_type, local_tenant_id, short_description, system_instance_aware, tags, links, release_status, labels, visibility, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash, documentation_labels, correlation_ids, last_update FROM "public"."capabilities" WHERE app_id = $1 AND (id IN (SELECT id FROM capabilities_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns()).AddRow(fixCapabilityRow(capabilityID, "name")...).AddRow(fixCapabilityRow(capabilityID, "name2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ExpectedModelEntities:     []interface{}{&capabilityModel1App, &capabilityModel2App},
		ExpectedDBEntities:        []interface{}{&entity1App, &entity2App},
		MethodArgs:                []interface{}{tenantID, resource.Application, appID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplicationTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Capabilities",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, ord_id, type, custom_type, local_tenant_id, short_description, system_instance_aware, tags, links, release_status, labels, visibility, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash, documentation_labels, correlation_ids, last_update FROM "public"."capabilities" WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns()).AddRow(fixCapabilityRowForAppTemplateVersion(capabilityID, "name")...).AddRow(fixCapabilityRowForAppTemplateVersion(capabilityID, "name2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ExpectedModelEntities:     []interface{}{&capabilityModel1AppTemplateVersion, &capabilityModel2AppTemplateVersion},
		ExpectedDBEntities:        []interface{}{&entity1AppTemplateVersion, &entity2AppTemplateVersion},
		MethodArgs:                []interface{}{tenantID, resource.ApplicationTemplateVersion, appTemplateVersionID},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForApplication.Run(t)
	suiteForApplicationTemplateVersion.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	entity := fixFullEntityCapabilityWithAppID(capabilityID, "name")
	capabilityModel, _ := fixFullCapabilityModelWithAppID("name")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Capability",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, ord_id, type, custom_type, local_tenant_id, short_description, system_instance_aware, tags, links, release_status, labels, visibility, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash, documentation_labels, correlation_ids, last_update FROM "public"."capabilities" WHERE id = $1 AND (id IN (SELECT id FROM capabilities_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{capabilityID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns()).AddRow(fixCapabilityRow(capabilityID, "name")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ExpectedModelEntity:       &capabilityModel,
		ExpectedDBEntity:          &entity,
		MethodArgs:                []interface{}{tenantID, capabilityID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	entity := fixFullEntityCapabilityWithAppID(capabilityID, "name")
	capabilityModel, _ := fixFullCapabilityModelWithAppID("name")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Capability Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, ord_id, type, custom_type, local_tenant_id, short_description, system_instance_aware, tags, links, release_status, labels, visibility, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, resource_hash, documentation_labels, correlation_ids, last_update FROM "public"."capabilities" WHERE id = $1`),
				Args:     []driver.Value{capabilityID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns()).AddRow(fixCapabilityRow(capabilityID, "name")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixCapabilityColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ExpectedModelEntity:       &capabilityModel,
		ExpectedDBEntity:          &entity,
		MethodName:                "GetByIDGlobal",
		MethodArgs:                []interface{}{capabilityID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	var nilCapabilityModel *model.Capability
	capabilityModel, _ := fixFullCapabilityModelWithAppID("name")
	capabilityEntity := fixFullEntityCapabilityWithAppID(capabilityID, "name")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Capability",
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
				Query:       `^INSERT INTO "public"."capabilities" \(.+\) VALUES \(.+\)$`,
				Args:        fixCapabilityCreateArgs(capabilityID, &capabilityModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ModelEntity:               &capabilityModel,
		DBEntity:                  &capabilityEntity,
		NilModelEntity:            nilCapabilityModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	var nilCapabilityModel *model.Capability
	capabilityModel, _ := fixFullCapabilityModelWithAppID("name")
	capabilityEntity := fixFullEntityCapabilityWithAppID(capabilityID, "name")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Capability Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO "public"."capabilities" \(.+\) VALUES \(.+\)$`,
				Args:        fixCapabilityCreateArgs(capabilityID, &capabilityModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ModelEntity:               &capabilityModel,
		DBEntity:                  &capabilityEntity,
		NilModelEntity:            nilCapabilityModel,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."capabilities" SET package_id = ?, name = ?, description = ?, ord_id = ?, type = ?, custom_type = ?, local_tenant_id = ?,
		short_description = ?, system_instance_aware = ?, tags = ?, links = ?, release_status = ?, labels = ?, visibility = ?,
		version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?,
		updated_at = ?, deleted_at = ?, error = ?, resource_hash = ?, documentation_labels = ?, correlation_ids = ?, last_update = ?
		WHERE id = ? AND (id IN (SELECT id FROM capabilities_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilCapabilityModel *model.Capability
	capabilityModel, _ := fixFullCapabilityModelWithAppID("update")
	entity := fixFullEntityCapabilityWithAppID(capabilityID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Capability",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description,
					entity.OrdID, entity.Type, entity.CustomType, entity.LocalTenantID, entity.ShortDescription, entity.SystemInstanceAware, entity.Tags,
					entity.Links, entity.ReleaseStatus, entity.Labels, entity.Visibility,
					entity.Version.Value, entity.Version.Deprecated,
					entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt,
					entity.Error, entity.ResourceHash, entity.DocumentationLabels, entity.CorrelationIDs, entity.LastUpdate, entity.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ModelEntity:               &capabilityModel,
		DBEntity:                  &entity,
		NilModelEntity:            nilCapabilityModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."capabilities" SET package_id = ?, name = ?, description = ?, ord_id = ?, type = ?, custom_type = ?, local_tenant_id = ?,
		short_description = ?, system_instance_aware = ?, tags = ?, links = ?, release_status = ?, labels = ?, visibility = ?,
		version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?,
		updated_at = ?, deleted_at = ?, error = ?, resource_hash = ?, documentation_labels = ?, correlation_ids = ?, last_update = ?
		WHERE id = ?`)

	var nilCapabilityModel *model.Capability
	capabilityModel, _ := fixFullCapabilityModelWithAppID("update")
	entity := fixFullEntityCapabilityWithAppID(capabilityID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Capability Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description,
					entity.OrdID, entity.Type, entity.CustomType, entity.LocalTenantID, entity.ShortDescription, entity.SystemInstanceAware, entity.Tags,
					entity.Links, entity.ReleaseStatus, entity.Labels, entity.Visibility,
					entity.Version.Value, entity.Version.Deprecated,
					entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt,
					entity.Error, entity.ResourceHash, entity.DocumentationLabels, entity.CorrelationIDs, entity.LastUpdate, entity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc:       capability.NewRepository,
		ModelEntity:               &capabilityModel,
		DBEntity:                  &entity,
		NilModelEntity:            nilCapabilityModel,
		UpdateMethodName:          "UpdateGlobal",
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Capability",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."capabilities" WHERE id = $1 AND (id IN (SELECT id FROM capabilities_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{capabilityID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc: capability.NewRepository,
		MethodArgs:          []interface{}{tenantID, capabilityID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Capability Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."capabilities" WHERE id = $1`),
				Args:          []driver.Value{capabilityID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.CapabilityConverter{}
		},
		RepoConstructorFunc: capability.NewRepository,
		MethodArgs:          []interface{}{capabilityID},
		MethodName:          "DeleteGlobal",
		IsGlobal:            true,
	}

	suite.Run(t)
}
