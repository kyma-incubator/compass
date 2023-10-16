package entitytype_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_GetForApplication(t *testing.T) {
	entityTypeModel := fixEntityTypeModel(ID)
	entityTypeEntity := fixEntityTypeEntity(ID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityType for Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, package_id, name, description, group_name, ord_id, local_tenant_id, short_description, system_instance_aware, policy_level, custom_policy_level, changelog_entries, links, tags, countries, release_status, sunset_date, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, implementation_standard, custom_implementation_standard, custom_implementation_standard_description, extensible, successors, resource_hash, documentation_labels, correlation_ids FROM "public"."event_api_definitions" WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{ID, appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeColumns()).
							AddRow(fixEntityTypeRow(ID)...),
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
		MethodArgs:                []interface{}{tenantID, ID, appID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByApplicationID",
	}

	suite.Run(t)
}
