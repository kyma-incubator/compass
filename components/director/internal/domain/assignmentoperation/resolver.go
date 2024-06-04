package assignmentoperation

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

//go:generate mockery --exported --name=assignmentOperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationService interface {
	GetLatestOperation(ctx context.Context, assignmentID, formationID string) (*model.AssignmentOperation, error)
}

//go:generate mockery --exported --name=assignmentOperationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationConverter interface {
	ToGraphQL(in *model.AssignmentOperation) *graphql.AssignmentOperation
}

// Resolver is a formation assignment resolver
type Resolver struct {
	transact                     persistence.Transactioner
	assignmentOperationService   assignmentOperationService
	assignmentOperationConverter assignmentOperationConverter
}

// NewResolver returns a new resolver
func NewResolver(transact persistence.Transactioner, assignmentOperationService assignmentOperationService, assignmentOperationConverter assignmentOperationConverter) *Resolver {
	return &Resolver{
		transact:                     transact,
		assignmentOperationService:   assignmentOperationService,
		assignmentOperationConverter: assignmentOperationConverter,
	}
}

// GetLatestOperation returns the latest operation for the given assignment and formation
func (r *Resolver) GetLatestOperation(ctx context.Context, assignmentID, formationID string) (*graphql.AssignmentOperation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	latestOperation, err := r.assignmentOperationService.GetLatestOperation(ctx, assignmentID, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting latest operation for assignment %s and formation %s", assignmentID, formationID)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.assignmentOperationConverter.ToGraphQL(latestOperation), nil
}
