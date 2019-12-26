package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
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
}

func NewAccountProvider(
	compassPool AccountPool,
	gardenerPool AccountPool) AccountProvider {

	return &accountProvider{
		compassPool:  compassPool,
		gardenerPool: gardenerPool,
	}
}

func (h *accountProvider) CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {

	return h.compassPool.Credentials(hyperscalerType, tenantName)

}

func (h *accountProvider) GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {

	return h.gardenerPool.Credentials(hyperscalerType, tenantName)

}

func (h *accountProvider) CompassSecretName(input *gqlschema.ProvisionRuntimeInput) (string, error) {

	if input.Credentials != nil && len(input.Credentials.SecretName) > 0 {
		return input.Credentials.SecretName, nil
	}

	// If no credentials given to connect to target cluster, try to get credentials from compassHyperscalerAccountPool

	// TODO: how to get accountProvider type in Compass use case?
	hyperscalerType, err := HyperscalerTypeFromProviderString("TBD")
	if err != nil {
		return "", err
	}
	// TODO: get tenant name from ...?
	credential, err := h.CompassCredentials(hyperscalerType, "tenant-name")

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}

func (h *accountProvider) GardenerSecretName(input *gqlschema.GardenerConfigInput) (string, error) {

	// If Gardener config already has a TargetSecret, just return that
	if len(input.TargetSecret) > 0 {
		return input.TargetSecret, nil
	}

	hyperscalerType, err := HyperscalerTypeFromProviderString(input.Provider)

	if err != nil {
		return "", err
	}
	// TODO: get tenant name from ...?
	credential, err := h.GardnerCredentials(hyperscalerType, "tenant2")

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}
