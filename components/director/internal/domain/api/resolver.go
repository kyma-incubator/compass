package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/mock"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/pkg/errors"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error)
	GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	Get(ctx context.Context, id string) (*model.Runtime, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) *graphql.FetchRequest
	InputFromGraphQL(in *graphql.FetchRequestInput) *model.FetchRequestInput
}

//go:generate mockery -name=APIRuntimeAuthConverter -output=automock -outpkg=automock -case=underscore
type APIRuntimeAuthConverter interface {
	ToGraphQL(in *model.APIRuntimeAuth) *graphql.APIRuntimeAuth
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=APIRuntimeAuthService -output=automock -outpkg=automock -case=underscore
type APIRuntimeAuthService interface {
	Get(ctx context.Context, apiID string, runtimeID string) (*model.APIRuntimeAuth, error)
	GetOrDefault(ctx context.Context, apiID string, runtimeID string) (*model.APIRuntimeAuth, error)
	ListForAllRuntimes(ctx context.Context, apiID string) ([]model.APIRuntimeAuth, error)
	Set(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) error
	Delete(ctx context.Context, apiID string, runtimeID string) error
}

type Resolver struct {
	transact            persistence.Transactioner
	svc                 APIService
	appSvc              ApplicationService
	rtmSvc              RuntimeService
	apiRtmAuthSvc       APIRuntimeAuthService
	converter           APIConverter
	authConverter       AuthConverter
	frConverter         FetchRequestConverter
	apiRtmAuthConverter APIRuntimeAuthConverter
}

func NewResolver(transact persistence.Transactioner, svc APIService, appSvc ApplicationService, rtmSvc RuntimeService, apiRtmAuthSvc APIRuntimeAuthService, converter APIConverter, authConverter AuthConverter, frConverter FetchRequestConverter, apiRtmAuthConverter APIRuntimeAuthConverter) *Resolver {
	return &Resolver{
		transact:            transact,
		svc:                 svc,
		appSvc:              appSvc,
		rtmSvc:              rtmSvc,
		apiRtmAuthSvc:       apiRtmAuthSvc,
		converter:           converter,
		frConverter:         frConverter,
		authConverter:       authConverter,
		apiRtmAuthConverter: apiRtmAuthConverter,
	}
}

func (r *Resolver) AddAPIDefinition(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.converter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if !found {
		return nil, errors.New("Cannot add API to not existing Application")
	}

	id, err := r.svc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}
func (r *Resolver) UpdateAPIDefinition(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.converter.InputFromGraphQL(&in)

	err = r.svc.Update(ctx, id, *convertedIn)
	if err != nil {
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	return gqlAPI, nil
}
func (r *Resolver) DeleteAPIDefinition(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(api), nil
}
func (r *Resolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	spec, err := r.svc.RefetchAPISpec(ctx, apiID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.APIDefinition{Spec: spec})
	return convertedOut.Spec, nil
}

func (r *Resolver) Auth(ctx context.Context, obj *graphql.APIDefinition, runtimeID string) (*graphql.APIRuntimeAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	_, err = r.rtmSvc.Get(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Runtime '%s'", runtimeID)
	}

	ra, err := r.apiRtmAuthSvc.GetOrDefault(ctx, obj.ID, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting API Runtime Auth for Runtime '%s'", runtimeID)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	out := r.apiRtmAuthConverter.ToGraphQL(ra)

	return out, nil
}

func (r *Resolver) Auths(ctx context.Context, obj *graphql.APIDefinition) ([]*graphql.APIRuntimeAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	auths, err := r.apiRtmAuthSvc.ListForAllRuntimes(ctx, obj.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing API Runtime Auths")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	var out []*graphql.APIRuntimeAuth
	for _, ra := range auths {
		c := r.apiRtmAuthConverter.ToGraphQL(&ra)
		out = append(out, c)
	}
	return out, nil
}

func (r *Resolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.APIRuntimeAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.authConverter.InputFromGraphQL(&in)
	if convertedIn == nil {
		return nil, errors.New("object cannot be empty")
	}

	err = r.apiRtmAuthSvc.Set(ctx, apiID, runtimeID, *convertedIn)
	if err != nil {
		return nil, err
	}

	apiRtmAuth, err := r.apiRtmAuthSvc.Get(ctx, apiID, runtimeID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	convertedOut := &graphql.APIRuntimeAuth{
		RuntimeID: apiRtmAuth.RuntimeID,
		Auth:      r.authConverter.ToGraphQL(apiRtmAuth.Value),
	}

	return convertedOut, nil
}

func (r *Resolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.APIRuntimeAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	apiRtmAuth, err := r.apiRtmAuthSvc.Get(ctx, apiID, runtimeID)
	if err != nil {
		return nil, err
	}

	err = r.apiRtmAuthSvc.Delete(ctx, apiID, runtimeID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	convertedOut := &graphql.APIRuntimeAuth{
		RuntimeID: apiRtmAuth.RuntimeID,
		Auth:      r.authConverter.ToGraphQL(apiRtmAuth.Value),
	}

	return convertedOut, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, errors.New("API Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if obj.DefinitionID == "" {
		return nil, errors.New("Internal Server Error: Cannot fetch FetchRequest. APIDefinition ID is empty")
	}

	fr, err := r.svc.GetFetchRequest(ctx, obj.DefinitionID)
	if err != nil {
		return nil, err
	}

	if fr == nil {
		return nil, nil
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	frGQL := r.frConverter.ToGraphQL(fr)
	return frGQL, nil
}

// TODO: Replace with real implementation
func (r *Resolver) AddAPIDefinitionToPackage(ctx context.Context, packageID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return mock.FixAPIDefinition("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), nil
}
