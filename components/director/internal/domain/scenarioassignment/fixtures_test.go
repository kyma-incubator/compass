package scenarioassignment_test

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"

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

var testTableColumns = []string{"scenario", "tenant_id", "selector_key", "selector_value"}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName: scenario,
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

func fixEntityWithScenarioName(scenario string) scenarioassignment.Entity {
	return scenarioassignment.Entity{
		Scenario:      scenario,
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

func fixLabelSelector() model.LabelSelector {
	return model.LabelSelector{
		Key:   "key",
		Value: "value",
	}
}

type sqlRow struct {
	scenario      string
	tenantId      string
	selectorKey   string
	selectorValue string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.scenario, row.tenantId, row.selectorKey, row.selectorValue)
	}
	return out
}

func fixAutomaticScenarioAssignmentRow(scenarioName, tenantID string) []driver.Value {
	return []driver.Value{scenarioName, tenantID, "key", "value"}
}

func fixAutomaticScenarioAssignmentColumns() []string {
	return []string{"scenario", "tenant_id", "selector_key", "selector_value"}
}
