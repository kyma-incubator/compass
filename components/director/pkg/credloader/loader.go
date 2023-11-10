package credloader

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

const certsListLoaderCorrelationID = "creds-loader-id"

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

type loader struct {
	config            CertConfig
	keysConfig        KeysConfig
	certCache         *certificateCache
	keyCache          *KeyCache
	secretManagers    map[string]Manager
	secretNamesTypes  map[string]CredentialType
	reconnectInterval time.Duration
}

// NewCertificateLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewCertificateLoader(config CertConfig, certCache *certificateCache, secretManagers map[string]Manager, secretNames map[string]CredentialType, reconnectInterval time.Duration) Loader {
	return &loader{
		config:            config,
		certCache:         certCache,
		secretManagers:    secretManagers,
		secretNamesTypes:  secretNames,
		reconnectInterval: reconnectInterval,
	}
}

// NewKeyLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewKeyLoader(keysConfig KeysConfig, keysCache *KeyCache, secretManagers map[string]Manager, secretNames map[string]CredentialType, reconnectInterval time.Duration) Loader {
	return &loader{
		keysConfig:        keysConfig,
		keyCache:          keysCache,
		secretManagers:    secretManagers,
		secretNamesTypes:  secretNames,
		reconnectInterval: reconnectInterval,
	}
}

// StartCertLoader prepares and run certificate loader goroutine
func StartCertLoader(ctx context.Context, certLoaderConfig CertConfig) (Cache, error) {
	parsedCertSecret, err := namespacedname.Parse(certLoaderConfig.ExternalClientCertSecret)
	if err != nil {
		return nil, err
	}

	parsedExtSvcCertSecret, err := namespacedname.Parse(certLoaderConfig.ExtSvcClientCertSecret)
	if err != nil {
		return nil, err
	}

	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}

	certCache := NewCertificateCache()
	secretManagers := map[string]Manager{
		parsedCertSecret.Name:       k8sClientSet.CoreV1().Secrets(parsedCertSecret.Namespace),
		parsedExtSvcCertSecret.Name: k8sClientSet.CoreV1().Secrets(parsedExtSvcCertSecret.Namespace),
	}
	secretNames := map[string]CredentialType{
		parsedCertSecret.Name:       CertificateCredential,
		parsedExtSvcCertSecret.Name: CertificateCredential,
	}

	certLoader := NewCertificateLoader(certLoaderConfig, certCache, secretManagers, secretNames, time.Second)
	go certLoader.Run(ctx)

	return certCache, nil
}

func StartKeyLoader(ctx context.Context, keysLoaderConfig KeysConfig) (KeysCache, error) {
	parsedKeysSecret, err := namespacedname.Parse(keysLoaderConfig.KeysSecret)
	if err != nil {
		return nil, err
	}

	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}

	keysCache := NewKeyCache()
	secretManagers := map[string]Manager{
		parsedKeysSecret.Name: k8sClientSet.CoreV1().Secrets(parsedKeysSecret.Namespace),
	}
	secretNames := map[string]CredentialType{
		parsedKeysSecret.Name: KeysCredential,
	}

	certLoader := NewKeyLoader(keysLoaderConfig, keysCache, secretManagers, secretNames, time.Second)
	go certLoader.Run(ctx)

	return keysCache, nil
}

// Run uses kubernetes watch mechanism to listen for resource changes and update certificate cache
func (cl *loader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	cl.startKubeWatch(ctx)
}

func (cl *loader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting watchers for secret changes...")

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
					log.C(ctx).WithError(err).Errorf("Could not initialize watcher for resource name %q. Sleep for %s and try again... %v", name, cl.reconnectInterval.String(), err)
					time.Sleep(cl.reconnectInterval)
					return
				}
				log.C(ctx).Infof("Waiting for secret %s events...", name)

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

func (cl *loader) processEvents(ctx context.Context, events <-chan watch.Event, secretName string, credentialType CredentialType) {
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

func parseCertificate(ctx context.Context, secretData map[string][]byte, config CertConfig) (*tls.Certificate, error) {
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
	log.C(ctx).Info("Parsing provided keys data...")
	dataBytes, exists := secretData[config.KeysData]

	if exists {
		return parseKeysBytes(dataBytes)
	}

	return nil, errors.New("There is no public/private key data provided")
}

func parseKeysBytes(dataBytes []byte) (*KeyStore, error) {
	data := &secretData{}

	if err := json.Unmarshal(dataBytes, data); err != nil {
		return nil, errors.Wrapf(err, "while unmarashalling secret data")
	}

	privateKeyBlock, _ := pem.Decode([]byte(data.PrivateKey))
	if privateKeyBlock == nil {
		return nil, errors.New("Error while decoding private key")
	}

	rsaPrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing private key block")
	}

	publicKeyBlock, _ := pem.Decode([]byte(data.PublicKey))
	if publicKeyBlock == nil {
		return nil, errors.New("Error while decoding public key")
	}

	genericPublicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing private key block")
	}

	rsaPublicKey, ok := genericPublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("error while casting public key")
	}

	return &KeyStore{
		PublicKey:  rsaPublicKey,
		PrivateKey: rsaPrivateKey,
	}, nil
}
