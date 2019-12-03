package credentials

type TargetProvider string
type Kubeconfig []byte
type ServiceAccount []byte

const (
	GCP   TargetProvider = "GCP"
	Azure TargetProvider = "Azure"
	AWS   TargetProvider = "AWS"
)

type Credentials struct {
	ServiceAccount ServiceAccount
}

type GardenerCredentials struct {
	ServiceAccountKubeconfig Kubeconfig
	SecretName               string
}

// GardenerProvider extracts credentials from Gardener
type GardenerProvider interface {
	// Gets credentials based on provider type (e.g. Azure), credential name and account name
	Get(target TargetProvider, credentialName string, accountName string) (GardenerCredentials, error)
}

// Provider extracts credentials for "bring your own license" use case
type Provider interface {
	// Gets credentials based on provider type (e.g. Azure), credential name and account name
	Get(target TargetProvider, credentialName string, accountName string) (Credentials, error)
}

// NewGardenerProvider creates GardenerProvider responsible for fetching credentials from Gardener
func NewGardenerProvider(gardenerKubeconfig Kubeconfig) GardenerProvider {
	return nil
}

// NewProvider creates Provider responsible for fetching credentials for "bring your own license" use case
func NewProvider() Provider {
	return nil
}
