package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type CertificateResolver struct {
}

func (r *CertificateResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error) {
	panic("not implemented")
}
func (r *CertificateResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}
func (r *CertificateResolver) GetCertificateSignignRequestInfo(ctx context.Context) (*gqlschema.CertificateSigningRequestInfo, error) {
	panic("not implemented")
}
