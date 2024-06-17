package operators

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// DoesNotContainResourceOfSubtypeOperator represents the DoesNotContainResourceOfSubtype operator
	DoesNotContainResourceOfSubtypeOperator = "DoesNotContainResourceOfSubtype"
)

// NewDoesNotContainResourceOfSubtypeInput is input constructor for DoesNotContainResourceOfSubtypeOperator operator. It returns empty OperatorInput
func NewDoesNotContainResourceOfSubtypeInput() OperatorInput {
	return &formationconstraint.DoesNotContainResourceOfSubtypeInput{}
}

// DoesNotContainResourceOfSubtype is a constraint operator. It checks if the formation contains resource with the same subtype as the resource subtype from the OperatorInput
func (e *ConstraintEngine) DoesNotContainResourceOfSubtype(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DoesNotContainResourceOfSubtypeOperator)

	i, ok := input.(*formationconstraint.DoesNotContainResourceOfSubtypeInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator %q", DoesNotContainResourceOfSubtypeOperator)
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", DoesNotContainResourceOfSubtypeOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

	switch i.ResourceType {
	case model.ApplicationResourceType:
		applications, err := e.applicationRepository.ListByScenariosNoPaging(ctx, i.Tenant, []string{i.FormationName})
		if err != nil {
			return false, errors.Wrapf(err, "while listing applications in scenario %q", i.FormationName)
		}

		for _, application := range applications {
			appSubtype, err := e.getObjectSubtype(ctx, i.Tenant, model.ApplicationResourceType, application.ID)
			if err != nil {
				return false, errors.Wrapf(err, "while getting subtype of resource with type: %q and id: %q", model.ApplicationResourceType, application.ID)
			}

			if i.ResourceSubtype == appSubtype {
				return false, nil
			}
		}
	default:
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	return true, nil
}
