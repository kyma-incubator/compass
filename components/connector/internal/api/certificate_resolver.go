package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
)

//go:generate mockery -name=CertificateResolver
type CertificateResolver interface {
	SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error)
	RevokeCertificate(ctx context.Context) (bool, error)
	Configuration(ctx context.Context) (*externalschema.Configuration, error)
}

type certificateResolver struct {
	authenticator authentication.Authenticator
	tokenService  tokens.Service
}

func NewCertificateResolver(authenticator authentication.Authenticator, tokenService tokens.Service) CertificateResolver {
	return &certificateResolver{
		authenticator: authenticator,
		tokenService:  tokenService,
	}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*externalschema.CertificationResult, error) {
	panic("not implemented")
}
func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}
func (r *certificateResolver) Configuration(ctx context.Context) (*externalschema.Configuration, error) {
	panic("not implemented")
}
