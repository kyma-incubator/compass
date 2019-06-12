package gqlschema

import (
	"context"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) GenerateApplicationToken(ctx context.Context, appID string) (*Token, error) {
	panic("not implemented")
}
func (r *mutationResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*Token, error) {
	panic("not implemented")
}
func (r *mutationResolver) SignCertificateSigningRequest(ctx context.Context, csr string, token *string) (*CertificationResult, error) {
	panic("not implemented")
}
func (r *mutationResolver) RenewCertificate(ctx context.Context, csr string) (*CertificationResult, error) {
	panic("not implemented")
}
func (r *mutationResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetCertificateSignignRequestInfo(ctx context.Context, token *string) (*CertificateSigningRequestInfo, error) {
	panic("not implemented")
}
