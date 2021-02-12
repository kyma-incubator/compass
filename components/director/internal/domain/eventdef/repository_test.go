package eventdef_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	// given
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	selectQuery := `^SELECT (.+) FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2$`

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
		modelApiDef, err := pgRepository.GetByID(ctx, tenantID, eventID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventID, modelApiDef.ID)
		assert.Equal(t, tenantID, modelApiDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

}

func TestPgRepository_GetForBundle(t *testing.T) {
	// given
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")

	selectQuery := `^SELECT (.+) FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2 AND bundle_id = \$3`

	t.Run("success", func(t *testing.T) {
		bundleID := bundleID
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(eventID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventID, bundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", eventDefEntity).Return(model.EventDefinition{Tenant: tenantID, BundleID: &bundleID, BaseEntity: &model.BaseEntity{ID: eventID}}, nil).Once()
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelApiDef, err := pgRepository.GetForBundle(ctx, tenantID, eventID, bundleID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventID, modelApiDef.ID)
		assert.Equal(t, tenantID, modelApiDef.Tenant)
		assert.Equal(t, &bundleID, modelApiDef.BundleID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := event.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, eventID, bundleID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelApiDef, err := repo.GetForBundle(ctx, tenantID, eventID, bundleID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelApiDef)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_ListForBundle(t *testing.T) {
	// GIVEN
	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityEventDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityEventDefinition(secondApiDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."event_api_definitions" 
		WHERE tenant_id = \$1 AND bundle_id = \$2
		ORDER BY id LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	rawCountQuery := `SELECT COUNT(*) FROM "public"."event_api_definitions" 
		WHERE tenant_id = $1 AND bundle_id = $2`
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixEventDefinitionColumns()).
			AddRow(fixEventDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixEventDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, bundleID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.EventDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}}, nil)
		pgRepository := event.NewRepository(convMock)
		// WHEN
		modelEventDef, err := pgRepository.ListForBundle(ctx, tenantID, bundleID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelEventDef.Data, 2)
		assert.Equal(t, firstApiDefID, modelEventDef.Data[0].ID)
		assert.Equal(t, secondApiDefID, modelEventDef.Data[1].ID)
		assert.Equal(t, "", modelEventDef.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelEventDef.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	eventDefModel, _ := fixFullEventDefinitionModel("placeholder")
	eventDefEntity := fixFullEntityEventDefinition(eventID, "placeholder")
	insertQuery := `^INSERT INTO "public"."event_api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixEventCreateArgs(eventID, &eventDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventDefModel).Return(eventDefEntity, nil).Once()
		pgRepository := event.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, &eventDefModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EventAPIDefinitionConverter{}
		pgRepository := event.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."event_api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first, _ := fixFullEventDefinitionModel("first")
		second, _ := fixFullEventDefinitionModel("second")
		third, _ := fixFullEventDefinitionModel("third")
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

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE "public"."event_api_definitions" SET bundle_id = ?, package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, changelog_entries = ?, links = ?, tags = ?, countries = ?, release_status = ?,
		sunset_date = ?, successor = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?, industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?,
		version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		eventModel, _ := fixFullEventDefinitionModel("update")
		entity := fixFullEntityEventDefinition(eventID, "update")
		entity.UpdatedAt = &fixedTimestamp
		entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

		convMock := &automock.EventAPIDefinitionConverter{}
		convMock.On("ToEntity", eventModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.BundleID, entity.PackageID, entity.Name, entity.Description, entity.GroupName, entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.ChangeLogEntries, entity.Links,
				entity.Tags, entity.Countries, entity.ReleaseStatus, entity.SunsetDate, entity.Successor, entity.Labels, entity.Visibility,
				entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := event.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &eventModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EventAPIDefinitionConverter{}
		pgRepository := event.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM "public"."event_api_definitions" WHERE tenant_id = \$1 AND id = \$2$`

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

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM "public"."event_api_definitions" WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, eventID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EventAPIDefinitionConverter{}
	pgRepository := event.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, eventID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
