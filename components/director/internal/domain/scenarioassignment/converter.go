package scenarioassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const SubaccountIDKey = "global_subaccount_id"

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// FromInputGraphQL missing godoc
func (c *converter) FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput, targetTenantInternalID string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   in.ScenarioName,
		TargetTenantID: targetTenantInternalID,
	}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in model.AutomaticScenarioAssignment, targetTenantExternalID string) graphql.AutomaticScenarioAssignment {
	return graphql.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
		Selector: &graphql.Label{
			Key:   SubaccountIDKey,
			Value: targetTenantExternalID,
		},
	}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in model.AutomaticScenarioAssignment) Entity {
	return Entity{
		TenantID:       in.Tenant,
		Scenario:       in.ScenarioName,
		TargetTenantID: in.TargetTenantID,
	}
}

// FromEntity missing godoc
func (c *converter) FromEntity(in Entity) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   in.Scenario,
		Tenant:         in.TenantID,
		TargetTenantID: in.TargetTenantID,
	}
}
