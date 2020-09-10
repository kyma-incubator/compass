package certificates

import (
	"testing"

	"k8s.io/apimachinery/pkg/version"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/namespacedname"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets/mocks"
)

type fakeApiServerClient struct {
	err error
}

func (f *fakeApiServerClient) ServerVersion() (*version.Info, error) {
	return nil, f.err
}

func TestLoader(t *testing.T) {
	t.Run("should load certs to cache and send notification when ready", func(t *testing.T) {
		const (
			caSecretName = "compass-system/caSecretName"
			roSecretName = "kyma-system/roSecretName"
		)

		var (
			CAmap = map[string][]byte{"testCA": []byte("test")}
			Romap = map[string][]byte{"testRO": []byte("test")}
		)

		notificationCh := make(chan struct{}, 1)
		exitCh := make(chan struct{}, 1)

		cache := NewCertificateCache()
		repo := &mocks.Repository{}

		repo.On("Get", namespacedname.Parse(caSecretName)).Return(CAmap, nil)
		repo.On("Get", namespacedname.Parse(roSecretName)).Return(Romap, nil)

		loader := NewCertificateLoader(cache, repo, namespacedname.Parse(caSecretName), namespacedname.Parse(roSecretName), notificationCh, &fakeApiServerClient{nil}, exitCh)
		go loader.Run()

		<-notificationCh

		ca, err := cache.Get("caSecretName")
		require.NoError(t, err)
		require.Equal(t, ca, CAmap)

		ro, err := cache.Get("roSecretName")
		require.NoError(t, err)
		require.Equal(t, ro, Romap)

		// force stop of loader.Run method
		exitCh <- struct{}{}
	})
}
