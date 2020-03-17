// Currently added in this package to be able to access plans schemas.
// In future we need to refactor that approach and in one single place have Plan Schemas, Plan Schemas Validator, Plan Schemas Defaults
// Currently it is shared between `broker`, `provider` and `proces/provisioning/input` packages.
package broker

import (
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=JSONSchemaValidator -output=automock -outpkg=automock -case=underscore

type JSONSchemaValidator interface {
	ValidateString(json string) (jsonschema.ValidationResult, error)
}

type PlansSchemaValidator map[string]JSONSchemaValidator

func NewPlansSchemaValidator() (PlansSchemaValidator, error) {
	planIDs := []string{GcpPlanID, AzurePlanID}
	validators := PlansSchemaValidator{}

	for _, id := range planIDs {
		schema := string(plans[id].provisioningRawSchema)
		validator, err := jsonschema.NewValidatorFromStringSchema(schema)
		if err != nil {
			return nil, errors.Wrapf(err, "while creating schema validator for Plan ID %s", id)
		}
		validators[id] = validator
	}

	return validators, nil
}
