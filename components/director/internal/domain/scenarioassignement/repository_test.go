package scenarioassignement_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignement"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignement/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
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
			WithArgs("scenario-A","tenant","key","value").
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignement.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, fixModel())

		// then
		assert.NoError(t, err)
	})
}

func fixModel() model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Tenant:       "tenant",
		Selector: model.LabelSelector{
			Key:   "key",
			Value: "value",
		},
	}
}

func fixEntity() scenarioassignement.Entity {
	return scenarioassignement.Entity{
		Scenario:"scenario-A",
		TenantID:"tenant",
		SelectorKey:"key",
		SelectorValue:"value",
	}
}
