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
	certificatesCache           Cache
	repository                  secrets.Repository
	caSecretName                types.NamespacedName
	rootCACertificateSecretName types.NamespacedName
}

func NewCertificateLoader(certificatesCache Cache,
	repository secrets.Repository,
	caSecretName types.NamespacedName,
	rootCACertificateSecretName types.NamespacedName) Loader {
	return &certLoader{
		certificatesCache:           certificatesCache,
		repository:                  repository,
		caSecretName:                caSecretName,
		rootCACertificateSecretName: rootCACertificateSecretName,
	}
}

func (cl *certLoader) Run() {
	for {
		if cl.caSecretName.Name != "" {
			cl.loadSecretToCache(cl.caSecretName)
		}
		if cl.rootCACertificateSecretName.Name != "" {
			cl.loadSecretToCache(cl.rootCACertificateSecretName)
		}
		time.Sleep(interval)
	}
}

func (cl *certLoader) loadSecretToCache(name types.NamespacedName) {
	secretData, appError := cl.repository.Get(name)

	if appError != nil {
		log.Errorf("Failed to load secret %s to cache: %s", name.String(), appError.Error())
		return
	}

	cl.certificatesCache.Put(name.Name, secretData)
}
