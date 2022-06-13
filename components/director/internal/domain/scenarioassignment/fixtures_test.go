package scenarioassignment_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"

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

func fixModelPage() model.AutomaticScenarioAssignmentPage {
	mod1 := fixModelWithScenarioName("foo")
	mod2 := fixModelWithScenarioName("bar")
	modItems := []*model.AutomaticScenarioAssignment{
		&mod1, &mod2,
	}
	return fixModelPageWithItems(modItems)
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

func fixGQLPage() graphql.AutomaticScenarioAssignmentPage {
	gql1 := fixGQLWithScenarioName("foo")
	gql2 := fixGQLWithScenarioName("bar")
	gqlItems := []*graphql.AutomaticScenarioAssignment{
		&gql1, &gql2,
	}
	return fixGQLPageWithItems(gqlItems)
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

func fixUpdateTenantIsolationSubquery() string {
	return regexp.QuoteMeta(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`)
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
