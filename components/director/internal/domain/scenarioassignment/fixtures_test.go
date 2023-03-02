package scenarioassignment_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const (
	tenantID               = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	externalTargetTenantID = "extTargetTenantID"
	targetTenantID         = "targetTenantID"
	externalTenantID       = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	scenarioName           = "scenario-A"
	errMsg                 = "some error"
)

func fixModel() model.AutomaticScenarioAssignment {
	return fixModelWithScenarioName(scenarioName)
}

func fixGQL() graphql.AutomaticScenarioAssignment {
	return fixGQLWithScenarioName(scenarioName)
}

var testTableColumns = []string{"scenario", "tenant_id", "target_tenant_id"}

func fixModelWithScenarioName(scenario string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   scenario,
		Tenant:         tenantID,
		TargetTenantID: targetTenantID,
	}
}

func fixModelPageWithItems(in []*model.AutomaticScenarioAssignment) model.AutomaticScenarioAssignmentPage {
	return model.AutomaticScenarioAssignmentPage{
		Data:       in,
		PageInfo:   &pagination.Page{},
		TotalCount: len(in),
	}
}

func fixGQLWithScenarioName(scenario string) graphql.AutomaticScenarioAssignment {
	return graphql.AutomaticScenarioAssignment{
		ScenarioName: scenario,
		Selector: &graphql.Label{
			Key:   scenarioassignment.SubaccountIDKey,
			Value: externalTargetTenantID,
		},
	}
}

func fixGQLPageWithItems(in []*graphql.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignmentPage {
	return graphql.AutomaticScenarioAssignmentPage{
		Data:       in,
		PageInfo:   &graphql.PageInfo{},
		TotalCount: len(in),
	}
}

func fixEntity() scenarioassignment.Entity {
	return scenarioassignment.Entity{
		Scenario:       scenarioName,
		TenantID:       tenantID,
		TargetTenantID: targetTenantID,
	}
}

func fixEntityWithScenarioName(scenario string) scenarioassignment.Entity {
	return scenarioassignment.Entity{
		Scenario:       scenario,
		TenantID:       tenantID,
		TargetTenantID: targetTenantID,
	}
}

func fixError() error {
	return errors.New(errMsg)
}

func fixCtxWithTenant() context.Context {
	return tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
}

type sqlRow struct {
	scenario       string
	tenantID       string
	targetTenantID string
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.scenario, row.tenantID, row.targetTenantID)
	}
	return out
}

func fixAutomaticScenarioAssignmentRow(scenarioName, tenantID, targetTenantID string) []driver.Value {
	return []driver.Value{scenarioName, tenantID, targetTenantID}
}

func fixAutomaticScenarioAssignmentColumns() []string {
	return []string{"scenario", "tenant_id", "target_tenant_id"}
}
