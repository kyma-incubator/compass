package revocation

import (
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestRevocationListRepository(t *testing.T) {

	configMapName := "revokedCertificates"

	t.Run("should return false if value is not present", func(t *testing.T) {
		// given
		cache := NewCache()
		someHash := "someHash"
		configListManagerMock := &mocks.Manager{}
		configMapName := "revokedCertificates"

		repository := NewRepository(configListManagerMock, configMapName, cache)

		// when
		isPresent := repository.Contains(someHash)

		// then
		assert.Equal(t, isPresent, false)
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return true if value is present", func(t *testing.T) {
		// given
		cache := NewCache()
		someHash := "someHash"
		cache.Put(map[string]string{
			someHash: someHash,
		})
		configListManagerMock := &mocks.Manager{}
		configMapName := "revokedCertificates"

		repository := NewRepository(configListManagerMock, configMapName, cache)

		// when
		isPresent := repository.Contains(someHash)

		// then
		assert.Equal(t, isPresent, true)
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should insert value to the list", func(t *testing.T) {
		// given
		cache := NewCache()
		someHash := "someHash"
		configListManagerMock := &mocks.Manager{}

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}).Return(&v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}, nil)

		repository := NewRepository(configListManagerMock, configMapName, cache)

		// when
		err := repository.Insert(someHash)
		require.NoError(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to update config map", func(t *testing.T) {
		// given
		cache := NewCache()
		someHash := "someHash"
		configListManagerMock := &mocks.Manager{}

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}).Return(nil, errors.New("some error"))

		repository := NewRepository(configListManagerMock, configMapName, cache)

		// when
		err := repository.Insert(someHash)
		require.Error(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})
}
