package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// Validate missing godoc
func (i FormationInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.State, validation.In(model.InitialFormationState)),
	)
}
