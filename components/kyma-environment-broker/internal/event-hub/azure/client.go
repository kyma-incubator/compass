package azure

type HyperscalerProvider interface {
	GetClientOrDie(config *Config) NamespaceClientInterface
}

var _ HyperscalerProvider = (*azureClient)(nil)

type azureClient struct{}

func NewAzureClient() HyperscalerProvider {
	return &azureClient{}
}

func (ac *azureClient) GetClientOrDie(config *Config) NamespaceClientInterface {
	return GetNamespacesClientOrDie(config)
}
