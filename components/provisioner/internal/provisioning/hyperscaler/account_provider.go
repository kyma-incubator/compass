package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
)

// AccountProvider provides access to the Kubernetes instance used to provision a new cluster by
// means of K8S secret containing credentials to the appropriate account. There are different pools of
// credentials for different use-cases (Gardener, Compass).
//go:generate mockery -name=AccountProvider
type AccountProvider interface {
	CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
	GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
	CompassSecretName(input *gqlschema.ProvisionRuntimeInput) (string, error)
	GardenerSecretName(input *gqlschema.GardenerConfigInput) (string, error)
}

type accountProvider struct {
	compassPool  AccountPool
	gardenerPool AccountPool
	tenantName   string
}

func NewAccountProvider(
	compassPool AccountPool,
	gardenerPool AccountPool,
	tenantName string) AccountProvider {

	return &accountProvider{
		compassPool:  compassPool,
		gardenerPool: gardenerPool,
		tenantName:   tenantName,
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

	return p.gardenerPool.Credentials(hyperscalerType, tenantName)

}

func (p *accountProvider) CompassSecretName(input *gqlschema.ProvisionRuntimeInput) (string, error) {

	if input.Credentials != nil && len(input.Credentials.SecretName) > 0 {
		return input.Credentials.SecretName, nil
	}

	// If no credentials given to connect to target cluster, try to get credentials from compassHyperscalerAccountPool
	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	if err != nil {
		return "", err
	}

	credential, err := p.CompassCredentials(hyperscalerType, p.tenantName)

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}

func (p *accountProvider) GardenerSecretName(input *gqlschema.GardenerConfigInput) (string, error) {

	// If Gardener config already has a TargetSecret, just return that
	if len(input.TargetSecret) > 0 {
		return input.TargetSecret, nil
	}

	hyperscalerType, err := HyperscalerTypeFromProviderString(input.Provider)

	if err != nil {
		return "", err
	}

	credential, err := p.GardnerCredentials(hyperscalerType, p.tenantName)

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}
