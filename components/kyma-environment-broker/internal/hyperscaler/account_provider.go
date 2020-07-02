package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=AccountProvider -output=automock -outpkg=automock -case=underscore
type AccountProvider interface {
	GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error)
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

	if input.ClusterConfig.GardenerConfig == nil {
		return Type(""), errors.New("can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig.GardenerConfig not specified (was nil)")
	}

	return HyperscalerTypeFromProviderString(input.ClusterConfig.GardenerConfig.Provider)
}

func (p *accountProvider) GardenerCredentials(hyperscalerType Type, tenantName string) (Credentials, error) {

	if p.gardenerPool == nil {
		return Credentials{},
			errors.New("failed to get Gardener Credentials. Gardener Account pool is not configured")
	}

	return p.gardenerPool.Credentials(hyperscalerType, tenantName)
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
