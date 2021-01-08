package certificates

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"k8s.io/apimachinery/pkg/types"
)

const interval = 1 * time.Minute
const certLoaderCorrelationID = "cert-loader"

type Loader interface {
	Run(ctx context.Context)
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

func (cl *certLoader) Run(ctx context.Context) {
	ctx = cl.configureLogger(ctx)
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping cert loader...")
			return
		default:
		}
		if cl.caCertSecret.Name != "" {
			cl.loadSecretToCache(ctx, cl.caCertSecret)
		}
		if cl.rootCACertSecret.Name != "" {
			cl.loadSecretToCache(ctx, cl.rootCACertSecret)
		}
		time.Sleep(interval)
	}
}

func (cl *certLoader) loadSecretToCache(ctx context.Context, secret types.NamespacedName) {
	secretData, appError := cl.secretsRepository.Get(ctx, secret)

	if appError != nil {
		log.C(ctx).WithError(appError).Errorf("Failed to load secret %s to cache", secret.String())
		return
	}

	cl.certsCache.Put(secret.Name, secretData)
}

func (cl *certLoader) configureLogger(ctx context.Context) context.Context {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certLoaderCorrelationID)
	return log.ContextWithLogger(ctx, entry)
}
