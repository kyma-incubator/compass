package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

//go:generate mockery -name=CertificateResolver
type CertificateResolver interface {
	SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error)
	RevokeCertificate(ctx context.Context) (bool, error)
	GetCertificateSigningRequestInfo(ctx context.Context) (*gqlschema.CertificateSigningRequestInfo, error)
}

type certificateResolver struct {
}

func NewCertificateResolver() CertificateResolver {
	return &certificateResolver{}
}

func (r *certificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error) {
	panic("not implemented")
}
func (r *certificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}
func (r *certificateResolver) GetCertificateSigningRequestInfo(ctx context.Context) (*gqlschema.CertificateSigningRequestInfo, error) {
	panic("not implemented")
}
