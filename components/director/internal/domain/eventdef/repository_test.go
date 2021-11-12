package eventdef_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)
/*
func TestPgRepository_GetByID(t *testing.T) {
	// given
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."event_api_definitions" WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventDefEntity).Return(model.EventDefinition{Tenant: tenantID, BaseEntity: &model.BaseEntity{ID: eventID}}, nil).Once()
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.GetByID(ctx, tenantID, eventID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventID, modelAPIDef.ID)
		assert.Equal(t, tenantID, modelAPIDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstEventDefID := "111111111-1111-1111-1111-111111111111"
	firstEventDefEntity := fixFullEntityEventDefinition(firstEventDefID, "placeholder")
	secondEventDefID := "222222222-2222-2222-2222-222222222222"
	secondEventDefEntity := fixFullEntityEventDefinition(secondEventDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."event_api_definitions" 
		WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: firstEventDefID}}, nil)
		convMock.On("FromEntity", secondEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: secondEventDefID}}, nil)
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDef, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDef, totalCount)
		assert.Equal(t, firstEventDefID, modelEventDef[0].ID)
		assert.Equal(t, secondEventDefID, modelEventDef[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListAllForBundle(t *testing.T) {
	// GIVEN
	inputPageSize := 3
	inputCursor := ""

	firstBndlID := "111111111-1111-1111-1111-111111111111"
	secondBndlID := "222222222-2222-2222-2222-222222222222"
	bundleIDs := []string{firstBndlID, secondBndlID}

	firstEventDefID := "111111111-1111-1111-1111-111111111111"
	firstEventDefEntity := fixFullEntityEventDefinition(firstEventDefID, "placeholder")
	secondEventDefID := "222222222-2222-2222-2222-222222222222"
	secondEventDefEntity := fixFullEntityEventDefinition(secondEventDefID, "placeholder")

	firstBundleRef := fixModelBundleReference(firstBndlID, firstEventDefID)
	secondBundleRef := fixModelBundleReference(secondBndlID, secondEventDefID)
	bundleRefs := []*model.BundleReference{firstBundleRef, secondBundleRef}

	totalCounts := map[string]int{firstBndlID: 1, secondBndlID: 1}

	selectQuery := fmt.Sprintf(`^SELECT (.+) 
		FROM "public"."event_api_definitions" 
		WHERE %s AND id IN \(\$2, \$3\)`, fixTenantIsolationSubquery())

	t.Run("success when there are no more pages", func(t *testing.T) {
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstEventDefID, secondEventDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: firstEventDefID}})
		convMock.On("FromEntity", secondEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: secondEventDefID}})
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDefs, 2)
		assert.Equal(t, firstEventDefID, modelEventDefs[0].Data[0].ID)
		assert.Equal(t, secondEventDefID, modelEventDefs[1].Data[0].ID)
		assert.Equal(t, "", modelEventDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelEventDefs[0].TotalCount)
		assert.False(t, modelEventDefs[0].PageInfo.HasNextPage)
		assert.Equal(t, "", modelEventDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelEventDefs[1].TotalCount)
		assert.False(t, modelEventDefs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page", func(t *testing.T) {
		totalCountForFirstBundle := 10
		totalCountForSecondBundle := 10
		totalCounts[firstBndlID] = 10
		totalCounts[secondBndlID] = 10

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstEventDefID, secondEventDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: firstEventDefID}})
		convMock.On("FromEntity", secondEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: secondEventDefID}})
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDefs, 2)
		assert.Equal(t, firstEventDefID, modelEventDefs[0].Data[0].ID)
		assert.Equal(t, secondEventDefID, modelEventDefs[1].Data[0].ID)
		assert.Equal(t, "", modelEventDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelEventDefs[0].TotalCount)
		assert.True(t, modelEventDefs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelEventDefs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelEventDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelEventDefs[1].TotalCount)
		assert.True(t, modelEventDefs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelEventDefs[1].PageInfo.EndCursor)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page and it can be traversed", func(t *testing.T) {
		totalCountForFirstBundle := 2
		totalCountForSecondBundle := 2
		totalCounts[firstBndlID] = 2
		totalCounts[secondBndlID] = 2

		thirdEventDefID := "333333333-3333-3333-3333-333333333333"
		thirdEventDefEntity := fixFullEntityEventDefinition(thirdEventDefID, "placeholder")
		fourthEventDefID := "444444444-4444-4444-4444-444444444444"
		fourthEventDefEntity := fixFullEntityEventDefinition(fourthEventDefID, "placeholder")

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rowsFirstPage := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstEventDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondEventDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstEventDefID, secondEventDefID).
			WillReturnRows(rowsFirstPage)

		rowsSecondPage := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(thirdEventDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(fourthEventDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, thirdEventDefID, fourthEventDefID).
			WillReturnRows(rowsSecondPage)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: firstEventDefID}})
		convMock.On("FromEntity", secondEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: secondEventDefID}})
		convMock.On("FromEntity", thirdEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: thirdEventDefID}})
		convMock.On("FromEntity", fourthEventDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: fourthEventDefID}})
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDefs, 2)
		assert.Equal(t, firstEventDefID, modelEventDefs[0].Data[0].ID)
		assert.Equal(t, secondEventDefID, modelEventDefs[1].Data[0].ID)
		assert.Equal(t, "", modelEventDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelEventDefs[0].TotalCount)
		assert.True(t, modelEventDefs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelEventDefs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelEventDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelEventDefs[1].TotalCount)
		assert.True(t, modelEventDefs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelEventDefs[1].PageInfo.EndCursor)
		endCursor := modelEventDefs[0].PageInfo.EndCursor

		thirdBundleRef := fixModelBundleReference(firstBndlID, thirdEventDefID)
		fourthBundleRef := fixModelBundleReference(secondBndlID, fourthEventDefID)
		bundleRefsSecondPage := []*model.BundleReference{thirdBundleRef, fourthBundleRef}

		modelEventDefsSecondPage, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefsSecondPage, totalCounts, inputPageSize, endCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDefsSecondPage, 2)
		assert.Equal(t, thirdEventDefID, modelEventDefsSecondPage[0].Data[0].ID)
		assert.Equal(t, fourthEventDefID, modelEventDefsSecondPage[1].Data[0].ID)
		assert.Equal(t, totalCountForFirstBundle, modelEventDefsSecondPage[0].TotalCount)
		assert.False(t, modelEventDefsSecondPage[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelEventDefsSecondPage[1].TotalCount)
		assert.False(t, modelEventDefsSecondPage[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns empty page", func(t *testing.T) {
		totalCountForFirstBundle := 0
		totalCountForSecondBundle := 0
		totalCounts[firstBndlID] = 0
		totalCounts[secondBndlID] = 0

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns())

		sqlMock.ExpectQuery(selectQuery).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN

		require.NoError(t, err)
		require.Len(t, modelEventDefs[0].Data, 0)
		require.Len(t, modelEventDefs[1].Data, 0)
		assert.Equal(t, totalCountForFirstBundle, modelEventDefs[0].TotalCount)
		assert.False(t, modelEventDefs[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelEventDefs[1].TotalCount)
		assert.False(t, modelEventDefs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		pgRepository := event.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstEventDefID, secondEventDefID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelEventDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelEventDefs)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
*/
func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	var nilEventDefModel *model.EventDefinition
	eventDefModel, _, _ := fixFullEventDefinitionModel("placeholder")
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Event",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
/*
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
			convMock.On("ToEntity", *item).Return(fixFullEntityEventDefinition(item.ID, item.Name), nil).Once()
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixEventCreateArgs(item.ID, item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := event.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, items)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/


func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE "public"."event_api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, changelog_entries = ?, links = ?, tags = ?, countries = ?, release_status = ?,
		sunset_date = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?, industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?,
		version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, extensible = ?, successors = ?, resource_hash = ? WHERE id = ? AND (id IN (SELECT id FROM event_api_definitions_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID))

	var nilEventDefModel *model.EventDefinition
	eventModel, _, _ := fixFullEventDefinitionModel("update")
	entity := fixFullEntityEventDefinition(eventID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Event",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.GroupName, entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.ChangeLogEntries, entity.Links,
					entity.Tags, entity.Countries, entity.ReleaseStatus, entity.SunsetDate, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.Extensible, entity.Successors, entity.ResourceHash, entity.ID},
				ValidResult: sqlmock.NewResult(-1, 1),
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

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM "public"."event_api_definitions" WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, eventID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := event.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, eventID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_DeleteAllByBundleID(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`DELETE FROM "public"."event_api_definitions"
		WHERE %s AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE %s AND bundle_id = \$3 AND event_def_id IS NOT NULL\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(2))

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, tenantID, bundleID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := event.NewRepository(convMock)
	//WHEN
	err := pgRepository.DeleteAllByBundleID(ctx, tenantID, bundleID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
*/

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name:                "Event Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
	}

	suite.Run(t)
}