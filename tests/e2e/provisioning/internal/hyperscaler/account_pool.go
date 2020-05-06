package hyperscaler

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type HyperscalerType string

const (
	Azure HyperscalerType = "azure"
)

type Credentials struct {
	CredentialName  string
	HyperscalerType HyperscalerType
	TenantName      string
	CredentialData  map[string][]byte
}

type AccountPool interface {
	Credentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error)
}

// NewAccountPool returns a new AccountPool
func NewAccountPool(secretsClient corev1.SecretInterface) AccountPool {
	return &secretsAccountPool{
		secretsClient: secretsClient,
	}
}

type secretsAccountPool struct {
	secretsClient corev1.SecretInterface
	mux           sync.Mutex
}

// Credentials returns the hyperscaler secret from k8s secret
func (p *secretsAccountPool) Credentials(hyperscalerType HyperscalerType, tenantName string) (Credentials, error) {
	labelSelector := fmt.Sprintf("tenantName=%s,hyperscalerType=%s", tenantName, hyperscalerType)
	secret, err := getK8SSecret(p.secretsClient, labelSelector)

	if err != nil {
		return Credentials{}, err
	}
	if secret != nil {
		return credentialsFromSecret(secret, hyperscalerType, tenantName), nil
	}

	labelSelector = fmt.Sprintf("!tenantName, hyperscalerType=%s", hyperscalerType)
	// lock so that only one thread can fetch an unassigned secret and assign it (update secret with tenantName)
	p.mux.Lock()
	defer p.mux.Unlock()
	secret, err = getK8SSecret(p.secretsClient, labelSelector)

	if err != nil {
		return Credentials{}, errors.Wrapf(err, "failed to fetch k8s secret")
	}
	if secret != nil {
		secret.Labels["tenantName"] = tenantName
		updatedSecret, err := p.secretsClient.Update(secret)
		if err != nil {
			return Credentials{}, errors.Wrapf(err, "accountPool error while updating secret with tenantName: %s", tenantName)
		}
		return credentialsFromSecret(updatedSecret, hyperscalerType, tenantName), nil
	}

	return Credentials{}, errors.Errorf("accountPool failed to find unassigned secret for hyperscalerType: %s",
		hyperscalerType)

}

func getK8SSecret(secretsClient corev1.SecretInterface, labelSelector string) (*apiv1.Secret, error) {
	secrets, err := secretsClient.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "accountPool error during secret list for LabelSelector: %s", labelSelector)
	}

	if secrets != nil && len(secrets.Items) < 1 {
		return nil, errors.Wrapf(err, "no secrets returned for LabelSelector: %s", labelSelector)
	}
	return &secrets.Items[0], nil
}

func credentialsFromSecret(secret *apiv1.Secret, hyperscalerType HyperscalerType, tenantName string) Credentials {
	return Credentials{
		CredentialName:  secret.Name,
		HyperscalerType: hyperscalerType,
		TenantName:      tenantName,
		CredentialData:  secret.Data,
	}
}
