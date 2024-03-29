package secrets

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets/mocks"

	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appName   = "appName"
	namespace = "kyma-integration"
)

var (
	namespacedName = types.NamespacedName{
		Name:      appName,
		Namespace: namespace,
	}

	expectedCaCrt = []byte("caCrtEncoded")
	expectedCaKey = []byte("caKeyEncoded")
)

func TestRepository_Get(t *testing.T) {

	t.Run("should get secret", func(t *testing.T) {
		// given
		ctx := context.Background()

		secretMap := make(map[string][]byte)
		secretMap["ca.crt"] = expectedCaCrt
		secretMap["ca.key"] = expectedCaKey

		secretsManager := &mocks.Manager{}
		secretsManager.On("Get", ctx, appName, metav1.GetOptions{}).Return(&v1.Secret{Data: secretMap}, nil)

		repository := NewRepository(prepareManagerConstructor(secretsManager))

		// when
		secretData, err := repository.Get(ctx, namespacedName)

		// then
		require.NoError(t, err)

		assert.Equal(t, expectedCaCrt, secretData["ca.crt"])
		assert.Equal(t, expectedCaKey, secretData["ca.key"])
	})

	t.Run("should fail in case secret not found", func(t *testing.T) {
		// given
		ctx := context.Background()

		k8sNotFoundError := &k8serrors.StatusError{
			ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound},
		}
		secretsManager := &mocks.Manager{}
		secretsManager.On("Get", ctx, appName, metav1.GetOptions{}).Return(nil, k8sNotFoundError)

		repository := NewRepository(prepareManagerConstructor(secretsManager))

		// when
		secretData, err := repository.Get(ctx, namespacedName)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.Nil(t, secretData)
	})

	t.Run("should fail if couldn't get secret", func(t *testing.T) {
		// given
		ctx := context.Background()

		secretsManager := &mocks.Manager{}
		secretsManager.On("Get", ctx, appName, metav1.GetOptions{}).Return(nil, &k8serrors.StatusError{})

		repository := NewRepository(prepareManagerConstructor(secretsManager))

		// when
		secretData, err := repository.Get(ctx, namespacedName)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, secretData)
	})
}

func prepareManagerConstructor(manager Manager) ManagerConstructor {
	return func(namespace string) Manager {
		return manager
	}
}
