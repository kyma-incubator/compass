package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=EventDefService -output=automock -outpkg=automock -case=underscore
type EventDefService interface {
	CreateInBundle(ctx context.Context, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput, spec *model.SpecInput) error
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	Delete(ctx context.Context, id string) error
	GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=EventDefConverter -output=automock -outpkg=automock -case=underscore
type EventDefConverter interface {
	ToGraphQL(in *model.EventDefinition, spec *model.Spec) (*graphql.EventDefinition, error)
	MultipleToGraphQL(in []*model.EventDefinition, specs []*model.Spec) ([]*graphql.EventDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, []*model.SpecInput, error)
	InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, *model.SpecInput, error)
}

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=BundleService -output=automock -outpkg=automock -case=underscore
type BundleService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

type Resolver struct {
	transact      persistence.Transactioner
	svc           EventDefService
	appSvc        ApplicationService
	bndlSvc       BundleService
	converter     EventDefConverter
	frConverter   FetchRequestConverter
	specConverter SpecConverter
	specService   SpecService
}

func NewResolver(transact persistence.Transactioner, svc EventDefService, appSvc ApplicationService, bndlSvc BundleService, converter EventDefConverter, frConverter FetchRequestConverter, specService SpecService, specConverter SpecConverter) *Resolver {
	return &Resolver{
		transact:      transact,
		svc:           svc,
		appSvc:        appSvc,
		bndlSvc:       bndlSvc,
		converter:     converter,
		frConverter:   frConverter,
		specConverter: specConverter,
		specService:   specService,
	}
}

func (r *Resolver) AddEventDefinitionToBundle(ctx context.Context, bundleID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Adding EventDefinition to bundle with id %s", bundleID)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to EventDefinition")
	}

	found, err := r.bndlSvc.Exist(ctx, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Bundle with id %s when adding EventDefinition", bundleID)
	}

	if !found {
		return nil, apperrors.NewInvalidDataError("cannot add Event Definition to not existing Bundle")
	}

	id, err := r.svc.CreateInBundle(ctx, bundleID, *convertedIn, convertedSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating EventDefinition in Bundle with id %s", bundleID)
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully added to bundle with id %s", id, bundleID)
	return gqlEvent, nil
}

func (r *Resolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Updating EventDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, convertedSpec, err := r.converter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting GraphQL input to EventDefinition with id %s", id)
	}

	err = r.svc.Update(ctx, id, *convertedIn, convertedSpec)
	if err != nil {
		return nil, err
	}

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully updated.", id)
	return gqlEvent, nil
}

func (r *Resolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Deleting EventDefinition with id %s", id)

	ctx = persistence.SaveToContext(ctx, tx)

	event, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	spec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	gqlEvent, err := r.converter.ToGraphQL(event, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting EventDefinition with id %q to graphQL", event.ID)
	}

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("EventDefinition with id %s successfully deleted.", id)
	return gqlEvent, nil
}

func (r *Resolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Refetching EventDefinitionSpec for EventDefinition with id %s", eventID)

	ctx = persistence.SaveToContext(ctx, tx)

	dbSpec, err := r.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, eventID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", eventID)
	}

	if dbSpec == nil {
		return nil, errors.Errorf("spec for Event with id %q not found", eventID)
	}

	spec, err := r.specService.RefetchSpec(ctx, dbSpec.ID)
	if err != nil {
		return nil, err
	}

	converted, err := r.specConverter.ToGraphQLEventSpec(spec)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Successfully refetched EventDefinitionSpec for EventDefinition with id %s", eventID)
	return converted, nil
}

func (r *Resolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Event Spec cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if obj.DefinitionID == "" {
		return nil, apperrors.NewInternalError("Cannot fetch FetchRequest. EventDefinition ID is empty")
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

	log.C(ctx).Infof("Successfully fetched request for EventDefinition %s", obj.DefinitionID)
	return r.frConverter.ToGraphQL(fr)
}
