package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type Resolver struct {
}

type mutationResolver struct {
	*Resolver
	TokenResolver
	CertificateResolver
}

type queryResolver struct {
	*Resolver
	CertificateResolver
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r, TokenResolver{}, CertificateResolver{}}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r, CertificateResolver{}}
}

type TokenResolver struct {
}

func (r *TokenResolver) GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
func (r *TokenResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error) {
	panic("not implemented")
}

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
