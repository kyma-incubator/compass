package scenarioassignment_test

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tenantID     = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	scenarioName = "scenario-A"
	errMsg       = "some error"
)

func fixModel() model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   "key",
			Value: "value",
		},
	}
}

func fixEntity() scenarioassignment.Entity {
	return scenarioassignment.Entity{
		Scenario:      scenarioName,
		TenantID:      tenantID,
		SelectorKey:   "key",
		SelectorValue: "value",
	}
}

func fixError() error {
	return errors.New(errMsg)
}

func fixCtxWithTenant() context.Context {
	return tenant.SaveToContext(context.TODO(), tenantID)
}

func fixAutomaticScenarioAssignmentRow(scenarioName, tenantID string) []driver.Value {
	return []driver.Value{scenarioName, tenantID, "key", "value"}
}
