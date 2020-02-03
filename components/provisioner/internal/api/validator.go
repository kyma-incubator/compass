package api

import (
	"errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

const RuntimeAgent = "compass-runtime-agent"

//go:generate mockery -name=Validator
type Validator interface {
	ValidateInput(input gqlschema.ProvisionRuntimeInput) error
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

func (v *validator) ValidateInput(input gqlschema.ProvisionRuntimeInput) error {
	if input.KymaConfig == nil {
		return errors.New("cannot provision Runtime since Kyma config is missing")
	}

	if len(input.KymaConfig.Components) == 0 {
		return errors.New("cannot provision Runtime since Kyma components list is empty")
	}

	if !configContainsRuntimeAgentComponent(input.KymaConfig.Components) {
		return errors.New("cannot provision Runtime since Kyma components list does not contain Compass Runtime Agent")
	}

	if input.RuntimeInput == nil {
		return errors.New("cannot provision Runtime since runtime input is missing")
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
