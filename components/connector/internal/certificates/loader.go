package certificates

import (
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

const interval = 1 * time.Minute

type Loader interface {
	Run()
}

type certLoader struct {
	certsCache        Cache
	secretsRepository secrets.Repository
	caCertSecret      types.NamespacedName
	rootCACertSecret  types.NamespacedName
}

func NewCertificateLoader(certsCache Cache,
	secretsRepository secrets.Repository,
	caCertSecret types.NamespacedName,
	rootCACertSecretName types.NamespacedName) Loader {
	return &certLoader{
		certsCache:        certsCache,
		secretsRepository: secretsRepository,
		caCertSecret:      caCertSecret,
		rootCACertSecret:  rootCACertSecretName,
	}
}

func (cl *certLoader) Run() {
	for {
		if cl.caCertSecret.Name != "" {
			cl.loadSecretToCache(cl.caCertSecret)
		}
		if cl.rootCACertSecret.Name != "" {
			cl.loadSecretToCache(cl.rootCACertSecret)
		}
		time.Sleep(interval)
	}
}

func (cl *certLoader) loadSecretToCache(secret types.NamespacedName) {
	secretData, appError := cl.secretsRepository.Get(secret)

	if appError != nil {
		log.Errorf("Failed to load secret %s to cache: %s", secret.String(), appError.Error())
		return
	}

	cl.certsCache.Put(secret.Name, secretData)
}
