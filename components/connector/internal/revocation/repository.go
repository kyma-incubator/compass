package revocation

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

//go:generate mockery --name=Manager --disable-version-string
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
}

//go:generate mockery --name=RevokedCertificatesRepository --disable-version-string
type RevokedCertificatesRepository interface {
	Insert(ctx context.Context, hash string) error
}

type revokedCertifiatesRepository struct {
	configMapManager Manager
	configMapName    string
}

func NewRepository(configMapManager Manager, configMapName string) RevokedCertificatesRepository {
	return &revokedCertifiatesRepository{
		configMapManager: configMapManager,
		configMapName:    configMapName,
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
