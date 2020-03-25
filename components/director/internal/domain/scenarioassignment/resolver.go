package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput, tenant string) model.AutomaticScenarioAssignment
	ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
}

func NewResolver(transact persistence.Transactioner, converter Converter, svc Service) *Resolver {
	return &Resolver{
		transact:  transact,
		converter: converter,
		svc:       svc,
	}

}

type Resolver struct {
	transact  persistence.Transactioner
	converter Converter
	svc       Service
}

func (r *Resolver) SetAutomaticScenarioAssignment(ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	convertedIn := r.converter.FromInputGraphQL(in, tnt)
	out, err := r.svc.Create(ctx, convertedIn)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Assignment")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlApp := r.converter.ToGraphQL(out)

	return &gqlApp, nil
}
