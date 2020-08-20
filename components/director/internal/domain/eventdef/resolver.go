package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=EventDefService -output=automock -outpkg=automock -case=underscore
type EventDefService interface {
	CreateInPackage(ctx context.Context, packageID string, in model.EventDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput) error
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.EventSpec, error)
	GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=EventDefConverter -output=automock -outpkg=automock -case=underscore
type EventDefConverter interface {
	ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition
	MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, error)
	InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, error)
}

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
	InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error)
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=PackageService -output=automock -outpkg=automock -case=underscore
type PackageService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	transact    persistence.Transactioner
	svc         EventDefService
	appSvc      ApplicationService
	pkgSvc      PackageService
	converter   EventDefConverter
	frConverter FetchRequestConverter
}

func NewResolver(transact persistence.Transactioner, svc EventDefService, appSvc ApplicationService, pkgSvc PackageService, converter EventDefConverter, frConverter FetchRequestConverter) *Resolver {
	return &Resolver{
		transact:    transact,
		svc:         svc,
		appSvc:      appSvc,
		pkgSvc:      pkgSvc,
		converter:   converter,
		frConverter: frConverter,
	}
}

func (r *Resolver) AddEventDefinitionToPackage(ctx context.Context, packageID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	log.Infof("Adding EventDefinition to package %s", packageID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		log.Error("Error while converting EventDefinition input from GraphQL. ", err.Error())
		return nil, errors.Wrap(err, "while converting EventDefinition input")
	}

	found, err := r.pkgSvc.Exist(ctx, packageID)
	if err != nil {
		log.Error("Error when checking existence of Package when adding EventDefinition. ", err.Error())
		return nil, errors.Wrapf(err, "while checking existence of Package")
	}

	if !found {
		log.Error("Error when adding EventDefinition due to not existing Package.")
		return nil, apperrors.NewInvalidDataError("cannot add Event Definition to not existing Package")
	}

	id, err := r.svc.CreateInPackage(ctx, packageID, *convertedIn)
	if err != nil {
		log.Errorf("Error when creating EventDefinition in package %s. %s", packageID, err.Error())
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

	log.Infof("EventDefinition successfully added to package %s", packageID)
	return gqlAPI, nil
}

func (r *Resolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	log.Infof("Updating EventDefinition %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		log.Errorf("Error while converting EventDefinition input from GraphQL with id %s : %s ", id, err.Error())
		return nil, errors.Wrap(err, "while converting EventDefinition input")
	}

	err = r.svc.Update(ctx, id, *convertedIn)
	if err != nil {
		log.Errorf("Error when updating EventDefinition %s : %s", id, err.Error())
		return nil, err
	}

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlAPI := r.converter.ToGraphQL(api)

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.Infof("EventDefinition %s successfully updated.", id)
	return gqlAPI, nil
}

func (r *Resolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	log.Infof("Deleting EventDefinition %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedAPI := r.converter.ToGraphQL(api)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		log.Errorf("Error when deleting EventDefinition %s : %s", id, err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.Infof("EventDefinition %s successfully deleted.", id)
	return deletedAPI, nil
}

func (r *Resolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	log.Infof("Refetching EventDefinitionSpec for EventDefinition %s", eventID)

	ctx = persistence.SaveToContext(ctx, tx)

	spec, err := r.svc.RefetchAPISpec(ctx, eventID)
	if err != nil {
		log.Errorf("Error when refetching EventDefinitionSpec for EventDefinition %s : %s", eventID, err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	convertedOut := r.converter.ToGraphQL(&model.EventDefinition{Spec: spec})

	log.Infof("Successfully refetched EventDefinitionSpec for EventDefinition %s", eventID)
	return convertedOut.Spec, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		log.Error("Error when fetching request for EventDefinition. Event Spec cannot be empty.")
		return nil, apperrors.NewInternalError("Event Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if obj.DefinitionID == "" {
		log.Error("Error when fetching FetchRequest. EventDefinition ID is empty.")
		return nil, apperrors.NewInternalError("Cannot fetch FetchRequest. EventDefinition ID is empty")
	}

	fr, err := r.svc.GetFetchRequest(ctx, obj.DefinitionID)
	if err != nil {
		return nil, err
	}

	if fr == nil {
		return nil, tx.Commit()
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully fetched request for EventDefinition %s", obj.DefinitionID)
	return r.frConverter.ToGraphQL(fr)
}
