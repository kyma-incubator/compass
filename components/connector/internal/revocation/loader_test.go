package revocation

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_revocationListLoader(t *testing.T) {
	configMapName := "revokedCertificates"

	t.Run("should load certs", func(t *testing.T) {
		// given
		cache := NewCache()
		someHash := "someHash"
		configListManagerMock := &mocks.Manager{}
		configMapName := "revokedCertificates"

		loader := NewRevocationListLoader(cache, configListManagerMock, configMapName)

		// when
		loader.Run()

		// then
		assert.Equal(t, isPresent, false)
		configListManagerMock.AssertExpectations(t)
	})

}
