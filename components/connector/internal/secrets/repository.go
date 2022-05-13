package secrets

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ManagerConstructor func(namespace string) Manager

//go:generate mockery --name=Manager --disable-version-string
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Secret, error)
}

//go:generate mockery --name=Repository --disable-version-string
type Repository interface {
	Get(ctx context.Context, name types.NamespacedName) (secretData map[string][]byte, appError apperrors.AppError)
}

type repository struct {
	secretsManagerConstructor ManagerConstructor
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManagerConstructor ManagerConstructor) Repository {
	return &repository{
		secretsManagerConstructor: secretsManagerConstructor,
	}
}

func (r *repository) Get(ctx context.Context, secret types.NamespacedName) (secretData map[string][]byte, appError apperrors.AppError) {
	secretsManager := r.secretsManagerConstructor(secret.Namespace)
	secretObj, err := secretsManager.Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("secret %s not found", secret)
		}
		return nil, apperrors.Internal("failed to get %s secret, %s", secret, err)
	}

	return secretObj.Data, nil
}
