package oauth20

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscorez
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
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *model.SystemAuth) *graphql.SystemAuth
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	CreateClient(ctx context.Context, objectType model.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error)
	DeleteClient(ctx context.Context, clientID string) error
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
	var err error
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var exists bool
	switch objType {
	case model.RuntimeReference:
		exists, err = r.rtmSvc.Exist(ctx, objID)
	case model.ApplicationReference:
		exists, err = r.appSvc.Exist(ctx, objID)
	case model.IntegrationSystemReference:
		exists, err = r.isSvc.Exist(ctx, objID)
	default:
		err = fmt.Errorf("invalid object type %s", objType)
	}

	if err != nil {
		return nil, errors.Wrap(err, "while checking if runtime exists")
	}
	if !exists {
		return nil, fmt.Errorf("%s with ID '%s' not found", objType, objID)
	}

	clientCreds, err := r.svc.CreateClient(ctx, objType)
	if err != nil {
		return nil, err
	}

	if clientCreds == nil {
		return nil, errors.New("client credentials cannot be empty")
	}

	cleanupOnFail := func() {
		err := r.svc.DeleteClient(ctx, clientCreds.ClientID)
		if err != nil {
			logrus.Error(errors.Wrap(err, "while deleting registered OAuth 2.0 Client on failure"))
		}
	}

	authInput := &model.AuthInput{
		Credential: &model.CredentialDataInput{
			Oauth: clientCreds,
		},
	}
	id := clientCreds.ClientID
	_, err = r.systemAuthSvc.CreateWithCustomID(ctx, id, objType, objID, authInput)
	if err != nil {
		cleanupOnFail()
		return nil, err
	}

	sysAuth, err := r.systemAuthSvc.Get(ctx, id)
	if err != nil {
		cleanupOnFail()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		cleanupOnFail()
		return nil, err
	}

	gqlSysAuth := r.systemAuthConv.ToGraphQL(sysAuth)
	return gqlSysAuth, nil
}
