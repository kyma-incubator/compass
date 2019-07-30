package labeldef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type Resolver struct {
	conv          Converter
	srv           Service
	transactioner persistence.Transactioner
}

func NewResolver(srv Service, conv Converter, transactioner persistence.Transactioner) *Resolver {
	return &Resolver{
		conv:          conv,
		srv:           srv,
		transactioner: transactioner,
	}
}

// dependencies
//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	FromGraphQL(input graphql.LabelDefinitionInput, tenant string) model.LabelDefinition
	ToGraphQL(definition model.LabelDefinition) graphql.LabelDefinition
	ToEntity(in model.LabelDefinition) (Entity, error)
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	Create(ctx context.Context, ld model.LabelDefinition) (model.LabelDefinition, error)
}

func (r *Resolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transactioner.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while starting transaction")
	}
	defer r.transactioner.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	ld := r.conv.FromGraphQL(in, tnt)
	createdLd, err := r.srv.Create(ctx, ld)
	if err != nil {
		return nil, errors.Wrap(err, "while creating label definition")
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	out := r.conv.ToGraphQL(createdLd)

	return &out, nil
}
