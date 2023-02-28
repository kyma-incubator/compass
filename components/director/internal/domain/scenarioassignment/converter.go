package scenarioassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// SubaccountIDKey is the key used for the subaccount_id label
const SubaccountIDKey = "global_subaccount_id"

// NewConverter creates a new instance of gqlConverter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in model.AutomaticScenarioAssignment, targetTenantExternalID string) graphql.AutomaticScenarioAssignment {
	return graphql.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
		Selector: &graphql.Label{
			Key:   SubaccountIDKey,
			Value: targetTenantExternalID,
		},
	}
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in model.AutomaticScenarioAssignment) Entity {
	return Entity{
		TenantID:       in.Tenant,
		Scenario:       in.ScenarioName,
		TargetTenantID: in.TargetTenantID,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(in Entity) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName:   in.Scenario,
		Tenant:         in.TenantID,
		TargetTenantID: in.TargetTenantID,
	}
}
