package hyperscaler

import (
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	machineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
)

// HyperscalerType is one of the large Cloud hosting providers: Azure, GCP, etc
type HyperscalerType string

const (
	// GCP stands for the Google Cloud Platform.
	GCP HyperscalerType = "gcp"
	// Azure stands for the Microsoft Azure Cloud Computing Platform.
	Azure HyperscalerType = "azure"
	// AWS stands for Amazon Web Services.
	AWS HyperscalerType = "aws"
)

func HyperscalerTypeFromProviderString(provider string) (HyperscalerType, error) {

	hyperscalerType := HyperscalerType(strings.ToLower(provider))

	switch hyperscalerType {
	case GCP, Azure, AWS:
		return hyperscalerType, nil
	}
	return "", errors.Errorf("Unknown Hyperscaler accountProvider type: %s", provider)
}

// Credential holds credentials needed to connect to a particular Hyperscaler account
type Credential struct {

	// Identifying name for this credential; allows looking up by name in a pool of credentials
	CredentialName string

	// The Hyperscaler accountProvider (Azure, GCP, etc)
	HyperscalerType HyperscalerType

	// The tenant name for the Kyma account
	TenantName string

	// The contents/data for the credential used to connect to a particular Hyperscaler account (Kubeconfig data, ServiceAccount data, etc)
	Credential []byte
}

// AccountPool represents a collection of credentials used by Hydroform/Terraform to provision clusters.
type AccountPool interface {

	// Retrieve a Credential from the pool based on accountProvider HyperscalerType and tenantName
	Credential(hyperscalerType HyperscalerType, tenantName string) (Credential, error)
}

// Get an instance of of AccountPool that retrieves credentials from Kubernetes secrets
func NewAccountPool(secretsClient corev1.SecretInterface) AccountPool {
	return &secretsAccountPool{
		secretsClient: secretsClient,
	}
}

// private struct for Kubernetes secrets-based implementation of AccountPool
type secretsAccountPool struct {
	secretsClient corev1.SecretInterface
}

func (p *secretsAccountPool) Credential(hyperscalerType HyperscalerType, tenantName string) (Credential, error) {

	// query the secrets client to get a secret with labels matching hyperscalerType and tenantName
	//p.secretsClient.Get()

	var credential = Credential{
		CredentialName:  "the-credential-name",
		HyperscalerType: hyperscalerType,
		TenantName:      tenantName,
		Credential:      nil,
	}

	return credential, nil
}

func ExampleTestUsage() {

	var (
		credentials = []byte("credentials")
		secret      = &apiv1.Secret{
			ObjectMeta: machineryv1.ObjectMeta{Name: "some-secret", Namespace: "some-namespace"},
			Data: map[string][]byte{
				"credentials": credentials,
			},
		}
	)

	mockSecrets := fake.NewSimpleClientset(secret).CoreV1().Secrets("some-namespace")
	pool := NewAccountPool(mockSecrets)
	hyperscalerCredentials, _ := pool.Credential(GCP, "the-tenant-name")

	print(hyperscalerCredentials.CredentialName)
}
