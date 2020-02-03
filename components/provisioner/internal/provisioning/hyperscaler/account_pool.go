package hyperscaler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

// Get a known HyperscalerType from the input string using case-insensitive matching.
// Returns HyperscalerType or Error if provider string not known HyperscalerType.
func HyperscalerTypeFromProviderString(provider string) (HyperscalerType, error) {

	hyperscalerType := HyperscalerType(strings.ToLower(provider))

	switch hyperscalerType {
	case GCP, Azure, AWS:
		return hyperscalerType, nil
	}
	return "", errors.Errorf("Unknown Hyperscaler provider type: %s", provider)
}

// Credentials holds credentials needed to connect to a particular Hyperscaler account
type Credentials struct {

	// Identifying name for this credential; allows looking up by name in a pool of credentials
	CredentialName string

	// The Hyperscaler accountProvider (Azure, GCP, etc)
	HyperscalerType HyperscalerType

	// The tenant name for the Kyma account
	TenantName string

	// The contents/data for the credential used to connect to a particular Hyperscaler account (Kubeconfig data, ServiceAccount data, etc)
	CredentialData map[string][]byte
}

// AccountPool represents a collection of credentials used by Hydroform/Terraform to provision clusters.
type AccountPool interface {

	// Retrieve a Credentials from the pool based on accountProvider HyperscalerType and tenantName
	Credentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
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
	// mutex to allow locking the critical section of code between fetching an unassigned secret and assigning it
	mux sync.Mutex
}

func (p *secretsAccountPool) Credentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {

	// query the secrets client to get a secret with labels matching hyperscalerType and tenantName
	labelSelector := fmt.Sprintf("tenantName=%s,hyperscalerType=%s", tenantName, hyperscalerType)
	secret, err := getSecret(p.secretsClient, labelSelector)

	if err != nil {
		return Credentials{}, err
	}
	if secret != nil {
		return credentialsFromSecret(secret, hyperscalerType, tenantName), nil
	}

	// assigned secret not found, query again for secret with no tenant assigned for this this hyperscalerType:
	labelSelector = fmt.Sprintf("!tenantName, hyperscalerType=%s", hyperscalerType)
	// lock so that only one thread can fetch an unassigned secret and assign it (update secret with tenantName)
	p.mux.Lock()
	defer p.mux.Unlock()
	secret, err = getSecret(p.secretsClient, labelSelector)

	if err != nil {
		return Credentials{}, err
	}
	if secret != nil {
		secret.Labels["tenantName"] = tenantName
		updatedSecret, err := p.secretsClient.Update(secret)
		if err != nil {
			return Credentials{},
				errors.Wrapf(err, "AccountPool error while updating secret with tenantName: %s", tenantName)
		}
		return credentialsFromSecret(updatedSecret, hyperscalerType, tenantName), nil
	}

	return Credentials{},
		errors.Errorf("AccountPool failed to find unassigned secret for hyperscalerType: %s",
			hyperscalerType)

}

func getSecret(secretsClient corev1.SecretInterface, labelSelector string) (*apiv1.Secret, error) {
	secrets, err := secretsClient.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil,
			errors.Wrapf(err, "AccountPool error during secret list for LabelSelector: %s", labelSelector)
	}

	if secrets != nil && len(secrets.Items) > 0 {
		return &secrets.Items[0], nil
	}
	// not found, return nil secret and nil error
	return nil, nil
}

func credentialsFromSecret(secret *apiv1.Secret, hyperscalerType HyperscalerType, tenantName string) Credentials {
	return Credentials{
		CredentialName:  secret.Name,
		HyperscalerType: hyperscalerType,
		TenantName:      tenantName,
		CredentialData:  secret.Data,
	}
}
