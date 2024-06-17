package operators

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// JSONSchemaValidatorOperator represents the JSONSchemaValidator operator
	JSONSchemaValidatorOperator = "JSONSchemaValidator"
)

// NewJSONSchemaValidatorOperatorInput is input constructor for JSONSchemaValidatorOperator operator. It returns empty OperatorInput
func NewJSONSchemaValidatorOperatorInput() OperatorInput {
	return &formationconstraint.JSONSchemaValidatorOperatorInput{}
}

// JSONSchemaValidator is an operator that validates the configuration of a formation assignment against a JSON schema
func (e *ConstraintEngine) JSONSchemaValidator(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Starting executing operator: %s", JSONSchemaValidatorOperator)

	i, ok := input.(*formationconstraint.JSONSchemaValidatorOperatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %s", JSONSchemaValidatorOperator)
	}
	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %s, subtype: %s and ID: %s. And source resource type: %s with ID: %s", JSONSchemaValidatorOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID, i.SourceResourceType, i.SourceResourceID)

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, i.FAMemoryAddress)
	if err != nil {
		return false, err
	}

	isAssignmentInitialCfgEmpty := isFormationAssignmentConfigEmpty(formationAssignment)
	if isAssignmentInitialCfgEmpty {
		log.C(ctx).Infof("The formation assignment with ID: %s has no initial configuration. Returning without processing the %q operator.", formationAssignment.ID, JSONSchemaValidatorOperator)
		return true, nil
	}

	if len(i.ExceptFormationTypes) > 0 {
		for _, exceptFormationType := range i.ExceptFormationTypes {
			if i.FormationTemplateName == exceptFormationType {
				log.C(ctx).Infof("Skipping initial config validation for assignment with ID: %s since the formation type: %s is excluded from validation", formationAssignment.ID, i.FormationTemplateName)
				return true, nil
			}
		}
	}

	if len(i.ExceptSubtypes) > 0 || len(i.OnlyForSourceSubtypes) > 0 {
		sourceSubType, err := e.getObjectSubtype(ctx, i.TenantID, i.SourceResourceType, i.SourceResourceID)
		if err != nil {
			return false, errors.Wrapf(err, "while getting subtype of resource with type: %s and ID: %s", i.SourceResourceType, i.SourceResourceID)
		}

		if len(i.ExceptSubtypes) > 0 {
			for _, exceptSubtype := range i.ExceptSubtypes {
				if sourceSubType == exceptSubtype {
					log.C(ctx).Infof("Skipping initial config validation for assignment with ID: %s, source resource type: %s and ID: %s since it's part of the except subtypes configuration", formationAssignment.ID, i.SourceResourceType, i.SourceResourceID)
					return true, nil
				}
			}
		}

		if len(i.OnlyForSourceSubtypes) > 0 {
			sourceSubtypeIsSupported := false
			for _, subtype := range i.OnlyForSourceSubtypes {
				if sourceSubType == subtype {
					sourceSubtypeIsSupported = true
					break
				}
			}
			if !sourceSubtypeIsSupported {
				log.C(ctx).Infof("Skipping initial config validation for assignment with ID: %s since the source with type: %s and subtype: %s is not part of the only source subtypes configuration", formationAssignment.ID, i.SourceResourceType, sourceSubType)
				return true, nil
			}
		}
	}

	if i.JSONSchema == "" {
		return false, errors.Errorf("Initial configuration for formation assignment with ID: %s is provided but could not be validated due to empty JSON schema.", formationAssignment.ID)
	}

	log.C(ctx).Infof("Validating the initial config for formation assignment with ID: %s against JSON Schema: %s", formationAssignment.ID, i.JSONSchema)
	validator, err := jsonschema.NewValidatorFromStringSchema(i.JSONSchema)
	if err != nil {
		return false, errors.Wrap(err, "while creating JSON Schema validator")
	}

	result, err := validator.ValidateString(string(formationAssignment.Value))
	if err != nil {
		return false, errors.Wrap(err, "while validating the initial config against the JSON Schema")
	}
	if !result.Valid {
		return false, errors.Wrap(result.Error, "while validating the initial config against the JSON Schema")
	}

	log.C(ctx).Infof("Finished executing operator: %s", JSONSchemaValidatorOperator)
	return true, nil
}
