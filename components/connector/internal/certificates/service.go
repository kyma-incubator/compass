package certificates

import (
	"crypto/x509"
	"encoding/base64"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
)

//go:generate mockery -name=Service
type Service interface {
	// SignCSR takes encoded CSR, validates subject and generates Certificate based on CA stored in secret
	// returns base64 encoded certificate chain
	SignCSR(encodedCSR []byte, subject CSRSubject, clientId string) (EncodedCertificateChain, apperrors.AppError)
}

type certificateService struct {
	certificateCache            Cache
	certUtil                    CertificateUtility
	caSecretName                string
	caCertificateSecretKey      string
	caKeySecretKey              string
	rootCACertificateSecretName string
	rootCACertificateSecretKey  string
	logger                      *logrus.Entry
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
		logger : logrus.WithField("CertificateService", "Certificate"),
	}
}

func (svc *certificateService) SignCSR(encodedCSR []byte, subject CSRSubject, clientId string) (EncodedCertificateChain, apperrors.AppError) {
	svc.logger.Infof("SignCSR for %s client started.", clientId)
	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	svc.logger.Infof("LoadCSR executed for %s client.", clientId)

	err = svc.checkCSR(csr, subject)
	if err != nil {
		logrus.Errorf("ERR failed to verify CSR for client %s: %s", clientId, err.Error())
		return EncodedCertificateChain{}, err
	}

	svc.logger.Infof("checkCSR executed for %s client.", clientId)

	return svc.signCSR(csr, clientId)
}

func (svc *certificateService) signCSR(csr *x509.CertificateRequest, clientId string) (EncodedCertificateChain, apperrors.AppError) {
	svc.logger.Infof("signCSR for %s client started. (1)", clientId)
	secretData, err := svc.certificateCache.Get(svc.caSecretName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	svc.logger.Infof("secret read for %s client. (2)", clientId)

	caCrt, err := svc.certUtil.LoadCert(secretData[svc.caCertificateSecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	svc.logger.Infof("cert loaded for %s client. (3)", clientId)

	caKey, err := svc.certUtil.LoadKey(secretData[svc.caKeySecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	svc.logger.Infof("key loaded for %s client. (4)", clientId)

	signedCrt, err := svc.certUtil.SignCSR(caCrt, csr, caKey)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	return svc.encodeCertificates(caCrt.Raw, signedCrt, clientId)
}

func (svc *certificateService) encodeCertificates(rawCaCertificate, rawClientCertificate []byte, clientId string) (EncodedCertificateChain, apperrors.AppError) {
	svc.logger.Infof("encodeCertificates for %s client started. (5)", clientId)

	caCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawCaCertificate)

	svc.logger.Infof("certificate header and footer added for client %s. (ca certificate) (6)", clientId)

	signedCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawClientCertificate)

	svc.logger.Infof("certificate header and footer added for client %s. (client certificate) (7)", clientId)
	if svc.rootCACertificateSecretName != "" && svc.rootCACertificateSecretKey != "" {
		rootCABytes, err := svc.loadRootCACert(clientId)
		if err != nil {
			svc.logger.Errorf("ERR failed to load root cert for client %s: %s", clientId, err.Error())
			return EncodedCertificateChain{}, err
		}
		svc.logger.Infof("root CA cert loaded for client %s. (8)", clientId)

		caCrtBytes = append(caCrtBytes, rootCABytes...)
		svc.logger.Infof("cert chain created for client %s. (root chain) (9)", clientId)
	}

	certChain := append(signedCrtBytes, caCrtBytes...)
	svc.logger.Infof("cert chain created for client %s. (client chain) (10)", clientId)

	return encodeCertificateBase64(certChain, signedCrtBytes, caCrtBytes), nil
}

func (svc *certificateService) loadRootCACert(clientId string) ([]byte, apperrors.AppError) {
	svc.logger.Infof("loadRootCACert for %s client started.", clientId)
	secretData, err := svc.certificateCache.Get(svc.rootCACertificateSecretName)
	if err != nil {
		return nil, err
	}

	svc.logger.Infof("secret for Root CA read for %s client.", clientId)
	rootCACrt, err := svc.certUtil.LoadCert(secretData[svc.rootCACertificateSecretKey])
	if err != nil {
		return nil, err
	}

	svc.logger.Infof("cert loaded for Root CA for %s client.", clientId)

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
