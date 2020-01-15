package api

import (
	"errors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

//go:generate mockery -name=Validator
type Validator interface {
	ValidateInput(input gqlschema.ProvisionRuntimeInput) error
	ValidateTenant(runtimeID, tenant string) error
}

type validator struct {
	persistenceService persistence.Service
}

func NewValidator(persistenceService persistence.Service) Validator {
	return &validator{
		persistenceService: persistenceService,
	}
}

func (v *validator) ValidateInput(input gqlschema.ProvisionRuntimeInput) error {
	if input.KymaConfig == nil {
		return errors.New("cannot provision Runtime since Kyma config is missing")
	}

	if len(input.KymaConfig.Components) == 0 {
		return errors.New("cannot provision Runtime since Kyma components list is empty")
	}

	if input.RuntimeInput == nil {
		return errors.New("cannot provision Runtime since runtime input is missing")
	}

	if input.Credentials == nil {
		return errors.New("cannot provision Runtime since credentials are missing")
	}

	return nil
}

func (v *validator) ValidateTenant(runtimeID, tenant string) error {
	dbTenant, err := v.persistenceService.GetTenant(runtimeID)

	if err != nil {
		return err
	}

	if tenant != dbTenant {
		return errors.New("provided tenant does not match tenant used to provision cluster")
	}
	return nil
}
