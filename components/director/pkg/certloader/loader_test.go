package certloader

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	secretName    = "secretName"
	secretCertKey = "tls.crt"
	secretKeyKey  = "tls.key"
	testCN        = "test-common-name"
)

type testWatch struct {
	events chan watch.Event
}

func (tw *testWatch) close() {
	close(tw.events)
}

func (tw *testWatch) putEvent(ev watch.Event) {
	tw.events <- ev
}

func (tw *testWatch) Stop() {}
func (tw *testWatch) ResultChan() <-chan watch.Event {
	return tw.events
}

func Test_CertificatesLoader(t *testing.T) {
	t.Run("should insert secret data on add event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 1)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.NoError(t, err)
			require.NotNil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})

	t.Run("should insert secret data on modify event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 1)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Modified,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.NoError(t, err)
			require.NotNil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})

	t.Run("should not insert secret data if the event object is not secret", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 1)

		// when
		watcher.putEvent(watch.Event{
			Type:   watch.Added,
			Object: &runtime.Unknown{},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.Error(t, err)
			require.Contains(t, err.Error(), fmt.Sprintf("Client certificate data not found in the cache for key: %s", secretName))
			require.Nil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})

	t.Run("should return empty cache after delete event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 1)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.NoError(t, err)
			require.NotNil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)

		watcher.putEvent(watch.Event{
			Type: watch.Deleted,
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.Error(t, err)
			require.Contains(t, err.Error(), "There is no certificate data in the cache")
			require.Nil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})

	t.Run("should try reconnect when there is error event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 2)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Error,
		})

		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.NoError(t, err)
			require.NotNil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)

		watcher.putEvent(watch.Event{
			Type: watch.Deleted,
		})
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})

	t.Run("should try reconnect when event channel is closed", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 2)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.close()
		newWatcher := &testWatch{
			events: make(chan watch.Event, 50),
		}

		secretManagerMock.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).Return(newWatcher, nil).Once()

		newWatcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert, err := cache.Get()
			require.NoError(t, err)
			require.NotNil(t, tlsCert)
			return true
		}, 2*time.Second, 100*time.Millisecond)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, 100*time.Millisecond)
		secretManagerMock.AssertExpectations(t)
	})
}

func preparation(ctx context.Context, number int) (Cache, *testWatch, *automock.Manager) {
	cache := NewCertificateCache(secretName)
	watcher := &testWatch{
		events: make(chan watch.Event, 50),
	}
	secretManagerMock := &automock.Manager{}
	secretManagerMock.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).Return(watcher, nil).Times(number)
	loader := NewCertificatesLoader(cache, secretManagerMock, secretName, time.Millisecond)
	go loader.Run(ctx)

	return cache, watcher, secretManagerMock
}

func generateTestCertAndKey(t *testing.T, commonName string) (crtPem, keyPem []byte) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		IsCA:         true,
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	parent := template
	certRaw, err := x509.CreateCertificate(rand.Reader, template, parent, &clientKey.PublicKey, clientKey)
	require.NoError(t, err)

	crtPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certRaw})
	keyPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return
}
