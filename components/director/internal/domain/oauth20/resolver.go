package oauth20

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscorez
type SystemAuthService interface {
	Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
	Get(ctx context.Context, id string) (*model.SystemAuth, error)
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *model.SystemAuth) *graphql.SystemAuth
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	CreateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) (*model.OAuthCredentialDataInput, error)
}

type Resolver struct {
	transact  persistence.Transactioner
	svc            Service
	systemAuthSvc  SystemAuthService
	systemAuthConv SystemAuthConverter
}

func NewResolver(transactioner persistence.Transactioner, systemAuthSvc SystemAuthService) *Resolver {
	return &Resolver{transact: transactioner, systemAuthSvc: systemAuthSvc}
}

func (r *Resolver) GenerateClientCredentialsForRuntime(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.generateClientCredentials(ctx, model.RuntimeReference, id)
}

func (r *Resolver) GenerateClientCredentialsForApplication(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.generateClientCredentials(ctx, model.ApplicationReference, id)
}

func (r *Resolver) GenerateClientCredentialsForIntegrationSystem(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.generateClientCredentials(ctx, model.IntegrationSystemReference, id)
}

func (r *Resolver) generateClientCredentials(ctx context.Context, objType model.SystemAuthReferenceObjectType, objID string) (*graphql.SystemAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	clientCreds, err := r.svc.CreateClientCredentials(ctx, objType, objID)

	authInput := &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Oauth: clientCreds,
		},
	}

	id, err := r.systemAuthSvc.Create(ctx, objType, objID, authInput)
	if err != nil {
		return nil, err
	}

	sysAuth, err := r.systemAuthSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlSysAuth := r.systemAuthConv.ToGraphQL(sysAuth)
	return gqlSysAuth, nil
}
