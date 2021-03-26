package api

import (
	"context"
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/pkg/errors"
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
	revokedCertsRepository         revocation.RevokedCertificatesRepository
}

func NewCertificateResolver(
	authenticator authentication.Authenticator,
	tokenService tokens.Service,
	certificatesService certificates.Service,
	csrSubjectConsts certificates.CSRSubjectConsts,
	directorURL string,
	certificateSecuredConnectorURL string,
	revokedCertsRepository revocation.RevokedCertificatesRepository) CertificateResolver {
	return &certificateResolver{
		authenticator:                  authenticator,
		tokenService:                   tokenService,
		certificatesService:            certificatesService,
		csrSubjectConsts:               csrSubjectConsts,
		directorURL:                    directorURL,
		certificateSecuredConnectorURL: certificateSecuredConnectorURL,
		revokedCertsRepository:         revokedCertsRepository,
	}
}

func (r *certificateResolver) Configuration(ctx context.Context) (*externalschema.Configuration, error) {
	log.C(ctx).Debug("Authenticating the call for configuration fetching.")

	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed authentication while fetching the configuration: %v", err)
		return nil, err
	}
	log.C(ctx).Infof("Fetching configuration for client with id %s", clientId)

	log.C(ctx).Infof("Getting one-time token as part of fetching configuration process for client with id %s", clientId)
	token, err := r.tokenService.GetToken(ctx, clientId)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while getting one-time token for client with id %s during fetching configuration process: %v", clientId, err)
		return nil, errors.Wrap(err, "Failed to get one-time token during fetching configuration process")
	}

	csrInfo := &externalschema.CertificateSigningRequestInfo{
		Subject:      r.csrSubjectConsts.ToString(clientId),
		KeyAlgorithm: "rsa2048",
	}

	log.C(ctx).Infof("Configuration for client with id %s successfully fetched.", clientId)

	return &externalschema.Configuration{
		Token:                         &externalschema.Token{Token: token},
		CertificateSigningRequestInfo: csrInfo,
		ManagementPlaneInfo: &externalschema.ManagementPlaneInfo{
			DirectorURL:                    &r.directorURL,
			CertificateSecuredConnectorURL: &r.certificateSecuredConnectorURL,
		},
	}, nil
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error) {
	log.C(ctx).Debug("Authenticating the call for signing the Certificate Signing Request.")

	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed authentication during the signing CSR process: %v", err)
		return nil, errors.Wrap(err, "Failed to authenticate with token")
	}

	log.C(ctx).Infof("Signing Certificate Signing Request for client with id %s", clientId)

	rawCSR, err := decodeStringFromBase64(csr)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to decode the input CSR of client with id %s during the certificate signing process: %v", clientId, err)
		return nil, errors.Wrap(err, "Error while decoding Certificate Signing Request")
	}

	subject := certificates.CSRSubject{
		CommonName:       clientId,
		CSRSubjectConsts: r.csrSubjectConsts,
	}

	encodedCertificates, err := r.certificatesService.SignCSR(ctx, rawCSR, subject)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error occurred while signing the CSR with Common Name %s of client with id %s: %v", subject.CommonName, clientId, err)
		return nil, errors.Wrap(err, "Error while signing Certificate Signing Request")
	}

	certificationResult := certificates.ToCertificationResult(encodedCertificates)

	log.C(ctx).Infof("Certificate Signing Request with Common Name %s of client with id %s successfully signed.", subject.CommonName, clientId)
	return &certificationResult, nil
}

func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	log.C(ctx).Debug("Authenticating the call for certificate revocation.")

	clientId, certificateHash, err := r.authenticator.AuthenticateCertificate(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed authentication while revoking the certificate: %v", err)
		return false, errors.Wrap(err, "Failed to authenticate with certificate")
	}

	log.C(ctx).Infof("Revoking certificate for client with id %s", clientId)

	log.C(ctx).Debugf("Inserting certificate hash of client with id %s to revocation list", clientId)
	err = r.revokedCertsRepository.Insert(certificateHash)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to add certificate hash of client with id %s to revocation list: %v", clientId, err)
		return false, errors.Wrap(err, "Failed to add hash to revocation list")
	}

	log.C(ctx).Infof("Certificate of client with id %s successfully revoked.", clientId)
	return true, nil
}

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("Error while parsing base64 content. Incorrect value provided.")
	}

	return bytes, nil
}
