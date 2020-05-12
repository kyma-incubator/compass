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

type Type string

const (
	GCP   Type = "gcp"
	Azure Type = "azure"
	AWS   Type = "aws"
)

func HyperscalerTypeFromProviderString(provider string) (Type, error) {

	hyperscalerType := Type(strings.ToLower(provider))

	switch hyperscalerType {
	case GCP, Azure, AWS:
		return hyperscalerType, nil
	}
	return "", errors.Errorf("unknown Hyperscaler provider type: %s", provider)
}

type Credentials struct {
	Name            string
	HyperscalerType Type
	TenantName      string
	CredentialData  map[string][]byte
}

type AccountPool interface {
	Credentials(hyperscalerType Type, tenantName string) (Credentials, error)
}

func NewAccountPool(secretsClient corev1.SecretInterface) AccountPool {
	return &secretsAccountPool{
		secretsClient: secretsClient,
	}
}

type secretsAccountPool struct {
	secretsClient corev1.SecretInterface
	mux           sync.Mutex
}

func (p *secretsAccountPool) Credentials(hyperscalerType Type, tenantName string) (Credentials, error) {

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
		return Credentials{}, err
	}

	if secret == nil {
		return Credentials{}, errors.Errorf("accountPool failed to find unassigned secret for hyperscalerType: %s", hyperscalerType)
	}

	secret.Labels["tenantName"] = tenantName
	updatedSecret, err := p.secretsClient.Update(secret)
	if err != nil {
		return Credentials{}, errors.Wrapf(err, "accountPool error while updating secret with tenantName: %s", tenantName)
	}

	return credentialsFromSecret(updatedSecret, hyperscalerType, tenantName), nil
}

func getK8SSecret(secretsClient corev1.SecretInterface, labelSelector string) (*apiv1.Secret, error) {
	secrets, err := secretsClient.List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil,
			errors.Wrapf(err, "accountPool error during secret list for LabelSelector: %s", labelSelector)
	}

	if secrets != nil && len(secrets.Items) > 0 {
		return &secrets.Items[0], nil
	}

	return nil, nil
}

func credentialsFromSecret(secret *apiv1.Secret, hyperscalerType Type, tenantName string) Credentials {
	return Credentials{
		Name:            secret.Name,
		HyperscalerType: hyperscalerType,
		TenantName:      tenantName,
		CredentialData:  secret.Data,
	}
}
