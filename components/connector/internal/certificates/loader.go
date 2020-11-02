package certificates

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"k8s.io/apimachinery/pkg/types"
)

const interval = 1 * time.Minute

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
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("context cancelled, stopping cert loader...")
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
	secretData, appError := cl.secretsRepository.Get(secret)

	if appError != nil {
		log.C(ctx).Errorf("Failed to load secret %s to cache: %s", secret.String(), appError.Error())
		return
	}

	cl.certsCache.Put(secret.Name, secretData)
}
