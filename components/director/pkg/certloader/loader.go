package certloader

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"

	"github.com/pkg/errors"

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

type certificatesLoader struct {
	certCache         *certificatesCache
	secretManager     Manager
	secretName        string
	reconnectInterval time.Duration
}

// NewCertificatesLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewCertificatesLoader(certCache *certificatesCache, secretManager Manager, secretName string, reconnectInterval time.Duration) Loader {
	return &certificatesLoader{
		certCache:         certCache,
		secretManager:     secretManager,
		secretName:        secretName,
		reconnectInterval: reconnectInterval,
	}
}

// StartCertLoader prepares and run certificate loader goroutine
func StartCertLoader(ctx context.Context, externalClientCertSecret string) (Cache, error) {
	parsedCertSecret, err := namespacedname.Parse(externalClientCertSecret)
	if err != nil {
		return nil, err
	}
	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}
	certCache := NewCertificateCache()
	certLoader := NewCertificatesLoader(certCache, k8sClientSet.CoreV1().Secrets(parsedCertSecret.Namespace), parsedCertSecret.Name, time.Second)
	go certLoader.Run(ctx)

	return certCache, nil
}

// Run uses kubernetes watch mechanism to listen for resource changes and update certificate cache
func (cl *certificatesLoader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	cl.startKubeWatch(ctx)
}

func (cl *certificatesLoader) startKubeWatch(ctx context.Context) {
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

func (cl *certificatesLoader) processEvents(ctx context.Context, events <-chan watch.Event) {
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
				tlsCert, err := cl.parseCertificate(ctx, secret.Data)
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

func (cl *certificatesLoader) parseCertificate(ctx context.Context, secretData map[string][]byte) (*tls.Certificate, error) {
	log.C(ctx).Info("Parsing certificate data from secret...")
	certBytes := secretData["tls.crt"]
	privateKeyBytes := secretData["tls.key"]

	if certBytes == nil || privateKeyBytes == nil {
		return nil, errors.New("There is no certificate data in the secret")
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

	log.C(ctx).Info("Successfully parse certificate from secret data")
	return &tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  privateKey,
	}, nil
}
