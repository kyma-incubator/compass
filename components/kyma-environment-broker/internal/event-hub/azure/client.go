package azure

type AzureClientInterface interface {
	GetNamespacesClientOrDie(config *Config) NamespaceClientInterface
}

type azureClient struct {}

func NewAzureClient() AzureClientInterface {
	return &azureClient{}
}

func (ac *azureClient) GetNamespacesClientOrDie(config *Config) NamespaceClientInterface {
	return GetNamespacesClientOrDie(config)
}
