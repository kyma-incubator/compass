package certloader

import (
	"context"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/director/pkg/key"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const certsListLoaderCorrelationID = "cert-loader-id"

type CredentialType string

const (
	KeysCredential        CredentialType = "KeysCredentials"
	CertificateCredential CredentialType = "CertificateCredentials"
)

// Loader provide mechanism to load certificate data into in-memory storage
type Loader interface {
	Run(ctx context.Context)
}

// Manager is a kubernetes secret manager that has methods to work with secret resources
//
//go:generate mockery --name=Manager --output=automock --outpkg=automock --case=underscore --disable-version-string
type Manager interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type certificateLoader struct {
	config            Config
	keysConfig        KeysConfig
	certCache         *certificateCache
	keyCache          *keyCache
	secretManagers    map[string]Manager
	secretNamesTypes  map[string]CredentialType
	reconnectInterval time.Duration
}

// NewCertificateLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewCertificateLoader(config Config, keysConfig KeysConfig, certCache *certificateCache, keysCache *keyCache, secretManagers map[string]Manager, secretNames map[string]CredentialType, reconnectInterval time.Duration) Loader {
	return &certificateLoader{
		config:            config,
		keysConfig:        keysConfig,
		certCache:         certCache,
		keyCache:          keysCache,
		secretManagers:    secretManagers,
		secretNamesTypes:  secretNames,
		reconnectInterval: reconnectInterval,
	}
}

// StartCertLoader prepares and run certificate loader goroutine
func StartCertLoader(ctx context.Context, certLoaderConfig Config, keysLoaderConfig KeysConfig) (Cache, error) {
	parsedCertSecret, err := namespacedname.Parse(certLoaderConfig.ExternalClientCertSecret)
	if err != nil {
		return nil, err
	}

	parsedExtSvcCertSecret, err := namespacedname.Parse(certLoaderConfig.ExtSvcClientCertSecret)
	if err != nil {
		return nil, err
	}

	parsedKeysSecret, err := namespacedname.Parse(keysLoaderConfig.KeysSecret)
	if err != nil {
		return nil, err
	}

	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}

	certCache := NewCertificateCache()
	keysCache := NewKeyCache()
	secretManagers := map[string]Manager{
		parsedCertSecret.Namespace:       k8sClientSet.CoreV1().Secrets(parsedCertSecret.Namespace),
		parsedExtSvcCertSecret.Namespace: k8sClientSet.CoreV1().Secrets(parsedExtSvcCertSecret.Namespace),
		parsedKeysSecret.Namespace:       k8sClientSet.CoreV1().Secrets(parsedKeysSecret.Namespace),
	}
	secretNames := map[string]CredentialType{
		parsedCertSecret.Name:       CertificateCredential,
		parsedExtSvcCertSecret.Name: CertificateCredential,
		parsedKeysSecret.Name:       KeysCredential,
	}

	certLoader := NewCertificateLoader(certLoaderConfig, keysLoaderConfig, certCache, keysCache, secretManagers, secretNames, time.Second)
	go certLoader.Run(ctx)

	return certCache, nil
}

// Run uses kubernetes watch mechanism to listen for resource changes and update certificate cache
func (cl *certificateLoader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	cl.startKubeWatch(ctx)
}

func (cl *certificateLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping certificate watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting certificate watchers for secret changes...")

		wg := &sync.WaitGroup{}
		for name, manager := range cl.secretManagers {
			wg.Add(1)

			go func(manager Manager, name string) {
				defer wg.Done()

				watcher, err := manager.Watch(ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=" + name,
					Watch:         true,
				})

				if err != nil {
					log.C(ctx).WithError(err).Errorf("Could not initialize watcher. Sleep for %s and try again... %v", cl.reconnectInterval.String(), err)
					time.Sleep(cl.reconnectInterval)
					return
				}
				log.C(ctx).Info("Waiting for certificate secret events...")

				credentialType := cl.secretNamesTypes[name]
				cl.processEvents(ctx, watcher.ResultChan(), name, credentialType)

				// Cleanup any allocated resources
				watcher.Stop()
				time.Sleep(cl.reconnectInterval)
			}(manager, name)
		}

		wg.Wait()
	}
}

func (cl *certificateLoader) processEvents(ctx context.Context, events <-chan watch.Event, secretName string, credentialType CredentialType) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			switch ev.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				log.C(ctx).Infof("Updating the cache with %s data...", credentialType)
				secret, ok := ev.Object.(*v1.Secret)
				if !ok {
					log.C(ctx).Error("Unexpected error: object is not secret. Try again")
					continue
				}
				if credentialType == CertificateCredential {
					tlsCert, err := parseCertificate(ctx, secret.Data, cl.config)
					if err != nil {
						log.C(ctx).WithError(err).Error("Fail during certificate parsing")
					}
					cl.certCache.put(secretName, tlsCert)
				} else {
					keys, err := parseKeys(ctx, secret.Data, cl.keysConfig)
					if err != nil {
						log.C(ctx).WithError(err).Error("Fail during keys parsing")
					}
					cl.keyCache.put(secretName, keys)
				}
			case watch.Deleted:
				log.C(ctx).Info("Removing %s secret data from cache...", credentialType)

				if credentialType == CertificateCredential {
					cl.certCache.put(secretName, nil)
				} else {
					cl.keyCache.put(secretName, nil)
				}
			case watch.Error:
				log.C(ctx).Errorf("Error event is received, stop %s secret watcher and try again...", credentialType)
				return
			}
		}
	}
}

func parseCertificate(ctx context.Context, secretData map[string][]byte, config Config) (*tls.Certificate, error) {
	log.C(ctx).Info("Parsing provided certificate data...")
	certChainBytes, existsCertKey := secretData[config.ExternalClientCertCertKey]
	privateKeyBytes, existsKeyKey := secretData[config.ExternalClientCertKeyKey]

	if existsCertKey && existsKeyKey {
		return cert.ParseCertificateBytes(certChainBytes, privateKeyBytes)
	}

	extSvcCertChainBytes, existsExtSvcCertKey := secretData[config.ExtSvcClientCertCertKey]
	extSvcPrivateKeyBytes, existsExtSvcKeyKey := secretData[config.ExtSvcClientCertKeyKey]

	if existsExtSvcCertKey && existsExtSvcKeyKey {
		return cert.ParseCertificateBytes(extSvcCertChainBytes, extSvcPrivateKeyBytes)
	}

	return nil, errors.New("There is no certificate data provided")
}

func parseKeys(ctx context.Context, secretData map[string][]byte, config KeysConfig) (*KeyStore, error) {
	log.C(ctx).Info("Parsing provided certificate data...")
	publicKeyBytes, existsPublicKey := secretData[config.KeysPublic]
	privateKeyBytes, existsPrivateKey := secretData[config.KeysPrivate]

	if existsPublicKey && existsPrivateKey {
		return key.ParseKeysBytes(publicKeyBytes, privateKeyBytes)
	}

	return nil, errors.New("There is no public/private key data provided")
}
