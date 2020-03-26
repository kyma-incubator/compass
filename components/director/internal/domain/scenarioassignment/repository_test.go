package scenarioassignment_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given

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

		// when
		err := repo.Create(ctx, fixModel())

		// then
		assert.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// given

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", fixModel()).Return(fixEntity()).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, fixModel())

		// then
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
		pgRepository := scenarioassignment.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetForScenarioName(ctx, tenantID, scenarioName)
		//THEN
		require.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, scenarioName).
			WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(nil)

		// when
		_, err := repo.GetForScenarioName(ctx, tenantID, scenarioName)

		// then
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", errMsg))
	})
}
