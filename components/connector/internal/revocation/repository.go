package revocation

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/retry"
)

//go:generate mockery -name=Manager
type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(configmap *v1.ConfigMap) (*v1.ConfigMap, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

//go:generate mockery -name=RevocationListRepository
type RevocationListRepository interface {
	Insert(hash string) error
	Contains(hash string) bool
}

type revocationListRepository struct {
	configListManager   Manager
	configMapName       string
	revocationListCache Cache
}

func NewRepository(configListManager Manager, revocationListCache Cache, configMapName string) RevocationListRepository {
	return &revocationListRepository{
		configListManager:   configListManager,
		configMapName:       configMapName,
		revocationListCache: revocationListCache,
	}
}

func (r *revocationListRepository) Insert(hash string) error {
	configMap, err := r.configListManager.Get(r.configMapName, metav1.GetOptions{})
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
		_, err = r.configListManager.Update(updatedConfigMap)
		return err
	})

	return err
}

func (r *revocationListRepository) Contains(hash string) bool {
	configMap := r.revocationListCache.Get()

	found := false
	if configMap != nil {
		_, found = configMap[hash]
	}

	return found
}
