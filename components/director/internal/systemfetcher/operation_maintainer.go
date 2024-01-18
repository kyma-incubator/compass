package systemfetcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

// OperationMaintainer is responsible for maintaining of different types of operations.
//
//go:generate mockery --name=OperationMaintainer --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationMaintainer interface {
	Maintain(ctx context.Context) error
}

// OperationService is responsible for the service-layer Operation operations.
//
//go:generate mockery --name=OperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationService interface {
	CreateMultiple(ctx context.Context, in []*model.OperationInput) error
	DeleteMultiple(ctx context.Context, ids []string) error
	ListAllByType(ctx context.Context, opType model.OperationType) ([]*model.Operation, error)
}

// BusinessTenantMappingService missing godoc
//
//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BusinessTenantMappingService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
	Exists(ctx context.Context, id string) error
	ExistsByExternalTenant(ctx context.Context, externalTenant string) error
}

// SytemFetcherOperationMaintainer consists of various resource services responsible for operations creation.
type SytemFetcherOperationMaintainer struct {
	transact                 persistence.Transactioner
	opSvc                    OperationService
	businessTenantMappingSvc BusinessTenantMappingService
}

// NewOperationMaintainer creates OperationMaintainer based on kind
func NewOperationMaintainer(kind model.OperationType, transact persistence.Transactioner, opSvc OperationService, businessTenantMappingSvc BusinessTenantMappingService) OperationMaintainer {
	if kind == model.OperationTypeSystemFetcherAggregation {
		return &SytemFetcherOperationMaintainer{
			transact:                 transact,
			opSvc:                    opSvc,
			businessTenantMappingSvc: businessTenantMappingSvc,
		}
	}
	return nil
}

// Maintain is responsible to create all missing and remove all obsolete operations
func (om *SytemFetcherOperationMaintainer) Maintain(ctx context.Context) error {
	tx, err := om.transact.Begin()
	if err != nil {
		return err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	operationsToCreate, operationsToDelete, err := om.buildNonExistingOperationInputs(ctx)
	if err != nil {
		return errors.Wrap(err, "while building operation inputs")
	}

	operationsToDeleteIDs := make([]string, 0)
	for _, op := range operationsToDelete {
		operationsToDeleteIDs = append(operationsToDeleteIDs, op.ID)
	}

	if err := om.opSvc.CreateMultiple(ctx, operationsToCreate); err != nil {
		return errors.Wrap(err, "while creating multiple operations")
	}

	if err := om.opSvc.DeleteMultiple(ctx, operationsToDeleteIDs); err != nil {
		return errors.Wrap(err, "while deleting multiple operations")
	}

	return tx.Commit()
}

func (om *SytemFetcherOperationMaintainer) buildNonExistingOperationInputs(ctx context.Context) ([]*model.OperationInput, []*model.Operation, error) {
	tenants, err := om.listBusinessTenantMappings(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while getting tenants")
	}

	desiredStateOperations := make([]*model.OperationInput, 0)
	for _, tenant := range tenants {
		switch tenant.Type {
		case tenantpkg.Customer, tenantpkg.Account:
			op, err := om.tenantToOperation(ctx, tenant)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "while creating operation from account tenant %q", tenant.ID)
			}
			desiredStateOperations = append(desiredStateOperations, op)
		default:
			// Unknown type
			log.C(ctx).Infof("No operation created for tenant as it is not of type customer or account: %q", tenant.ID)
		}
	}

	existingOperations, err := om.opSvc.ListAllByType(ctx, model.OperationTypeSystemFetcherAggregation)
	if err != nil {
		return nil, nil, err
	}

	operationsToCreate, err := om.getOperationsToCreate(desiredStateOperations, existingOperations)
	if err != nil {
		return nil, nil, err
	}

	operationsToDelete, err := om.getOperationsToDelete(existingOperations, desiredStateOperations)
	if err != nil {
		return nil, nil, err
	}

	return operationsToCreate, operationsToDelete, nil
}

func (om *SytemFetcherOperationMaintainer) getOperationsToCreate(desiredOperations []*model.OperationInput, existingOperations []*model.Operation) ([]*model.OperationInput, error) {
	result := make([]*model.OperationInput, 0)
	for _, currentDesiredOperation := range desiredOperations {
		found := false
		currentDesiredOperationData, err := ParseSystemFetcherOperationData(currentDesiredOperation.Data)
		if err != nil {
			return nil, err
		}
		for _, currentExistingOperation := range existingOperations {
			currentExistingOperationData, err := ParseSystemFetcherOperationData(currentExistingOperation.Data)
			if err != nil {
				return nil, err
			}
			if currentDesiredOperationData.Equal(currentExistingOperationData) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, currentDesiredOperation)
		}
	}
	return result, nil
}

func (om *SytemFetcherOperationMaintainer) getOperationsToDelete(existingOperations []*model.Operation, desiredOperations []*model.OperationInput) ([]*model.Operation, error) {
	result := make([]*model.Operation, 0)
	for _, currentExistingOperation := range existingOperations {
		found := false
		currentExistingOperationData, err := ParseSystemFetcherOperationData(currentExistingOperation.Data)
		if err != nil {
			return nil, err
		}
		for _, currentDesiredOperation := range desiredOperations {
			currentDesiredOperationData, err := ParseSystemFetcherOperationData(currentDesiredOperation.Data)
			if err != nil {
				return nil, err
			}
			if currentDesiredOperationData.Equal(currentExistingOperationData) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, currentExistingOperation)
		}
	}
	return result, nil
}

func (om *SytemFetcherOperationMaintainer) listBusinessTenantMappings(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	tenants, err := om.businessTenantMappingSvc.List(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while fetching tenants")
		return nil, err
	}

	return tenants, nil
}

func (om *SytemFetcherOperationMaintainer) tenantToOperation(ctx context.Context, tenant *model.BusinessTenantMapping) (*model.OperationInput, error) {
	opData := NewSystemFetcherOperationData(tenant.ID)
	data, err := opData.GetData()
	if err != nil {
		return nil, err
	}
	operation := buildSystemFetcherOperationInput(data)

	return operation, nil
}

func buildSystemFetcherOperationInput(data string) *model.OperationInput {
	now := time.Now()
	return &model.OperationInput{
		OpType:    model.OperationTypeSystemFetcherAggregation,
		Status:    model.OperationStatusScheduled,
		Data:      json.RawMessage(data),
		Error:     nil,
		Priority:  int(operationsmanager.LowOperationPriority),
		CreatedAt: &now,
		UpdatedAt: nil,
	}
}
