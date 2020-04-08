package api

import (
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

const RuntimeAgent = "compass-runtime-agent"

//go:generate mockery -name=Validator
type Validator interface {
	ValidateProvisioningInput(input gqlschema.ProvisionRuntimeInput) error
	ValidateUpgradeInput(input gqlschema.UpgradeRuntimeInput) error
	ValidateTenant(runtimeID, tenant string) error
}

type validator struct {
	readSession dbsession.ReadSession
}

func NewValidator(readSession dbsession.ReadSession) Validator {
	return &validator{
		readSession: readSession,
	}
}

func (v *validator) ValidateProvisioningInput(input gqlschema.ProvisionRuntimeInput) error {
	err := v.validateKymaConfig(input.KymaConfig)
	if err != nil {
		return fmt.Errorf("validation error while starting Runtime provisioning: %s", err.Error())
	}

	if input.RuntimeInput == nil {
		return fmt.Errorf("validation error while starting Runtime provisioning: runtime input is missing")
	}

	return nil
}

func (v *validator) ValidateUpgradeInput(input gqlschema.UpgradeRuntimeInput) error {
	err := v.validateKymaConfig(input.KymaConfig)
	if err != nil {
		return fmt.Errorf("validation error while starting Runtime upgrade: %s", err.Error())
	}

	// TODO: align with others if we should validate versions

	return nil
}

func (v *validator) validateKymaConfig(kymaConfig *gqlschema.KymaConfigInput) error {
	if kymaConfig == nil {
		return errors.New("error: Kyma config not provided")
	}

	if len(kymaConfig.Components) == 0 {
		return errors.New("error: Kyma components list is empty")
	}

	if !configContainsRuntimeAgentComponent(kymaConfig.Components) {
		return errors.New("error: Kyma components list does not contain Compass Runtime Agent")
	}

	return nil
}

func (v *validator) ValidateTenant(runtimeID, tenant string) error {
	dbTenant, err := v.readSession.GetTenant(runtimeID)

	if err != nil {
		return err
	}

	if tenant != dbTenant {
		return errors.New("provided tenant does not match tenant used to provision cluster")
	}
	return nil
}

func configContainsRuntimeAgentComponent(components []*gqlschema.ComponentConfigurationInput) bool {
	for _, component := range components {
		if component.Component == RuntimeAgent {
			return true
		}
	}
	return false
}
