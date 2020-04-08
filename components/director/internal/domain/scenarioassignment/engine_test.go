package scenarioassignment_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/require"
)

func TestEngine_EnsureScenarioAssigned(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	input := model.AutomaticScenarioAssignment{
		ScenarioName: "SCENARIO",
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   "KEY",
			Value: "VALUE",
		},
	}
	asaRepo := &automock.Repository{}
	asaRepo.On("EnsureScenarioAssigned", ctx, input).Return(nil)
	eng := scenarioassignment.NewEngine(asaRepo)

	//WHEN
	err := eng.EnsureScenarioAssigned(ctx, input)

	//THEN
	require.NoError(t, err)
	asaRepo.AssertExpectations(t)
}
