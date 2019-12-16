package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

// AccountProvider provides access to the Kubernetes instance used to provision a new cluster by
// means of K8S secret containing credentials to the appropriate account. There are different pools of
// credentials for different use-cases (Gardener, Compass).
type AccountProvider interface {
	CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credential, error)
	GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credential, error)
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

func (h *accountProvider) CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (Credential, error) {

	return h.compassPool.Credential(hyperscalerType, tenantName)

}

func (h *accountProvider) GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (Credential, error) {

	return h.gardenerPool.Credential(hyperscalerType, tenantName)

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

	if len(input.TargetSecret) > 0 {
		return input.TargetSecret, nil
	}

	// TODO: which type resolution is better/more correct?
	//hyperscalerType, err := HyperscalerTypeFromProviderInput(input.ProviderSpecificConfig)
	hyperscalerType, err := HyperscalerTypeFromProviderString(input.Provider)

	if err != nil {
		return "", err
	}
	// TODO: get tenant name from ...?
	credential, err := h.GardnerCredentials(hyperscalerType, "tenant-name")

	if err != nil {
		return "", err
	}
	return credential.CredentialName, nil

}
