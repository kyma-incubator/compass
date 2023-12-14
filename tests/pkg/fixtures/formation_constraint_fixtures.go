package fixtures

import (
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func FixFormationConstraintInputContainsScenarioGroups(resourceSubtype string, targetOperation graphql.TargetOperation, inputTemplate string) graphql.FormationConstraintInput {
	return graphql.FormationConstraintInput{
		Name:            "TestContainsScenarioGroupsAssign",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: targetOperation,
		Operator:        formationconstraintpkg.ContainsScenarioGroups,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
}
