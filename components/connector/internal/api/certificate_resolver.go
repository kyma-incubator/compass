package api

import (
	"context"
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type CertificateResolver interface {
	SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error)
	RevokeCertificate(ctx context.Context) (bool, error)
	Configuration(ctx context.Context) (*gqlschema.Configuration, error)
}

type certificateResolver struct {
	authenticator                  authentication.Authenticator
	tokenService                   tokens.Service
	certificatesService            certificates.Service
	csrSubjectConsts               certificates.CSRSubjectConsts
	directorURL                    string
	certificateSecuredConnectorURL string
	log                            *logrus.Entry
}

func NewCertificateResolver(
	authenticator authentication.Authenticator,
	tokenService tokens.Service,
	certificatesService certificates.Service,
	csrSubjectConsts certificates.CSRSubjectConsts,
	directorURL string,
	certificateSecuredConnectorURL string) CertificateResolver {
	return &certificateResolver{
		authenticator:                  authenticator,
		tokenService:                   tokenService,
		certificatesService:            certificatesService,
		csrSubjectConsts:               csrSubjectConsts,
		directorURL:                    directorURL,
		certificateSecuredConnectorURL: certificateSecuredConnectorURL,
		log:                            logrus.WithField("Resolver", "Certificate"),
	}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error) {
	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		r.log.Errorf(err.Error())
		return nil, errors.Wrap(err, "Failed to authenticate with token")
	}

	r.log.Infof("Signing Certificate Signing Request for %s client.", clientId)

	rawCSR, err := decodeStringFromBase64(csr)
	if err != nil {
		r.log.Errorf(err.Error())
		return nil, errors.Wrap(err, "Error while decoding Certificate Signing Request")
	}

	subject := certificates.CSRSubject{
		CommonName:       clientId,
		CSRSubjectConsts: r.csrSubjectConsts,
	}

	encodedCertificates, err := r.certificatesService.SignCSR(rawCSR, subject)
	if err != nil {
		r.log.Errorf(err.Error())
		return nil, errors.Wrap(err, "Error while signing Certificate Signing Request")
	}

	certificationResult := certificates.ToCertificationResult(encodedCertificates)

	r.log.Infof("Certificate Signing Request signed.")
	return &certificationResult, nil
}

func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}

func (r *certificateResolver) Configuration(ctx context.Context) (*gqlschema.Configuration, error) {
	clientId, err := r.authenticator.Authenticate(ctx)
	if err != nil {
		r.log.Errorf(err.Error())
		return nil, err
	}

	r.log.Infof("Fetching configuration for %s client.", clientId)

	token, err := r.tokenService.CreateToken(clientId, tokens.CSRToken)
	if err != nil {
		r.log.Errorf(err.Error())
		return nil, err
	}

	csrInfo := &gqlschema.CertificateSigningRequestInfo{
		Subject:      r.csrSubjectConsts.ToString(clientId),
		KeyAlgorithm: "rsa2048",
	}

	return &gqlschema.Configuration{
		Token:                         &gqlschema.Token{Token: token},
		CertificateSigningRequestInfo: csrInfo,
		ManagementPlaneInfo: &gqlschema.ManagementPlaneInfo{
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
