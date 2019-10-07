package oauth20

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct {
	transactioner persistence.Transactioner
	systemAuthSvc systemauth.SystemAuthService
}

func NewResolver(transactioner persistence.Transactioner, systemAuthSvc systemauth.SystemAuthService) *Resolver {
	return &Resolver{transactioner: transactioner, systemAuthSvc: systemAuthSvc}
}

func (r *Resolver) GenerateClientCredentialsForRuntime(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	//TODO:
	panic("not implemented")
}
func (r *Resolver) GenerateClientCredentialsForApplication(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	//TODO:
	panic("not implemented")
}
func (r *Resolver) GenerateClientCredentialsForIntegrationSystem(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	//TODO:
	panic("not implemented")
}

func (r *Resolver) generateClientCredentials() {

}