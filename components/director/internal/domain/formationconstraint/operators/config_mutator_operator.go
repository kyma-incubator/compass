package operators

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
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
		return false, errors.New("Incompatible input")
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", i.ResourceType, i.ResourceSubtype, i.Location.ConstraintType, i.Location.OperationName, i.Operation)

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, i.JoinPointDetailsFAMemoryAddress)
	if err != nil {
		return false, err
	}

	fmt.Println(">>>>>>>>> Assignment STATE is: ", formationAssignment.State)
	if i.State != nil {
		fmt.Println(">>>>>>>>> SETTING STATE TO: ", *i.State)
		formationAssignment.State = *i.State
	}

	if i.Configuration != nil {
		formationAssignment.Value = json.RawMessage(*i.Configuration)
	}

	spew.Dump("FORMATION ASSIGNMENT UPDATED:::::: ", formationAssignment)
	return true, nil
}
