package eventing

import (
	"context"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=EventingService -output=automock -outpkg=automock -case=underscore
type EventingService interface {
	SetForApplication(ctx context.Context, runtimeID uuid.UUID, app model.Application) (*model.ApplicationEventingConfiguration, error)
	UnsetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error)
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Get(ctx context.Context, id string) (*model.Application, error)
}

type Resolver struct {
	transact    persistence.Transactioner
	eventingSvc EventingService
	appSvc      ApplicationService
}

func NewResolver(transact persistence.Transactioner, eventingSvc EventingService, appSvc ApplicationService) *Resolver {
	return &Resolver{
		transact:    transact,
		eventingSvc: eventingSvc,
		appSvc:      appSvc,
	}
}

func (r *Resolver) SetEventingForApplication(ctx context.Context, appID string, runtime string) (*graphql.ApplicationEventingConfiguration, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing application ID as UUID")
	}

	runtimeID, err := uuid.Parse(runtime)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	app, err := r.appSvc.Get(ctx, appUUID.String())
	if err != nil {
		return nil, errors.Wrap(err, "while getting application")
	}

	eventingCfg, err := r.eventingSvc.SetForApplication(ctx, runtimeID, *app)
	if err != nil {
		return nil, errors.Wrap(err, "while setting eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}

func (r *Resolver) UnsetEventingForApplication(ctx context.Context, appID string) (*graphql.ApplicationEventingConfiguration, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing application ID as UUID")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	app, err := r.appSvc.Get(ctx, appUUID.String())
	if err != nil {
		return nil, errors.Wrap(err, "while getting application")
	}

	eventingCfg, err := r.eventingSvc.UnsetForApplication(ctx, *app)
	if err != nil {
		return nil, errors.Wrap(err, "while unsetting eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}
