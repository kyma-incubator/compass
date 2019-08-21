package api

import (
	"context"
	"encoding/base64"
	"fmt"

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
}

func NewCertificateResolver(authenticator authentication.Authenticator, tokenService tokens.Service, certificatesService certificates.Service) CertificateResolver {
	return &certificateResolver{
		authenticator:       authenticator,
		tokenService:        tokenService,
		certificatesService: certificatesService,
	}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error) {

	// TODO: Check if there is valid token, valid cert or both. Subject should be extracted.
	subject := certificates.CSRSubject{
		CommonName:         "commonname",
		Country:            "country",
		Organization:       "organization",
		OrganizationalUnit: "organizationalunit",
		Locality:           "locality",
		Province:           "province",
	}

	rawCSR, err := decodeStringFromBase64(csr)
	if err != nil {
		return nil, fmt.Errorf("Error while decoding Certificate Signing Request: %v", err)
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
	panic("not implemented")
}

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("Error while parsing base64 content. Incorrect value provided.")
	}

	return bytes, nil
}
