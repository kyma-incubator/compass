package certloader

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
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

func Test_CertificateLoaderWatch(t *testing.T) {
	config := Config{
		ExternalClientCertSecret:  "namespace/resource-name",
		ExternalClientCertCertKey: "tls.crt",
		ExternalClientCertKeyKey:  "tls.key",
	}

	t.Run("should insert secret data on add event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, secretManagerMock := preparation(ctx, 1, config)
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
			tlsCert := cache.Get()
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
		cache, watcher, secretManagerMock := preparation(ctx, 1, config)
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
			tlsCert := cache.Get()
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
		cache, watcher, secretManagerMock := preparation(ctx, 1, config)

		// when
		watcher.putEvent(watch.Event{
			Type:   watch.Added,
			Object: &runtime.Unknown{},
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert := cache.Get()
			require.Nil(t, tlsCert[0])
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
		cache, watcher, secretManagerMock := preparation(ctx, 1, config)
		certBytes, keyBytes := generateTestCertAndKey(t, testCN)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.Secret{
				Data: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
			},
		})

		assert.Eventually(t, func() bool {
			tlsCert := cache.Get()
			require.NotNil(t, tlsCert[0])
			return true
		}, 2*time.Second, 100*time.Millisecond)

		watcher.putEvent(watch.Event{
			Type: watch.Deleted,
		})

		// then
		assert.Eventually(t, func() bool {
			tlsCert := cache.Get()
			require.Nil(t, tlsCert[0])
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
		cache, watcher, secretManagerMock := preparation(ctx, 2, config)
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
			tlsCert := cache.Get()
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
		cache, watcher, secretManagerMock := preparation(ctx, 2, config)
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
			tlsCert := cache.Get()
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

func Test_CertificateParsing(t *testing.T) {
	ctx := context.Background()
	config := Config{
		ExternalClientCertCertKey: "tls.crt",
		ExternalClientCertKeyKey:  "tls.key",
	}
	certBytes, keyBytes := generateTestCertAndKey(t, testCN)
	invalidCert := "-----BEGIN CERTIFICATE-----\naZOCUHlJ1wKwnYiLnOofB1xyIUZhVLaJy7Ob\n-----END CERTIFICATE-----\n"
	invalidKey := "-----BEGIN RSA PRIVATE KEY-----\n7qFmWkbkOAM9CUPx5RwSRt45oxlQjvDniZALWqbYxgO5f8cYZsEAyOU1n2DXgiei\n-----END RSA PRIVATE KEY-----\n"

	testCases := []struct {
		Name             string
		SecretData       map[string][]byte
		ExpectedErrorMsg string
	}{
		{
			Name:       "Successfully get certificate from cache",
			SecretData: map[string][]byte{secretCertKey: certBytes, secretKeyKey: keyBytes},
		},
		{
			Name:             "Error when secret data is empty",
			SecretData:       map[string][]byte{},
			ExpectedErrorMsg: "There is no certificate data provided",
		},
		{
			Name:             "Error when certificate data is invalid",
			SecretData:       map[string][]byte{secretCertKey: []byte("invalid"), secretKeyKey: []byte("invalid")},
			ExpectedErrorMsg: "Error while decoding certificate pem block",
		},
		{
			Name:             "Error when parsing certificate",
			SecretData:       map[string][]byte{secretCertKey: []byte(invalidCert), secretKeyKey: []byte("invalid")},
			ExpectedErrorMsg: "malformed certificate",
		},
		{
			Name:             "Error when private key is invalid",
			SecretData:       map[string][]byte{secretCertKey: certBytes, secretKeyKey: []byte("invalid")},
			ExpectedErrorMsg: "Error while decoding private key pem block",
		},
		{
			Name:             "Error when parsing private key",
			SecretData:       map[string][]byte{secretCertKey: certBytes, secretKeyKey: []byte(invalidKey)},
			ExpectedErrorMsg: "structure error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tlsCert, err := parseCertificate(ctx, testCase.SecretData, config)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				require.Nil(t, tlsCert)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tlsCert)
			}
		})
	}
}

func preparation(ctx context.Context, number int, config Config) (Cache, *testWatch, *automock.Manager) {
	cache := NewCertificateCache()
	watcher := &testWatch{
		events: make(chan watch.Event, 50),
	}
	secretManagerMock := &automock.Manager{}
	secretManagerMock.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).Return(watcher, nil).Times(number)
	loader := NewCertificateLoader(config, cache, []Manager{secretManagerMock}, []string{secretName}, time.Millisecond)
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
