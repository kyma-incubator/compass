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

type Cache interface {
	Put(secretData map[string][]byte)
	Get() (*tls.Certificate, error)
}

type certificatesCache struct {
	secretName string
	cache      *cache.Cache
}

func NewCertificateCache(secretName string) Cache {
	return &certificatesCache{
		secretName: secretName,
		cache:      cache.New(0, 0),
	}
}

func (cc *certificatesCache) Put(secretData map[string][]byte) {
	cc.cache.Set(cc.secretName, secretData, 0)
}

func (cc *certificatesCache) Get() (*tls.Certificate, error) {
	secretData, err := cc.getSecretDataFromCache()
	if err != nil {
		log.D().Error(errors.Wrap(err, "Failed to get client certificate data from cache"))
		return nil, errors.Wrap(err, "Failed to get client certificate data from cache")
	}

	certBytes := secretData["tls.crt"]
	clientKeyBytes := secretData["tls.key"]

	clientCrtPem, _ := pem.Decode(certBytes)
	if clientCrtPem == nil {
		return nil, errors.New("Error while decoding certificate pem block.")
	}

	clientCert, err := x509.ParseCertificate(clientCrtPem.Bytes)
	if err != nil {
		return nil, err
	}

	clientKeyPem, _ := pem.Decode(clientKeyBytes)
	if clientKeyPem == nil {
		return nil, errors.New("Error while decoding private ket pem block.")
	}

	if pkcs1PrivateKey, err := x509.ParsePKCS1PrivateKey(clientKeyPem.Bytes); err == nil {
		return &tls.Certificate{
			Certificate: [][]byte{clientCert.Raw},
			PrivateKey:  pkcs1PrivateKey,
		}, nil
	}

	pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(clientKeyPem.Bytes)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{clientCert.Raw},
		PrivateKey:  pkcs8PrivateKey.(*rsa.PrivateKey),
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
