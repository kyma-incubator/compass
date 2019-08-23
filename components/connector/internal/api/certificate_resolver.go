package api

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

//go:generate mockery -name=CertificateResolver
type CertificateResolver interface {
	SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error)
	RevokeCertificate(ctx context.Context) (bool, error)
	Configuration(ctx context.Context) (*externalschema.Configuration, error)
}

type certificateResolver struct {
	authenticator       authentication.Authenticator
	tokenService        tokens.Service
	certificatesService certificates.Service
	csrSubjectConsts    certificates.CSRSubjectConsts
	log                 *logrus.Entry
}

func NewCertificateResolver(authenticator authentication.Authenticator, tokenService tokens.Service, certificatesService certificates.Service, csrSubjectConsts certificates.CSRSubjectConsts) CertificateResolver {
	return &certificateResolver{
		authenticator:       authenticator,
		tokenService:        tokenService,
		certificatesService: certificatesService,
		csrSubjectConsts:    csrSubjectConsts,
		log:                 logrus.WithField("Resolver", "Certificate"),
	}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error) {
	commonName, err := r.authenticator.AuthenticateTokenOrCertificate(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to authenticate with token or certificate: %v", err)
	}

	rawCSR, err := decodeStringFromBase64(csr)
	if err != nil {
		return nil, fmt.Errorf("Error while decoding Certificate Signing Request: %v", err)
	}

	subject := certificates.CSRSubject{
		CommonName:       commonName,
		CSRSubjectConsts: r.csrSubjectConsts,
	}

	encodedCertificates, err := r.certificatesService.SignCSR(rawCSR, subject)
	if err != nil {
		return nil, fmt.Errorf("Error while signing Certificate Signing Request: %v", err)
	}

	certificationResult := certificates.ToCertificationResult(encodedCertificates)

	return &certificationResult, nil
}
func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}
func (r *certificateResolver) Configuration(ctx context.Context) (*externalschema.Configuration, error) {
	clientId, err := r.authenticator.AuthenticateTokenOrCertificate(ctx)
	if err != nil {
		r.log.Error(err.Error())
		return nil, err
	}

	r.log.Info("Fetching configuration for %s client.", clientId)

	// TODO

	return nil, nil
}

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("Error while parsing base64 content. Incorrect value provided.")
	}

	return bytes, nil
}
