package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=AccountProvider -output=automock -outpkg=automock -case=underscore
type AccountProvider interface {
	CompassCredentials(hyperscalerType Type, tenantName string) (Credentials, error)
	GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error)
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

func HyperscalerTypeFromProvisionInput(input *gqlschema.ProvisionRuntimeInput) (Type, error) {

	if input == nil {
		return Type(""), errors.New("can't determine hyperscaler type because ProvisionRuntimeInput not specified (was nil)")
	}
	if input.ClusterConfig == nil {
		return Type(""), errors.New("can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig not specified (was nil)")
	}

	if input.ClusterConfig.GcpConfig != nil {
		return GCP, nil
	}

	if input.ClusterConfig.GardenerConfig != nil {
		return HyperscalerTypeFromProviderString(input.ClusterConfig.GardenerConfig.Provider)
	}

	return Type(""), errors.New("can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig hyperscaler config not specified")
}

func (p *accountProvider) CompassCredentials(hyperscalerType Type, tenantName string) (Credentials, error) {

	return p.compassPool.Credentials(hyperscalerType, tenantName)
}

func (p *accountProvider) GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error) {

	if p.gardenerPool == nil {
		return Credentials{},
			errors.New("failed to get Gardener Credentials. Gardener Account pool is not configured")
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
	return credential.Name, nil

}

func (p *accountProvider) GardenerSecretName(input *gqlschema.GardenerConfigInput, tenantName string) (string, error) {

	if len(input.TargetSecret) > 0 {
		return input.TargetSecret, nil
	}

	hyperscalerType, err := HyperscalerTypeFromProviderString(input.Provider)

	if err != nil {
		return "", err
	}

	credential, err := p.GardenerCredentials(hyperscalerType, tenantName)

	if err != nil {
		return "", err
	}
	return credential.Name, nil

}
