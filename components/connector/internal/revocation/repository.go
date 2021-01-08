package revocation

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/retry"
)

//go:generate mockery -name=Manager
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(ctx context.Context, configmap *v1.ConfigMap, options metav1.UpdateOptions) (*v1.ConfigMap, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

//go:generate mockery -name=RevokedCertificatesRepository
type RevokedCertificatesRepository interface {
	Insert(ctx context.Context, hash string) error
	Contains(hash string) bool
}

type revokedCertifiatesRepository struct {
	configMapManager  Manager
	configMapName     string
	revokedCertsCache Cache
}

func NewRepository(configMapManager Manager, configMapName string, revokedCertsCache Cache) RevokedCertificatesRepository {
	return &revokedCertifiatesRepository{
		configMapManager:  configMapManager,
		configMapName:     configMapName,
		revokedCertsCache: revokedCertsCache,
	}
}

func (r *revokedCertifiatesRepository) Insert(ctx context.Context, hash string) error {
	configMap, err := r.configMapManager.Get(ctx, r.configMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	revokedCerts := configMap.Data
	if revokedCerts == nil {
		revokedCerts = map[string]string{}
	}
	revokedCerts[hash] = hash

	updatedConfigMap := configMap
	updatedConfigMap.Data = revokedCerts

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err = r.configMapManager.Update(ctx, updatedConfigMap, metav1.UpdateOptions{})
		return err
	})

	return err
}

func (r *revokedCertifiatesRepository) Contains(hash string) bool {
	configMap := r.revokedCertsCache.Get()

	found := false
	if configMap != nil {
		_, found = configMap[hash]
	}

	return found
}
