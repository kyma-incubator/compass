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
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, app_id, app_template_version_id, ord_id, local_id, correlation_ids, level, title, short_description, description, system_instance_aware, changelog_entries, package_id, visibility, links, part_of_products, policy_level, custom_policy_level, release_status, sunset_date, successors, extensible, tags, labels, documentation_labels, resource_hash, version_value, version_deprecated, version_deprecated_since, version_for_removal FROM public.entity_types WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM entity_types_tenants WHERE tenant_id = $3))`),
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
