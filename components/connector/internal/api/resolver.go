package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type Resolver struct{}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) GenerateApplicationToken(ctx context.Context, appID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
func (r *mutationResolver) GenerateRuntimeToken(ctx context.Context, runtimeID string) (*gqlschema.Token, error) {
	panic("not implemented")
}
func (r *mutationResolver) SignCertificateSigningRequest(ctx context.Context, csr string) (*gqlschema.CertificationResult, error) {
	panic("not implemented")
}
func (r *mutationResolver) RevokeCertificate(ctx context.Context) (bool, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetCertificateSignignRequestInfo(ctx context.Context) (*gqlschema.CertificateSigningRequestInfo, error) {
	panic("not implemented")
}
