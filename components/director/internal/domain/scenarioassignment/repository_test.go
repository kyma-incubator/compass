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
		require.EqualError(t, err, fmt.Sprintf("while inserting row to 'public.automatic_scenario_assignments' table: %s", errMsg))
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
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", errMsg))
	})
}

func TestRepository_GetForSelector(t *testing.T) {
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
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT scenario, tenant_id, selector_key, selector_value FROM public.automatic_scenario_assignments WHERE tenant_id=$1 AND selector_key = 'key' AND selector_value = 'value'`)).
			WithArgs(tenantID).
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
		require.EqualError(t, err, "while getting automatic scenario assignments from db: while fetching list of objects from DB: some error")
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
		WHERE tenant_id=\$1
		ORDER BY scenario LIMIT %d OFFSET %d`, ExpectedLimit, ExpectedOffset)

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.automatic_scenario_assignments
		WHERE tenant_id=$1`)
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
		require.EqualError(t, err, fmt.Sprintf("while deleting from database: %s", errMsg))
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
		require.EqualError(t, err, "while deleting from database: some error")
	})
}

func TestRepository_EnsureScenarioAssigned(t *testing.T) {
	key := "KEY"
	value := "VALUE"

	input := model.AutomaticScenarioAssignment{
		ScenarioName: "TEST",
		Tenant:       tenantID,
		Selector:     model.LabelSelector{Key: key, Value: value}}

	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		mock := mockUpdateQuery(dbMock, key, value)
		mock.WillReturnResult(sqlmock.NewResult(-1, 3))
		ctx := persistence.SaveToContext(context.TODO(), db)

		input := model.AutomaticScenarioAssignment{
			ScenarioName: "TEST",
			Tenant:       tenantID,
			Selector:     model.LabelSelector{Key: key, Value: value}}

		repo := scenarioassignment.NewRepository(nil)
		//WHEN
		err := repo.EnsureScenarioAssigned(ctx, input)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Error no persistance", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		repo := scenarioassignment.NewRepository(nil)

		//WHEN
		err := repo.EnsureScenarioAssigned(ctx, input)
		//THEN
		require.Error(t, err)
		assert.Error(t, err, "while getting persitance from context: unable to fetch database from context")
	})

	t.Run("Error while executing query", func(t *testing.T) {

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		mock := mockUpdateQuery(dbMock, key, value)
		mock.WillReturnError(errors.New("test error"))
		ctx := persistence.SaveToContext(context.TODO(), db)

		input := model.AutomaticScenarioAssignment{
			ScenarioName: "TEST",
			Tenant:       tenantID,
			Selector:     model.LabelSelector{Key: key, Value: value}}

		repo := scenarioassignment.NewRepository(nil)
		//WHEN
		err := repo.EnsureScenarioAssigned(ctx, input)

		//THEN
		require.Error(t, err)
		assert.Error(t, err, "while updating scenarios: test error")
	})
}

func mockUpdateQuery(dbMock testdb.DBMock, key, value string) *sqlmock.ExpectedExec {
	return dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE labels AS l SET value=SCENARIOS.SCENARIOS 
		FROM (SELECT array_to_json(array_agg(scenario)) AS SCENARIOS FROM automatic_scenario_assignments 
					WHERE selector_key=$1 AND selector_value=$2 AND tenant_id=$3) AS SCENARIOS
		WHERE l.runtime_id IN (SELECT runtime_id FROM labels  
									WHERE key =$1 AND value ?| array[$2] AND runtime_id IS NOT NULL AND tenant_ID=$3) 
			AND l.key ='scenarios'
			AND l.tenant_id=$3`)).
		WithArgs(key, value, tenantID)
}
