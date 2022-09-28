package formationassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Service interface {
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationAssignmentPage, error)
	GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error)
}

// FormationAssignmentConverter missing godoc
//go:generate mockery --name=FormationAssignmentConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentConverter interface {
	MultipleToGraphQL(in []*model.FormationAssignment) []*graphql.FormationAssignment
	ToGraphQL(in *model.FormationAssignment) *graphql.FormationAssignment
}

// Resolver is the formation assignment resolver
type Resolver struct {
	transact persistence.Transactioner
	service  Service
	conv     FormationAssignmentConverter
}

// todo::: consider removing this

// NewResolver creates formation assignment resolver
func NewResolver(transact persistence.Transactioner, service Service, conv FormationAssignmentConverter) *Resolver {
	return &Resolver{
		transact: transact,
		service:  service,
		conv:     conv,
	}
}

// FormationAssignment missing godoc
func (r *Resolver) FormationAssignment(ctx context.Context, obj *graphql.Formation, id string) (*graphql.FormationAssignment, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Formation Assignment cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtimeContext, err := r.service.GetForFormation(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(runtimeContext), nil
}

// FormationAssignments returns paginated FormationAssignment based on first and after
func (r *Resolver) FormationAssignments(ctx context.Context, obj *graphql.Formation, first *int, after *graphql.PageCursor) (*graphql.FormationAssignmentPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	faPage, err := r.service.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	formations := r.conv.MultipleToGraphQL(faPage.Data)

	return &graphql.FormationAssignmentPage{
		Data:       formations,
		TotalCount: faPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(faPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(faPage.PageInfo.EndCursor),
			HasNextPage: faPage.PageInfo.HasNextPage,
		},
	}, nil
}
