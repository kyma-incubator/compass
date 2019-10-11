package oauth20

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	CreateWithCustomID(ctx context.Context, id string, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *model.AuthInput) (string, error)
	Get(ctx context.Context, id string) (*model.SystemAuth, error)
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=IntegrationSystemService -output=automock -outpkg=automock -case=underscore
type IntegrationSystemService interface {
	Exists(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *model.SystemAuth) *graphql.SystemAuth
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	CreateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error)
	DeleteClientCredentials(ctx context.Context, clientID string) error
}

type Resolver struct {
	transact       persistence.Transactioner
	svc            Service
	systemAuthSvc  SystemAuthService
	systemAuthConv SystemAuthConverter
	appSvc         ApplicationService
	rtmSvc         RuntimeService
	isSvc          IntegrationSystemService
}

func NewResolver(transactioner persistence.Transactioner, svc Service, appSvc ApplicationService, rtmSvc RuntimeService, isSvc IntegrationSystemService, systemAuthSvc SystemAuthService, systemAuthConv SystemAuthConverter) *Resolver {
	return &Resolver{transact: transactioner, svc: svc, appSvc: appSvc, rtmSvc: rtmSvc, systemAuthSvc: systemAuthSvc, isSvc: isSvc, systemAuthConv: systemAuthConv}
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

	exists, err := r.checkObjectExist(ctx, objType, objID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking if %s with ID '%s' exists", objType, objID)
	}
	if !exists {
		return nil, fmt.Errorf("%s with ID '%s' not found", objType, objID)
	}

	clientCreds, err := r.svc.CreateClientCredentials(ctx, objType)
	if err != nil {
		return nil, err
	}
	if clientCreds == nil {
		return nil, errors.New("client credentials cannot be empty")
	}
	cleanupOnError := func(originalErr error) error {
		cleanupErr := r.svc.DeleteClientCredentials(ctx, clientCreds.ClientID)
		if cleanupErr != nil {
			return multierror.Append(err, cleanupErr)
		}

		return originalErr
	}

	id := clientCreds.ClientID
	_, err = r.systemAuthSvc.CreateWithCustomID(ctx, id, objType, objID, &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Oauth: clientCreds,
		},
	})
	if err != nil {
		finalErr := cleanupOnError(err)
		return nil, finalErr
	}

	sysAuth, err := r.systemAuthSvc.Get(ctx, id)
	if err != nil {
		finalErr := cleanupOnError(err)
		return nil, finalErr
	}

	err = tx.Commit()
	if err != nil {
		finalErr := cleanupOnError(err)
		return nil, finalErr
	}

	gqlSysAuth := r.systemAuthConv.ToGraphQL(sysAuth)
	return gqlSysAuth, nil
}

func (r *Resolver) checkObjectExist(ctx context.Context, objType model.SystemAuthReferenceObjectType, objID string) (bool, error) {
	switch objType {
	case model.RuntimeReference:
		return r.rtmSvc.Exist(ctx, objID)
	case model.ApplicationReference:
		return r.appSvc.Exist(ctx, objID)
	case model.IntegrationSystemReference:
		return r.isSvc.Exists(ctx, objID)
	}

	return false, fmt.Errorf("invalid object type %s", objType)
}
