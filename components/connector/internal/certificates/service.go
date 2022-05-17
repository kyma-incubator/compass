package certificates

import (
	"context"
	"crypto/x509"
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
)

//go:generate mockery --name=Service --disable-version-string
type Service interface {
	// SignCSR takes encoded CSR, validates subject and generates Certificate based on CA stored in secret
	// returns base64 encoded certificate chain
	SignCSR(ctx context.Context, encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError)
}

type certificateService struct {
	certsCache           Cache
	certUtil             CertificateUtility
	caCertSecretName     string
	caCertSecretKey      string
	caKeySecretKey       string
	rootCACertSecretName string
	rootCACertSecretKey  string
}

func NewCertificateService(
	certsCache Cache,
	certUtil CertificateUtility,
	caCertSecretName, rootCACertSecretName string,
	caCertSecretKey, caKeySecretKey, rootCACertSecretKey string) Service {

	return &certificateService{
		certsCache:           certsCache,
		certUtil:             certUtil,
		caCertSecretName:     caCertSecretName,
		caCertSecretKey:      caCertSecretKey,
		caKeySecretKey:       caKeySecretKey,
		rootCACertSecretName: rootCACertSecretName,
		rootCACertSecretKey:  rootCACertSecretKey,
	}
}

func (svc *certificateService) SignCSR(ctx context.Context, encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError) {
	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while loading the CSR with Common Name %s: %v", subject.CommonName, err)
		return EncodedCertificateChain{}, err
	}
	log.C(ctx).Debugf("Successfully loaded the CSR with Common Name %s", subject.CommonName)

	err = svc.checkCSR(csr, subject)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while checking the values of the CSR with Common Name %s: %v", subject.CommonName, err)
		return EncodedCertificateChain{}, err
	}
	log.C(ctx).Debugf("Successfully checked the values of the CSR with Common Name %s", subject.CommonName)

	encodedCertChain, err := svc.signCSR(csr)
	if err != nil {
		return EncodedCertificateChain{}, err
	}
	log.C(ctx).Debugf("Successfully signed CSR with Common Name %s", subject.CommonName)

	return encodedCertChain, nil
}

func (svc *certificateService) signCSR(csr *x509.CertificateRequest) (EncodedCertificateChain, apperrors.AppError) {
	secretData, err := svc.certsCache.Get(svc.caCertSecretName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caCrt, err := svc.certUtil.LoadCert(secretData[svc.caCertSecretKey])
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

	if svc.rootCACertSecretName != "" && svc.rootCACertSecretKey != "" {
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
	secretData, err := svc.certsCache.Get(svc.rootCACertSecretName)
	if err != nil {
		return nil, err
	}

	rootCACrt, err := svc.certUtil.LoadCert(secretData[svc.rootCACertSecretKey])
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
