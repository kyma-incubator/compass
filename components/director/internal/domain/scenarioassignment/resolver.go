package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput) model.AutomaticScenarioAssignment
	ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment
	LabelSelectorFromInput(in graphql.LabelSelectorInput) model.LabelSelector
	MultipleToGraphQL(assignments []*model.AutomaticScenarioAssignment) []*graphql.AutomaticScenarioAssignment
}

//go:generate mockery -name=Service -output=automock -outpkg=automock -case=underscore
type Service interface {
	Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error)
	ListForSelector(ctx context.Context, in model.LabelSelector) ([]*model.AutomaticScenarioAssignment, error)
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
	DeleteManyForSameSelector(ctx context.Context, in []*model.AutomaticScenarioAssignment) error
	Delete(ctx context.Context, in model.AutomaticScenarioAssignment) error
}

func NewResolver(transact persistence.Transactioner, svc Service, converter Converter) *Resolver {
	return &Resolver{
		transact:  transact,
		svc:       svc,
		converter: converter,
	}
}

type Resolver struct {
	transact  persistence.Transactioner
	converter Converter
	svc       Service
}

func (r *Resolver) CreateAutomaticScenarioAssignment(ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

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

func (r *Resolver) GetAutomaticScenarioAssignmentForScenarioName(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	out, err := r.svc.GetForScenarioName(ctx, scenarioName)
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

func (r *Resolver) AutomaticScenarioAssignmentsForSelector(ctx context.Context, in graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	modelInput := r.converter.LabelSelectorFromInput(in)

	assignments, err := r.svc.ListForSelector(ctx, modelInput)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the assignments")
	}

	gqlAssignments := r.converter.MultipleToGraphQL(assignments)

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}
	return gqlAssignments, nil
}

func (r *Resolver) AutomaticScenarioAssignments(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.AutomaticScenarioAssignmentPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	page, err := r.svc.List(ctx, *first, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while listing the assignments")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gqlApps := r.converter.MultipleToGraphQL(page.Data)

	return &graphql.AutomaticScenarioAssignmentPage{
		Data:       gqlApps,
		TotalCount: page.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) DeleteAutomaticScenarioAssignmentsForSelector(ctx context.Context, in graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	selector := r.converter.LabelSelectorFromInput(in)

	assignments, err := r.svc.ListForSelector(ctx, selector)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the Assignments for selector [%v]", selector)
	}

	err = r.svc.DeleteManyForSameSelector(ctx, assignments)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting the Assignments for selector [%v]", selector)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.converter.MultipleToGraphQL(assignments), nil
}

func (r *Resolver) DeleteAutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while beginning transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	assignment, err := r.svc.GetForScenarioName(ctx, scenarioName)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the Assignment for scenario [name=%s]", scenarioName)
	}

	err = r.svc.Delete(ctx, assignment)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting the Assignment for scenario [name=%s]", scenarioName)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	gql := r.converter.ToGraphQL(assignment)

	return &gql, nil
}
