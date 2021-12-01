package certloader

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const certsListLoaderCorrelationID = "cert-loader"

// Loader provide mechanism to load certificate data into in-memory storage
type Loader interface {
	Run(ctx context.Context)
}

// Manager is a kubernetes secret manager that has methods to work with secret resources
//go:generate mockery --name=Manager --output=automock --outpkg=automock --case=underscore
type Manager interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type localCertificateLoader struct {
	config    Config
	certCache *certificateCache
}

type certificateLoader struct {
	localCertificateLoader
	secretManager     Manager
	secretName        string
	reconnectInterval time.Duration
}

// NewCertificateLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewCertificateLoader(config Config, certCache *certificateCache, secretManager Manager, secretName string, reconnectInterval time.Duration) Loader {
	return &certificateLoader{
		localCertificateLoader: localCertificateLoader{
			config:    config,
			certCache: certCache,
		},
		secretManager:     secretManager,
		secretName:        secretName,
		reconnectInterval: reconnectInterval,
	}
}

func newLocalCertificateLoader(certCache *certificateCache, config Config) Loader {
	return &localCertificateLoader{
		certCache: certCache,
		config:    config,
	}
}

// StartCertLoader prepares and run certificate loader goroutine
func StartCertLoader(ctx context.Context, certLoaderConfig Config) (Cache, error) {
	certCache := NewCertificateCache()

	// If is local setup(only director) and there are certificate and key provided in the environment, populate cert cache with them.
	// If certificate and key are not provided return empty cache.
	if certLoaderConfig.IsLocalSetup {
		if certLoaderConfig.Cert != "" && certLoaderConfig.Key != "" {
			localCertLoader := newLocalCertificateLoader(certCache, certLoaderConfig)
			localCertLoader.Run(ctx)
			return certCache, nil
		}
		return certCache, nil
	}

	parsedCertSecret, err := namespacedname.Parse(certLoaderConfig.ExternalClientCert.Secret)
	if err != nil {
		return nil, err
	}
	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}
	certLoader := NewCertificateLoader(certLoaderConfig, certCache, k8sClientSet.CoreV1().Secrets(parsedCertSecret.Namespace), parsedCertSecret.Name, time.Second)
	go certLoader.Run(ctx)

	return certCache, nil
}

// Run populates the certificate cache on local setup with provided certificate and key
func (lcl *localCertificateLoader) Run(ctx context.Context) {
	ctx = initializeLoggerContext(ctx)
	certSecretData := map[string][]byte{lcl.config.CertKey: []byte(lcl.config.Cert), lcl.config.KeyKey: []byte(lcl.config.Key)}
	tlsCert, err := parseCertificate(ctx, certSecretData, lcl.config)
	log.C(ctx).Error(err)
	lcl.certCache.put(tlsCert)
}

// Run uses kubernetes watch mechanism to listen for resource changes and update certificate cache
func (cl *certificateLoader) Run(ctx context.Context) {
	ctx = initializeLoggerContext(ctx)
	cl.startKubeWatch(ctx)
}

func initializeLoggerContext(ctx context.Context) context.Context {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	return log.ContextWithLogger(ctx, entry)
}

func (cl *certificateLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping certificate watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting certificate watcher for secret changes...")
		watcher, err := cl.secretManager.Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + cl.secretName,
			Watch:         true,
		})
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Could not initialize watcher. Sleep for %s and try again... %v", cl.reconnectInterval.String(), err)
			time.Sleep(cl.reconnectInterval)
			continue
		}
		log.C(ctx).Info("Waiting for certificate secret events...")

		cl.processEvents(ctx, watcher.ResultChan())

		// Cleanup any allocated resources
		watcher.Stop()
		time.Sleep(cl.reconnectInterval)
	}
}

func (cl *certificateLoader) processEvents(ctx context.Context, events <-chan watch.Event) {
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
				log.C(ctx).Info("Updating the cache with certificate data...")
				secret, ok := ev.Object.(*v1.Secret)
				if !ok {
					log.C(ctx).Error("Unexpected error: object is not secret. Try again")
					continue
				}
				tlsCert, err := parseCertificate(ctx, secret.Data, cl.config)
				if err != nil {
					log.C(ctx).WithError(err).Error("Fail during certificate parsing")
				}
				cl.certCache.put(tlsCert)
			case watch.Deleted:
				log.C(ctx).Info("Removing certificate secret data from cache...")
				cl.certCache.put(nil)
			case watch.Error:
				log.C(ctx).Error("Error event is received, stop certificate secret watcher and try again...")
				return
			}
		}
	}
}

func parseCertificate(ctx context.Context, secretData map[string][]byte, config Config) (*tls.Certificate, error) {
	log.C(ctx).Info("Parsing provided certificate data...")
	certBytes := secretData[config.CertKey]
	privateKeyBytes := secretData[config.KeyKey]

	if certBytes == nil || privateKeyBytes == nil {
		return nil, errors.New("There is no certificate data provided")
	}

	clientCrtPem, _ := pem.Decode(certBytes)
	if clientCrtPem == nil {
		return nil, errors.New("Error while decoding certificate pem block")
	}

	clientCert, err := x509.ParseCertificate(clientCrtPem.Bytes)
	if err != nil {
		return nil, err
	}

	privateKeyPem, _ := pem.Decode(privateKeyBytes)
	if privateKeyPem == nil {
		return nil, errors.New("Error while decoding private key pem block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPem.Bytes)
	if err != nil {
		pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyPem.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = pkcs8PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, err
		}
	}

	log.C(ctx).Info("Successfully parse certificate")
	return &tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  privateKey,
	}, nil
}
