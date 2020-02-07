package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=AccountProvider
type AccountProvider interface {
	CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
	GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
	CompassSecretName(input *gqlschema.ProvisionRuntimeInput, tenantName string) (string, error)
	GardenerSecretName(input *gqlschema.GardenerConfigInput, tenantName string) (string, error)
}

type accountProvider struct {
	compassPool  AccountPool
	gardenerPool AccountPool
}

func NewAccountProvider(compassPool AccountPool, gardenerPool AccountPool) AccountProvider {

	return &accountProvider{
		compassPool:  compassPool,
		gardenerPool: gardenerPool,
	}
}

func HyperscalerTypeFromProvisionInput(input *gqlschema.ProvisionRuntimeInput) (HyperscalerType, error) {

	if input == nil {
		return HyperscalerType(""), errors.New("Can't determine hyperscaler type because ProvisionRuntimeInput not specified (was nil)")
	}
	if input.ClusterConfig == nil {
		return HyperscalerType(""), errors.New("Can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig not specified (was nil)")
	}

	if input.ClusterConfig.GcpConfig != nil {
		return GCP, nil
	}

	return HyperscalerType(""), errors.New("Can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig hyperscaler config not specified")
}

func (p *accountProvider) CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {

	return p.compassPool.Credentials(hyperscalerType, tenantName)
}

func (p *accountProvider) GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {

	if p.gardenerPool == nil {
		return Credentials{},
			errors.New("Failed to get Gardener Credentials. Gardener Account pool is not configured")
	}

	return p.gardenerPool.Credentials(hyperscalerType, tenantName)
}

func (p *accountProvider) CompassSecretName(input *gqlschema.ProvisionRuntimeInput, tenantName string) (string, error) {

	if input.Credentials != nil && len(input.Credentials.SecretName) > 0 {
		return input.Credentials.SecretName, nil
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	if err != nil {
		return "", err
	}

	credential, err := p.CompassCredentials(hyperscalerType, tenantName)

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}

func (p *accountProvider) GardenerSecretName(input *gqlschema.GardenerConfigInput, tenantName string) (string, error) {

	if len(input.TargetSecret) > 0 {
		return input.TargetSecret, nil
	}

	hyperscalerType, err := HyperscalerTypeFromProviderString(input.Provider)

	if err != nil {
		return "", err
	}

	credential, err := p.GardnerCredentials(hyperscalerType, tenantName)

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}
