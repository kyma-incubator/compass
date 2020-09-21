package api

import (
	"context"
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CertificateResolver interface {
	SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error)
	RevokeCertificate(ctx context.Context) (bool, error)
	Configuration(ctx context.Context) (*externalschema.Configuration, error)
}

type certificateResolver struct {
	authenticator                  authentication.Authenticator
	tokenService                   tokens.Service
	certificatesService            certificates.Service
	csrSubjectConsts               certificates.CSRSubjectConsts
	directorURL                    string
	certificateSecuredConnectorURL string
	revocationList                 revocation.RevocationListRepository
	log                            *logrus.Entry
}

func NewCertificateResolver(
	authenticator authentication.Authenticator,
	tokenService tokens.Service,
	certificatesService certificates.Service,
	csrSubjectConsts certificates.CSRSubjectConsts,
	directorURL string,
	certificateSecuredConnectorURL string,
	revocationList revocation.RevocationListRepository) CertificateResolver {
	return &certificateResolver{
		authenticator:                  authenticator,
		tokenService:                   tokenService,
		certificatesService:            certificatesService,
		csrSubjectConsts:               csrSubjectConsts,
		directorURL:                    directorURL,
		certificateSecuredConnectorURL: certificateSecuredConnectorURL,
		revocationList:                 revocationList,
		log:                            logrus.WithField("Resolver", "Certificate"),
	}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error) {
	r.log.Debug("Authenticating the call for signing the Certificate Signing Request.")

	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		r.log.Error("Failed authentication during the signing CSR process. ", err.Error())
		return nil, errors.Wrap(err, "Failed to authenticate with token")
	}

	r.log.Infof("Signing Certificate Signing Request for client with id %s", clientId)

	rawCSR, err := decodeStringFromBase64(csr)
	if err != nil {
		r.log.Errorf("Failed to decode the input CSR of client with id %s during the certificate signing process. %s", clientId, err.Error())
		return nil, errors.Wrap(err, "Error while decoding Certificate Signing Request")
	}

	subject := certificates.CSRSubject{
		CommonName:       clientId,
		CSRSubjectConsts: r.csrSubjectConsts,
	}

	encodedCertificates, err := r.certificatesService.SignCSR(rawCSR, subject)
	if err != nil {
		r.log.Errorf("Error occurred while signing the CSR with Common Name %s of client with id %s : %s ", subject.CommonName, clientId, err.Error())
		return nil, errors.Wrap(err, "Error while signing Certificate Signing Request")
	}

	certificationResult := certificates.ToCertificationResult(encodedCertificates)

	r.log.Infof("Certificate Signing Request with Common Name %s of client with id %s successfully signed.", subject.CommonName, clientId)
	return &certificationResult, nil
}

func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	r.log.Debug("Authenticating the call for certificate revocation.")

	clientId, certificateHash, err := r.authenticator.AuthenticateCertificate(ctx)
	if err != nil {
		r.log.Error("Failed authentication while revoking the certificate. ", err.Error())
		return false, errors.Wrap(err, "Failed to authenticate with certificate")
	}

	r.log.Infof("Revoking certificate for client with id %s", clientId)

	r.log.Debugf("Inserting certificate hash of client with id %s to revocation list", clientId)
	err = r.revocationList.Insert(certificateHash)
	if err != nil {
		r.log.Errorf("Failed to add certificate hash of client with id %s to revocation list. %s ", clientId, err.Error())
		return false, errors.Wrap(err, "Failed to add hash to revocation list")
	}

	r.log.Infof("Certificate of client with id %s successfully revoked.", clientId)
	return true, nil
}

func (r *certificateResolver) Configuration(ctx context.Context) (*externalschema.Configuration, error) {
	r.log.Debug("Authenticating the call for configuration fetching.")

	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		r.log.Error("Failed authentication while fetching the configuration. ", err.Error())
		return nil, err
	}
	r.log.Infof("Fetching configuration for client with id %s", clientId)

	r.log.Infof("Creating one-time token as part of fetching configuration process for client with id %s", clientId)
	token, err := r.tokenService.CreateToken(clientId, tokens.CSRToken)
	if err != nil {
		r.log.Errorf("Error occurred while creating one-time token for client with id %s during fetching configuration process: %s", clientId, err.Error())
		return nil, errors.Wrap(err, "Failed to create one-time token during fetching configuration process")
	}

	csrInfo := &externalschema.CertificateSigningRequestInfo{
		Subject:      r.csrSubjectConsts.ToString(clientId),
		KeyAlgorithm: "rsa2048",
	}

	r.log.Infof("Configuration for client with id %s successfully fetched.", clientId)

	return &externalschema.Configuration{
		Token:                         &externalschema.Token{Token: token},
		CertificateSigningRequestInfo: csrInfo,
		ManagementPlaneInfo: &externalschema.ManagementPlaneInfo{
			DirectorURL:                    &r.directorURL,
			CertificateSecuredConnectorURL: &r.certificateSecuredConnectorURL,
		},
	}, nil
}

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("Error while parsing base64 content. Incorrect value provided.")
	}

	return bytes, nil
}
