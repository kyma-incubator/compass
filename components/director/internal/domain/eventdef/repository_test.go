package eventdef_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_GetByID(t *testing.T) {
	eventDefModel := fixEventDefinitionModel(eventID, "placeholder")
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Document",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, changelog_entries, links, tags, countries, release_status, sunset_date, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, extensible, successors, resource_hash, documentation_labels FROM "public"."event_api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{eventID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEventDefinitionColumns()).
							AddRow(fixEventDefinitionRow(eventID, "placeholder")...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEventDefinitionColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc:       event.NewRepository,
		ExpectedModelEntity:       eventDefModel,
		ExpectedDBEntity:          eventDefEntity,
		MethodArgs:                []interface{}{tenantID, eventID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	firstEventDefID := "111111111-1111-1111-1111-111111111111"
	firstEventDefModel := fixEventDefinitionModel(firstEventDefID, "placeholder")
	firstEventDefEntity := fixFullEntityEventDefinition(firstEventDefID, "placeholder")
	secondEventDefID := "222222222-2222-2222-2222-222222222222"
	secondEventDefModel := fixEventDefinitionModel(secondEventDefID, "placeholder")
	secondEventDefEntity := fixFullEntityEventDefinition(secondEventDefID, "placeholder")

	suite := testdb.RepoListTestSuite{
		Name: "List APIs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, changelog_entries, links, tags, countries, release_status, sunset_date, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, extensible, successors, resource_hash, documentation_labels FROM "public"."event_api_definitions" WHERE app_id = $1 AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEventDefinitionColumns()).AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEventDefinitionColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc:       event.NewRepository,
		ExpectedModelEntities:     []interface{}{firstEventDefModel, secondEventDefModel},
		ExpectedDBEntities:        []interface{}{firstEventDefEntity, secondEventDefEntity},
		MethodArgs:                []interface{}{tenantID, appID},
		MethodName:                "ListByApplicationID",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListAllForBundle(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageBundleID := "emptyPageBundleID"

	onePageBundleID := "onePageBundleID"
	firstEventDefID := "firstEventDefID"
	firstEventDef, _, _ := fixFullEventDefinitionModelWithID(firstEventDefID, "placeholder")
	firstEntity := fixFullEntityEventDefinition(firstEventDefID, "placeholder")
	firstBundleRef := fixModelBundleReference(onePageBundleID, firstEventDefID)

	multiplePagesBundleID := "multiplePagesBundleID"

	secondEventDefID := "secondEventDefID"
	secondEventDef, _, _ := fixFullEventDefinitionModelWithID(secondEventDefID, "placeholder")
	secondEntity := fixFullEntityEventDefinition(secondEventDefID, "placeholder")
	secondBundleRef := fixModelBundleReference(multiplePagesBundleID, secondEventDefID)

	totalCounts := map[string]int{
		emptyPageBundleID:     0,
		onePageBundleID:       1,
		multiplePagesBundleID: 2,
	}

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Events for multiple bundles with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, package_id, name, description, group_name, ord_id, short_description, system_instance_aware, changelog_entries, links, tags, countries, release_status, sunset_date, labels, visibility, disabled, part_of_products, line_of_business, industry, version_value, version_deprecated, version_deprecated_since, version_for_removal, ready, created_at, updated_at, deleted_at, error, extensible, successors, resource_hash, documentation_labels FROM "public"."event_api_definitions" WHERE id IN ($1, $2) AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{firstEventDefID, secondEventDefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEventDefinitionColumns()).AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.EventDefinitionPage{
					Data: []*model.EventDefinition{},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 0,
				},
			},
			{
				ExpectedModelEntities: []interface{}{&firstEventDef},
				ExpectedDBEntities:    []interface{}{firstEntity},
				ExpectedPage: &model.EventDefinitionPage{
					Data: []*model.EventDefinition{&firstEventDef},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{&secondEventDef},
				ExpectedDBEntities:    []interface{}{secondEntity},
				ExpectedPage: &model.EventDefinitionPage{
					Data: []*model.EventDefinition{&secondEventDef},
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
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc: event.NewRepository,
		MethodName:          "ListByBundleIDs",
		MethodArgs: []interface{}{tenantID, []string{emptyPageBundleID, onePageBundleID, multiplePagesBundleID},
			[]*model.BundleReference{firstBundleRef, secondBundleRef}, totalCounts, pageSize, cursor},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilEventDefModel *model.EventDefinition
	eventDefModel, _, _ := fixFullEventDefinitionModel("placeholder")
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Event",
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
				Query:       `^INSERT INTO "public"."event_api_definitions" \(.+\) VALUES \(.+\)$`,
				Args:        fixEventCreateArgs(eventID, &eventDefModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc:       event.NewRepository,
		ModelEntity:               &eventDefModel,
		DBEntity:                  eventDefEntity,
		NilModelEntity:            nilEventDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."event_api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first, _, _ := fixFullEventDefinitionModel("first")
		second, _, _ := fixFullEventDefinitionModel("second")
		third, _, _ := fixFullEventDefinitionModel("third")
		items := []*model.EventDefinition{&first, &second, &third}

		convMock := &automock.EventAPIDefinitionConverter{}
		for _, item := range items {
			convMock.On("ToEntity", item).Return(fixFullEntityEventDefinition(item.ID, item.Name), nil).Once()
			sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3")).WithArgs(tenantID, appID, true).WillReturnRows(testdb.RowWhenObjectExist())
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixEventCreateArgs(item.ID, item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := event.NewRepository(convMock)
		// WHEN
		err := pgRepository.CreateMany(ctx, tenantID, items)
		// THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."event_api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, changelog_entries = ?, links = ?, tags = ?, countries = ?, release_status = ?,
		sunset_date = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?, industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?,
		version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, extensible = ?, successors = ?, resource_hash = ?, documentation_labels = ? WHERE id = ? AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilEventDefModel *model.EventDefinition
	eventModel, _, _ := fixFullEventDefinitionModel("update")
	entity := fixFullEntityEventDefinition(eventID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Event",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.GroupName, entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.ChangeLogEntries, entity.Links,
					entity.Tags, entity.Countries, entity.ReleaseStatus, entity.SunsetDate, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval,
					entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.Extensible, entity.Successors, entity.ResourceHash, entity.DocumentationLabels, entity.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc:       event.NewRepository,
		ModelEntity:               &eventModel,
		DBEntity:                  entity,
		NilModelEntity:            nilEventDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Event Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."event_api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{eventID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc: event.NewRepository,
		MethodArgs:          []interface{}{tenantID, eventID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteAllByBundleID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Event Delete By BundleID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM "public"."event_api_definitions" WHERE id IN (SELECT event_def_id FROM public.bundle_references WHERE bundle_id = $1 AND event_def_id IS NOT NULL) AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{bundleID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc: event.NewRepository,
		MethodName:          "DeleteAllByBundleID",
		MethodArgs:          []interface{}{tenantID, bundleID},
		IsDeleteMany:        true,
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Event Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM "public"."event_api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{eventID, tenantID},
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
			return &automock.EventAPIDefinitionConverter{}
		},
		RepoConstructorFunc: event.NewRepository,
		TargetID:            eventID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, eventID},
	}

	suite.Run(t)
}
