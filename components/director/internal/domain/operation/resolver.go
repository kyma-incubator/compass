package operation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// OperationService is responsible for service-layer operation operations
//
//go:generate mockery --name=OperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationService interface {
	Get(ctx context.Context, id string) (*model.Operation, error)
	RescheduleOperation(ctx context.Context, operationID string, priority int) error
}

// OperationConverter is responsible for converting operations
//
//go:generate mockery --name=OperationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationConverter interface {
	ToGraphQL(in *model.Operation) (*graphql.Operation, error)
}

// Resolver is the operation resolver
type Resolver struct {
	transact persistence.Transactioner
	service  OperationService
	conv     OperationConverter
}

// NewResolver creates operation resolver
func NewResolver(transact persistence.Transactioner, service OperationService, conv OperationConverter) *Resolver {
	return &Resolver{
		transact: transact,
		service:  service,
		conv:     conv,
	}
}

// Operation returns a single operation by a given ID
func (r *Resolver) Operation(ctx context.Context, id string) (*graphql.Operation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Getting operation with ID %q", id)
	op, err := r.service.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(op)
}

// Schedule reschedules a specified operation with a given priority. Default priority is operationsmanager.HighOperationPriority
func (r *Resolver) Schedule(ctx context.Context, id string, priority *int) (*graphql.Operation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	opPriority := int(operationsmanager.HighOperationPriority)
	if priority != nil {
		opPriority = *priority
	}

	log.C(ctx).Infof("Rescheduling operation with ID %q and priority %d", id, opPriority)
	if err = r.service.RescheduleOperation(ctx, id, opPriority); err != nil {
		return nil, err
	}

	op, err := r.service.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Successfully rescheduled operation with ID %q", id)

	return r.conv.ToGraphQL(op)
}
