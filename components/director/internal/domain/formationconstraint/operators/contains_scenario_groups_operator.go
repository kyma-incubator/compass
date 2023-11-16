package operators

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"
)

const (
	// ContainsScenarioGroupsOperator represents the ContainsScenarioGroups operator
	ContainsScenarioGroupsOperator = "ContainsScenarioGroups"
)

// NewContainsScenarioGroupsInput is input constructor for ContainsScenarioGroupsOperator operator. It returns empty OperatorInput.
func NewContainsScenarioGroupsInput() OperatorInput {
	return &formationconstraint.ContainsScenarioGroupsInput{}
}

// ContainsScenarioGroups is a constraint operator. It checks if the resource from the OperatorInput is already part of formation of the type that the operator is associated with
func (e *ConstraintEngine) ContainsScenarioGroups(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %s", ContainsScenarioGroupsOperator)

	i, ok := input.(*formationconstraint.ContainsScenarioGroupsInput)
	if !ok {
		return false, errors.New("Incompatible input")
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", ContainsScenarioGroupsOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

	if i.ResourceType != model.ApplicationResourceType {
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	hasCorrectScenarioGroups, err := e.hasCorrectScenarioGroups(ctx, i.ResourceID, i.Tenant, i.RequiredScenarioGroups)
	if err != nil {
		return false, err
	}

	return hasCorrectScenarioGroups, nil
}

func (e *ConstraintEngine) hasCorrectScenarioGroups(ctx context.Context, applicationID, tenant string, requiredScenarioGroups []string) (bool, error) {
	if len(requiredScenarioGroups) == 0 {
		return true, nil
	}

	application, err := e.applicationRepository.GetByID(ctx, tenant, applicationID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting application with ID %q", applicationID)
	}

	if application.Status == nil || string(application.Status.Condition) != string(graphql.ApplicationStatusConditionConnected) {
		return false, errors.Errorf("Application with ID %q is not in status %s", applicationID, graphql.ApplicationStatusConditionConnected)
	}

	auths, err := e.systemAuthSvc.ListForObject(ctx, pkgmodel.ApplicationReference, applicationID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting system auths for application with ID %q", applicationID)
	}

	for _, auth := range auths {
		if auth.Value == nil || auth.Value.OneTimeToken == nil || !auth.Value.OneTimeToken.Used {
			continue
		}
		if len(auth.Value.OneTimeToken.ScenarioGroups) == 0 {
			// If scenario groups are empty, this means that these are legacy tokens,
			// which should be interpreted as unrestricted
			return true, nil
		}
		scenarioGroups, err := onetimetoken.UnmarshalScenarioGroups(auth.Value.OneTimeToken.ScenarioGroups)
		if err != nil {
			for _, scenarioGroup := range auth.Value.OneTimeToken.ScenarioGroups {
				if slices.Contains(requiredScenarioGroups, scenarioGroup) {
					return true, nil
				}
			}
			continue
		}
		for _, scenarioGroup := range scenarioGroups {
			if slices.Contains(requiredScenarioGroups, scenarioGroup.Key) {
				return true, nil
			}
		}
	}

	return false, nil
}
