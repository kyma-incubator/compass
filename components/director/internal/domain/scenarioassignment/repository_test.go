package scenarioassignment_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", fixModel()).Return(fixEntity()).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.automatic_scenario_assignments ( scenario, tenant_id, selector_key, selector_value ) VALUES ( ?, ?, ?, ? )`)).
			WithArgs(scenarioName, tenantID, "key", "value").
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// WHEN
		err := repo.Create(ctx, fixModel())

		// THEN
		assert.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// GIVEN

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", fixModel()).Return(fixEntity()).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// WHEN
		err := repo.Create(ctx, fixModel())

		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_GetByScenarioName(t *testing.T) {
	ent := scenarioassignment.Entity{
		Scenario:      scenarioName,
		TenantID:      tenantID,
		SelectorKey:   "key",
		SelectorValue: "value",
	}

	selectQuery := `SELECT scenario, tenant_id, selector_key, selector_value FROM public.automatic_scenario_assignments WHERE tenant_id = \$1 AND scenario = \$2`

	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rows := sqlmock.NewRows(fixAutomaticScenarioAssignmentColumns()).
			AddRow(fixAutomaticScenarioAssignmentRow(scenarioName, tenantID)...)

		dbMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, scenarioName).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", ent).Return(fixModel()).Once()
		defer convMock.AssertExpectations(t)
		repo := scenarioassignment.NewRepository(convMock)
		// WHEN
		_, err := repo.GetForScenarioName(ctx, tenantID, scenarioName)
		//THEN
		require.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, scenarioName).
			WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		_, err := repo.GetForScenarioName(ctx, tenantID, scenarioName)

		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_ListForSelector(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		scenarioEntities := []scenarioassignment.Entity{fixEntityWithScenarioName(scenarioName),
			fixEntityWithScenarioName("scenario-B")}
		scenarioModels := []model.AutomaticScenarioAssignment{fixModelWithScenarioName(scenarioName),
			fixModelWithScenarioName("scenario-B")}

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", scenarioEntities[0]).Return(scenarioModels[0]).Once()
		mockConverter.On("FromEntity", scenarioEntities[1]).Return(scenarioModels[1]).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{scenario: scenarioName, tenantId: tenantID, selectorKey: "key", selectorValue: "value"},
			{scenario: "scenario-B", tenantId: tenantID, selectorKey: "key", selectorValue: "value"},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT scenario, tenant_id, selector_key, selector_value FROM public.automatic_scenario_assignments WHERE tenant_id = $1 AND selector_key = $2 AND selector_value = $3`)).
			WithArgs(tenantID, "key", "value").
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// WHEN
		result, err := repo.ListForSelector(ctx, fixLabelSelector(), tenantID)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, scenarioModels[0], *result[0])
		assert.Equal(t, scenarioModels[1], *result[1])
	})

	t.Run("DB error", func(t *testing.T) {
		// GIVEN

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		result, err := repo.ListForSelector(ctx, fixLabelSelector(), tenantID)

		// THEN
		require.EqualError(t, err, "while getting automatic scenario assignments from db: Internal Server Error: Unexpected error while executing SQL query")
		assert.Nil(t, result)
	})
}

func TestRepository_List(t *testing.T) {
	// GIVEN
	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2

	scenarioName1 := "foo"
	scenarioName2 := "bar"

	ent1 := fixEntityWithScenarioName(scenarioName1)
	ent2 := fixEntityWithScenarioName(scenarioName2)

	mod1 := fixModelWithScenarioName(scenarioName1)
	mod2 := fixModelWithScenarioName(scenarioName2)

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.automatic_scenario_assignments
		WHERE tenant_id = \$1
		ORDER BY scenario LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.automatic_scenario_assignments
		WHERE tenant_id = $1`)
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("Success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAutomaticScenarioAssignmentColumns()).
			AddRow(fixAutomaticScenarioAssignmentRow(scenarioName1, tenantID)...).
			AddRow(fixAutomaticScenarioAssignmentRow(scenarioName2, tenantID)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", ent1).Return(mod1)
		convMock.On("FromEntity", ent2).Return(mod2)
		repo := scenarioassignment.NewRepository(convMock)
		// WHEN
		modelAssignment, err := repo.List(ctx, tenantID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAssignment.Data, 2)
		assert.Equal(t, scenarioName1, modelAssignment.Data[0].ScenarioName)
		assert.Equal(t, scenarioName2, modelAssignment.Data[1].ScenarioName)
		assert.Equal(t, "", modelAssignment.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelAssignment.TotalCount)

		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		repo := scenarioassignment.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// WHEN
		modelAssignment, err := repo.List(ctx, tenantID, inputPageSize, inputCursor)

		// THEN
		sqlMock.AssertExpectations(t)

		assert.Nil(t, modelAssignment)
		require.EqualError(t, err, fmt.Sprintf("while fetching list of objects from DB: %s", testError.Error()))
	})
}

func TestRepository_DeleteForSelector(t *testing.T) {
	deleteQuery := regexp.QuoteMeta(`DELETE FROM public.automatic_scenario_assignments WHERE tenant_id = $1 AND selector_key = $2 AND selector_value = $3`)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(deleteQuery).
			WithArgs(tenantID, "key", "value").
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		err := repo.DeleteForSelector(ctx, tenantID, fixLabelSelector())

		// THEN
		require.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(deleteQuery).
			WithArgs(tenantID, "key", "value").
			WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		err := repo.DeleteForSelector(ctx, tenantID, fixLabelSelector())

		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_DeleteForScenarioName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(fmt.Sprintf(`^DELETE FROM public.automatic_scenario_assignments WHERE tenant_id = \$1 AND scenario = \$2$`)).
			WithArgs(tenantID, scenarioName).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		err := repo.DeleteForScenarioName(ctx, tenantID, scenarioName)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(fmt.Sprintf(`^DELETE FROM public.automatic_scenario_assignments WHERE tenant_id = \$1 AND scenario = \$2$`)).
			WithArgs(tenantID, scenarioName).
			WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// WHEN
		err := repo.DeleteForScenarioName(ctx, tenantID, scenarioName)

		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
