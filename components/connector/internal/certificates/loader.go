package certificates

import (
	"time"

	"k8s.io/client-go/discovery"

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
	readinessCh                 chan<- struct{}
	apiServerClient             discovery.ServerVersionInterface
	repository                  secrets.Repository
	caSecretName                types.NamespacedName
	rootCACertificateSecretName types.NamespacedName
	exitCh                      <-chan struct{}
}

func NewCertificateLoader(certificatesCache Cache,
	repository secrets.Repository,
	caSecretName types.NamespacedName,
	rootCACertificateSecretName types.NamespacedName,
	readinessCh chan<- struct{}, apiServerClient discovery.ServerVersionInterface,
	exitCh <-chan struct{}) Loader {
	return &certLoader{
		certificatesCache:           certificatesCache,
		readinessCh:                 readinessCh,
		apiServerClient:             apiServerClient,
		repository:                  repository,
		caSecretName:                caSecretName,
		rootCACertificateSecretName: rootCACertificateSecretName,
		exitCh:                      exitCh,
	}
}

func (cl *certLoader) Run() {
	notificationCh := make(chan struct{}, 1)
	shouldNotify := true
	go cl.notifyOnConnection(notificationCh)

	for {
		allLoaded := cl.loadSecretsToCache()
		if allLoaded && shouldNotify {
			shouldNotify = false
			log.Info("All needed secrets are loaded, notifying readiness")
			cl.readinessCh <- struct{}{}
			close(cl.readinessCh)
		}

		select {
		case <-notificationCh:
			log.Info("Received notification")
		case <-cl.exitCh:
			log.Info("Received exit signal, loader exiting..")
			return
		case <-time.After(interval):
		}
	}
}

func (cl *certLoader) notifyOnConnection(notificationCh chan<- struct{}) {
	for {
		_, err := cl.apiServerClient.ServerVersion()
		if err != nil {
			log.Errorf("Failed to access API Server: %s.", err.Error())
			time.Sleep(time.Second * 2)
		} else {
			log.Info("Sending notification to fetch secrets")
			notificationCh <- struct{}{}
			break
		}
	}
}

func (cl *certLoader) loadSecretsToCache() bool {
	allLoaded := true

	if cl.caSecretName.Name != "" {
		isLoaded := cl.loadSecretToCache(cl.caSecretName)
		allLoaded = isLoaded
	}
	if cl.rootCACertificateSecretName.Name != "" {
		isLoaded := cl.loadSecretToCache(cl.rootCACertificateSecretName)
		allLoaded = allLoaded && isLoaded
	}

	return allLoaded
}

func (cl *certLoader) loadSecretToCache(name types.NamespacedName) bool {
	secretData, appError := cl.repository.Get(name)

	if appError != nil {
		log.Errorf("Failed to load secret %s to cache: %s", name.String(), appError.Error())
		return false
	}

	log.Debugf("Putting %s secret in cache", name.Name)
	cl.certificatesCache.Put(name.Name, secretData)
	return true
}
