package certloader

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// Cache is used to interact with the cache
type Cache interface {
	Put(secretData map[string][]byte)
	Get() (*tls.Certificate, error)
}

type certificatesCache struct {
	secretName string
	cache      *cache.Cache
}

// NewCertificateCache creates a cache that will store a client certificate
func NewCertificateCache(secretName string) Cache {
	return &certificatesCache{
		secretName: secretName,
		cache:      cache.New(0, 0),
	}
}

// Put inserts secret data containing client certificate and private key in the cache
func (cc *certificatesCache) Put(secretData map[string][]byte) {
	cc.cache.Set(cc.secretName, secretData, 0)
}

// Get returns a parsed certificate, build from the cache or error
func (cc *certificatesCache) Get() (*tls.Certificate, error) {
	secretData, err := cc.getSecretDataFromCache()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get client certificate data from cache")
	}

	certBytes := secretData["tls.crt"]
	privateKeyBytes := secretData["tls.key"]

	if certBytes == nil || privateKeyBytes == nil {
		return nil, errors.New("There is no certificate data in the cache")
	}

	clientCrtPem, _ := pem.Decode(certBytes)
	if clientCrtPem == nil {
		return nil, errors.New("Error while decoding certificate pem block")
	}

	clientCert, err := x509.ParseCertificate(clientCrtPem.Bytes)
	if err != nil {
		return nil, err
	}

	privateKeyPem, _ := pem.Decode(privateKeyBytes)
	if privateKeyPem == nil {
		return nil, errors.New("Error while decoding private key pem block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPem.Bytes)
	if err != nil {
		pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyPem.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = pkcs8PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, err
		}
	}

	return &tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  privateKey,
	}, nil
}

func (cc *certificatesCache) getSecretDataFromCache() (map[string][]byte, error) {
	log.D().Debugf("Getting client certificate data from cache with key: %s", cc.secretName)
	data, found := cc.cache.Get(cc.secretName)
	if !found {
		return nil, errors.New(fmt.Sprintf("Client certificate data not found in the cache for key: %s", cc.secretName))
	}

	certData, ok := data.(map[string][]byte)
	if !ok {
		return nil, errors.New("certificate cache did not have the expected secret type")
	}

	return certData, nil
}
