package hyperscaler_account

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

// HyperscalerAccountProvider provides access to the Kubernetes instance used to provision a new cluster by
// means of K8S secret containing credentials to the appropriate account. There are different pools of
// credentials for different use-cases (Gardener, Compass).
type HyperscalerAccountProvider interface {
	CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (HyperscalerCredential, error)
	GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (HyperscalerCredential, error)
	CompassSecretName(input *gqlschema.ProvisionRuntimeInput) (string, error)
	GardenerSecretName(input *gqlschema.GardenerConfigInput) (string, error)
}

type hyperscalerAccountProvider struct {
	compassHyperscalerAccountPool  HyperscalerAccountPool
	gardenerHyperscalerAccountPool HyperscalerAccountPool
}

func NewHyperscalerAccountProvider(
	compassHyperscalerAccountPool HyperscalerAccountPool,
	gardenerHyperscalerAccountPool HyperscalerAccountPool) HyperscalerAccountProvider {

	return &hyperscalerAccountProvider{
		compassHyperscalerAccountPool:  compassHyperscalerAccountPool,
		gardenerHyperscalerAccountPool: gardenerHyperscalerAccountPool,
	}
}

func (h *hyperscalerAccountProvider) CompassCredentials(hyperscalerType HyperscalerType, tenantName string) (HyperscalerCredential, error) {

	return h.compassHyperscalerAccountPool.Credential(hyperscalerType, tenantName)

}

func (h *hyperscalerAccountProvider) GardnerCredentials(hyperscalerType HyperscalerType, tenantName string) (HyperscalerCredential, error) {

	return h.gardenerHyperscalerAccountPool.Credential(hyperscalerType, tenantName)

}

func (h *hyperscalerAccountProvider) CompassSecretName(input *gqlschema.ProvisionRuntimeInput) (string, error) {

	if input.Credentials != nil && len(input.Credentials.SecretName) > 0 {
		return input.Credentials.SecretName, nil
	}

	// If no credentials given to connect to target cluster, try to get credentials from compassHyperscalerAccountPool

	// TODO: how to get provider type in Compass use case?
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

func (h *hyperscalerAccountProvider) GardenerSecretName(input *gqlschema.GardenerConfigInput) (string, error) {

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
