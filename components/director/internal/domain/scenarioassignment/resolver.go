package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/mock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput) model.AutomaticScenarioAssignment
	ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
	GetByScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
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

	convertedIn := r.converter.FromInputGraphQL(in)

	out, err := r.svc.Create(ctx, convertedIn)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Assignment")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	assignment := r.converter.ToGraphQL(out)

	return &assignment, nil
}

func (r *Resolver) GetAutomaticScenarioAssignmentByScenarioName(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	out, err := r.svc.GetByScenarioName(ctx, scenarioName)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Assignment")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	assignment := r.converter.ToGraphQL(out)

	return &assignment, nil
}

func (r *Resolver) DeleteAutomaticScenarioAssignmentForSelector(ctx context.Context, selector graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	sel := &graphql.Label{Key: selector.Key, Value: selector.Value}
	data := []*graphql.AutomaticScenarioAssignment{
		mock.FixAssignmentForScenarioWithSelector("DEFAULT", sel),
		mock.FixAssignmentForScenarioWithSelector("Foo", sel),
	}

	return data, nil
}

func (r *Resolver) DeleteAutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	return mock.FixAssignmentForScenario(scenarioName), nil
}

func (r *Resolver) AutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	return mock.FixAssignmentForScenario(scenarioName), nil
}

func (r *Resolver) AutomaticScenarioAssignmentForSelector(ctx context.Context, selector graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	sel := &graphql.Label{Key: selector.Key, Value: selector.Value}
	data := []*graphql.AutomaticScenarioAssignment{
		mock.FixAssignmentForScenarioWithSelector("DEFAULT", sel),
		mock.FixAssignmentForScenarioWithSelector("Foo", sel),
	}

	return data, nil
}

func (r *Resolver) AutomaticScenarioAssignments(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.AutomaticScenarioAssignmentPage, error) {
	data := []*graphql.AutomaticScenarioAssignment{
		mock.FixAssignmentForScenario("DEFAULT"),
		mock.FixAssignmentForScenario("Foo"),
		mock.FixAssignmentForScenario("bar"),
		mock.FixAssignmentForScenario("fooBar"),
	}
	return &graphql.AutomaticScenarioAssignmentPage{
		Data:       data,
		TotalCount: len(data),
	}, nil
}
