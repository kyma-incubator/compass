package scenarioassignment

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Converter interface {
	FromInputGraphql(in graphql.AutomaticScenarioAssignmentSetInput, tenant string) (model.AutomaticScenarioAssignment, error)
	ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment
}

type Service interface {
	Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
}

func NewResolver(transact persistence.Transactioner, converter Converter, svc Service) *resolver {
	return &resolver{
		transact:  transact,
		converter: converter,
		svc:       svc,
	}

}

type resolver struct {
	transact  persistence.Transactioner
	converter Converter
	svc       Service
}

func (r *resolver) SetAutomaticScenarioAssignment(ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	convertedIn, err := r.converter.FromInputGraphql(in, tnt)
	if err != nil {
		return nil, err
	}
	out, err := r.svc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.converter.ToGraphQL(out)

	return &gqlApp, nil
}
