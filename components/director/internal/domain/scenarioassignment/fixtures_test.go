package scenarioassignment_test

import (
	"context"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

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

func fixEntity() scenarioassignment.Entity {
	return scenarioassignment.Entity{
		Scenario:      "scenario-A",
		TenantID:      "tenant",
		SelectorKey:   "key",
		SelectorValue: "value",
	}
}

func fixError() error {
	return errors.New("some error")
}

func fixCtxWithTenant() context.Context {
	return tenant.SaveToContext(context.TODO(), "tenant")
}
