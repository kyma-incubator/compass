package ord

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// OperationMaintainer is responsible for maintaining of different types of operations.
//
//go:generate mockery --name=OperationMaintainer --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationMaintainer interface {
	Maintain(ctx context.Context) error
}

// ORDOperationMaintainer consists of various resource services responsible for operations creation.
type ORDOperationMaintainer struct {
	transact   persistence.Transactioner
	opSvc      operationsmanager.OperationService
	webhookSvc WebhookService
	appSvc     ApplicationService
}

// NewOperationMaintainer creates OperationMaintainer based on kind
func NewOperationMaintainer(kind model.OperationType, transact persistence.Transactioner, opSvc operationsmanager.OperationService, webhookSvc WebhookService, appSvc ApplicationService) OperationMaintainer {
	if kind == model.OperationTypeOrdAggregation {
		return &ORDOperationMaintainer{
			transact:   transact,
			opSvc:      opSvc,
			webhookSvc: webhookSvc,
			appSvc:     appSvc,
		}
	}
	return nil
}

// Maintain is responsible to create all missing and remove all obsolete operations
func (oc *ORDOperationMaintainer) Maintain(ctx context.Context) error {
	tx, err := oc.transact.Begin()
	if err != nil {
		return err
	}
	defer oc.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	operationsToCreate, operationsToDelete, err := oc.buildNonExistingOperationInputs(ctx)
	if err != nil {
		return errors.Wrap(err, "while building operation inputs")
	}

	operationsToDeleteIDs := make([]string, 0)
	for _, op := range operationsToDelete {
		operationsToDeleteIDs = append(operationsToDeleteIDs, op.ID)
	}

	if err := oc.opSvc.CreateMultiple(ctx, operationsToCreate); err != nil {
		return errors.Wrap(err, "while creating multiple operations")
	}

	if err := oc.opSvc.DeleteMultiple(ctx, operationsToDeleteIDs); err != nil {
		return errors.Wrap(err, "while deleting multiple operations")
	}

	return tx.Commit()
}

func (oc *ORDOperationMaintainer) buildNonExistingOperationInputs(ctx context.Context) ([]*model.OperationInput, []*model.Operation, error) {
	ordWebhooks, err := oc.getWebhooksWithOrdType(ctx)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while getting webhooks of type %s", model.WebhookTypeOpenResourceDiscovery)
	}

	desiredStateOperations := make([]*model.OperationInput, 0)
	for _, webhook := range ordWebhooks {
		switch webhook.ObjectType {
		case model.ApplicationTemplateWebhookReference:
			ops, err := oc.appTemplateWebhookToOperations(ctx, webhook)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "while creating operations from application template webhook")
			}
			desiredStateOperations = append(desiredStateOperations, ops...)
		case model.ApplicationWebhookReference:
			opData := NewOrdOperationData(webhook.ObjectID, "")
			data, err := opData.GetData()
			if err != nil {
				return nil, nil, err
			}
			desiredStateOperations = append(desiredStateOperations, buildORDOperationInput(data))
		}
	}

	existingOperations, err := oc.opSvc.ListAllByType(ctx, model.OperationTypeOrdAggregation)
	if err != nil {
		return nil, nil, err
	}

	operationsToCreate, err := oc.getOperationsToCreate(desiredStateOperations, existingOperations)
	if err != nil {
		return nil, nil, err
	}

	operationsToDelete, err := oc.getOperationsToDelete(existingOperations, desiredStateOperations)
	if err != nil {
		return nil, nil, err
	}

	return operationsToCreate, operationsToDelete, nil
}

func (oc *ORDOperationMaintainer) getOperationsToCreate(desiredOperations []*model.OperationInput, existingOperations []*model.Operation) ([]*model.OperationInput, error) {
	result := make([]*model.OperationInput, 0)
	for _, currentDesiredOperation := range desiredOperations {
		found := false
		currentDesiredOperationData, err := ParseOrdOperationData(currentDesiredOperation.Data)
		if err != nil {
			return nil, err
		}
		for _, currentExistingOperation := range existingOperations {
			currentExistingOperationData, err := ParseOrdOperationData(currentExistingOperation.Data)
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

func (oc *ORDOperationMaintainer) getOperationsToDelete(existingOperations []*model.Operation, desiredOperations []*model.OperationInput) ([]*model.Operation, error) {
	result := make([]*model.Operation, 0)
	for _, currentExistingOperation := range existingOperations {
		found := false
		currentExistingOperationData, err := ParseOrdOperationData(currentExistingOperation.Data)
		if err != nil {
			return nil, err
		}
		for _, currentDesiredOperation := range desiredOperations {
			currentDesiredOperationData, err := ParseOrdOperationData(currentDesiredOperation.Data)
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

func (oc *ORDOperationMaintainer) appTemplateWebhookToOperations(ctx context.Context, webhook *model.Webhook) ([]*model.OperationInput, error) {
	operations := make([]*model.OperationInput, 0)
	if webhook.ObjectType != model.ApplicationTemplateWebhookReference {
		return operations, nil
	}

	opData := NewOrdOperationData("", webhook.ObjectID)
	data, err := opData.GetData()
	if err != nil {
		return nil, err
	}
	operations = append(operations, buildORDOperationInput(data))

	apps, err := oc.getApplicationsForAppTemplate(ctx, webhook.ObjectID)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		opData := NewOrdOperationData(app.ID, webhook.ObjectID)
		data, err := opData.GetData()
		if err != nil {
			return nil, err
		}
		operations = append(operations, buildORDOperationInput(data))
	}

	return operations, nil
}

func (oc *ORDOperationMaintainer) getWebhooksWithOrdType(ctx context.Context) ([]*model.Webhook, error) {
	tx, err := oc.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer oc.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhooks, err := oc.webhookSvc.ListByWebhookType(ctx, model.WebhookTypeOpenResourceDiscovery)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks with type %s", model.WebhookTypeOpenResourceDiscovery)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ordWebhooks, nil
}

func (oc *ORDOperationMaintainer) getApplicationsForAppTemplate(ctx context.Context, appTemplateID string) ([]*model.Application, error) {
	tx, err := oc.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer oc.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	apps, err := oc.appSvc.ListAllByApplicationTemplateID(ctx, appTemplateID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return apps, err
}

func buildORDOperationInput(data string) *model.OperationInput {
	now := time.Now()
	return &model.OperationInput{
		OpType:    model.OperationTypeOrdAggregation,
		Status:    model.OperationStatusScheduled,
		Data:      json.RawMessage(data),
		Error:     nil,
		Priority:  int(operationsmanager.LowOperationPriority),
		CreatedAt: &now,
		UpdatedAt: nil,
	}
}
