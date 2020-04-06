package scenarioassignment_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

const (
	key   = "KEY"
	value = "VALUE"
)

func TestEngine_EnsureScenarioAssigned(t *testing.T) {
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

		eng := scenarioassignment.NewEngine()
		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, input, tenantID)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Error no persistance", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		eng := scenarioassignment.NewEngine()

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, input, tenantID)
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

		eng := scenarioassignment.NewEngine()
		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, input, tenantID)

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
