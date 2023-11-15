package operators

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// ConfigMutatorOperator represents the ConfigMutator operator
	ConfigMutatorOperator = "ConfigMutator"
)

// NewConfigMutatorInput is input constructor for ConfigMutatorOperator operator. It returns empty OperatorInput.
func NewConfigMutatorInput() OperatorInput {
	return &formationconstraint.ConfigMutatorInput{}
}

// MutateConfig is a constraint operator. It mutates the Formation assignment state and configuration based on the provided input
func (e *ConstraintEngine) MutateConfig(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %s", ConfigMutatorOperator)

	i, ok := input.(*formationconstraint.ConfigMutatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %s", ConfigMutatorOperator)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", i.ResourceType, i.ResourceSubtype, i.Location.ConstraintType, i.Location.OperationName, i.Operation)

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, i.FAMemoryAddress)
	if err != nil {
		return false, err
	}

	if len(i.OnlyForSourceSubtypes) != 0 {
		sourceSubType, err := e.getObjectSubtype(ctx, i.Tenant, model.ResourceType(formationAssignment.SourceType), formationAssignment.Source)
		if err != nil {
			return false, errors.Wrapf(err, "while getting subtype of resource with type: %q and id: %q", formationAssignment.SourceType, formationAssignment.Source)
		}

		sourceSubtypeIsSupported := false
		for _, subtype := range i.OnlyForSourceSubtypes {
			if sourceSubType == subtype {
				sourceSubtypeIsSupported = true
				break
			}
		}
		if !sourceSubtypeIsSupported {
			log.C(ctx).Infof("Skipping configuration and state mutation of assignment with ID: %q source resource of type: %q, subtype: %q, ID: %q, and target resource of type: %q and ID: %q", formationAssignment.ID, formationAssignment.SourceType, sourceSubType, formationAssignment.Source, formationAssignment.TargetType, formationAssignment.Target)
			return true, nil
		}
	}

	notificationStatusReport, err := RetrieveNotificationStatusReportPointer(ctx, i.NotificationStatusReportMemoryAddress)
	if err != nil {
		return false, err
	}
	if i.State != nil {
		log.C(ctx).Infof("Updating state in notification status report from: %s, to: %s", notificationStatusReport.State, *i.State)
		notificationStatusReport.State = *i.State
	}

	if i.ModifiedConfiguration != nil {
		log.C(ctx).Infof("Updating configuration in notification status report")
		notificationStatusReport.Configuration = json.RawMessage(*i.ModifiedConfiguration)
	}

	return true, nil
}
