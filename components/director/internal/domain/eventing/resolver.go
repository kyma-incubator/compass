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
	SetAsDefaultForApplication(ctx context.Context, runtimeID uuid.UUID, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error)
	DeleteDefaultForApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error)
}

type Resolver struct {
	transact    persistence.Transactioner
	eventingSvc EventingService
}

func NewResolver(transact persistence.Transactioner, eventingSvc EventingService) *Resolver {
	return &Resolver{
		transact:    transact,
		eventingSvc: eventingSvc,
	}
}

func (r *Resolver) SetDefaultEventingForApplication(ctx context.Context, app string, runtime string) (*graphql.ApplicationEventingConfiguration, error) {
	appID, err := uuid.Parse(app)
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

	eventingCfg, err := r.eventingSvc.SetAsDefaultForApplication(ctx, runtimeID, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}

func (r *Resolver) DeleteDefaultEventingForApplication(ctx context.Context, app string) (*graphql.ApplicationEventingConfiguration, error) {
	appID, err := uuid.Parse(app)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing application ID as UUID")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventingCfg, err := r.eventingSvc.DeleteDefaultForApplication(ctx, appID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}
