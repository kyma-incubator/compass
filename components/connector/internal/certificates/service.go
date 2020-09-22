package certificates

import (
	"crypto/x509"
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=Service
type Service interface {
	// SignCSR takes encoded CSR, validates subject and generates Certificate based on CA stored in secret
	// returns base64 encoded certificate chain
	SignCSR(encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError)
}

type certificateService struct {
	certificateCache            Cache
	certUtil                    CertificateUtility
	caSecretName                string
	caCertificateSecretKey      string
	caKeySecretKey              string
	rootCACertificateSecretName string
	rootCACertificateSecretKey  string
}

func NewCertificateService(
	certificateCache Cache,
	certUtil CertificateUtility,
	caSecretName, rootCACertificateSecretName string,
	caCertificateSecretKey, caKeySecretKey, rootCACertificateSecretKey string) Service {

	return &certificateService{
		certificateCache:            certificateCache,
		certUtil:                    certUtil,
		caSecretName:                caSecretName,
		caCertificateSecretKey:      caCertificateSecretKey,
		caKeySecretKey:              caKeySecretKey,
		rootCACertificateSecretName: rootCACertificateSecretName,
		rootCACertificateSecretKey:  rootCACertificateSecretKey,
	}
}

func (svc *certificateService) SignCSR(encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError) {
	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		log.Errorf("Error occurred while loading the CSR with Common Name %s", subject.CommonName)
		return EncodedCertificateChain{}, err
	}
	log.Debugf("Successfully loaded the CSR with Common Name %s", subject.CommonName)

	err = svc.checkCSR(csr, subject)
	if err != nil {
		log.Errorf("Error occurred while checking the values of the CSR with Common Name %s", subject.CommonName)
		return EncodedCertificateChain{}, err
	}
	log.Debugf("Successfully checked the values of the CSR with Common Name %s", subject.CommonName)

	encodedCertChain, err := svc.signCSR(csr)
	if err != nil {
		return EncodedCertificateChain{}, err
	}
	log.Debugf("Successfully signed CSR with Common Name %s", subject.CommonName)

	return encodedCertChain, nil
}

func (svc *certificateService) signCSR(csr *x509.CertificateRequest) (EncodedCertificateChain, apperrors.AppError) {
	secretData, err := svc.certificateCache.Get(svc.caSecretName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caCrt, err := svc.certUtil.LoadCert(secretData[svc.caCertificateSecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caKey, err := svc.certUtil.LoadKey(secretData[svc.caKeySecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	signedCrt, err := svc.certUtil.SignCSR(caCrt, csr, caKey)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	return svc.encodeCertificates(caCrt.Raw, signedCrt)
}

func (svc *certificateService) encodeCertificates(rawCaCertificate, rawClientCertificate []byte) (EncodedCertificateChain, apperrors.AppError) {
	caCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawCaCertificate)
	signedCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawClientCertificate)

	if svc.rootCACertificateSecretName != "" && svc.rootCACertificateSecretKey != "" {
		rootCABytes, err := svc.loadRootCACert()
		if err != nil {
			return EncodedCertificateChain{}, err
		}

		caCrtBytes = append(caCrtBytes, rootCABytes...)
	}

	certChain := append(signedCrtBytes, caCrtBytes...)

	return encodeCertificateBase64(certChain, signedCrtBytes, caCrtBytes), nil
}

func (svc *certificateService) loadRootCACert() ([]byte, apperrors.AppError) {
	secretData, err := svc.certificateCache.Get(svc.rootCACertificateSecretName)
	if err != nil {
		return nil, err
	}

	rootCACrt, err := svc.certUtil.LoadCert(secretData[svc.rootCACertificateSecretKey])
	if err != nil {
		return nil, err
	}

	return svc.certUtil.AddCertificateHeaderAndFooter(rootCACrt.Raw), nil
}

func (svc *certificateService) checkCSR(csr *x509.CertificateRequest, expectedSubject CSRSubject) apperrors.AppError {
	return svc.certUtil.CheckCSRValues(csr, expectedSubject)
}

func encodeCertificateBase64(certChain, clientCRT, caCRT []byte) EncodedCertificateChain {
	return EncodedCertificateChain{
		CertificateChain:  encodeStringBase64(certChain),
		ClientCertificate: encodeStringBase64(clientCRT),
		CaCertificate:     encodeStringBase64(caCRT),
	}
}

func encodeStringBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
