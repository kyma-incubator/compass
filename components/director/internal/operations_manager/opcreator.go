package operationsmanager

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"time"
)

const (
	// OrdCreatorType specifies open resource discovery creator type
	OrdCreatorType = "ORD"
	// OrdAggregationOpType specifies open resource discovery operation type
	OrdAggregationOpType = "ORD_AGGREGATION"
)

// OperationCreator is responsible for creation of different types of operations.
//go:generate mockery --name=OperationCreator --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationCreator interface {
	Create(ctx context.Context) error
}

// ORDOperationCreator consists of various resource services responsible for operations creation.
type ORDOperationCreator struct {
	transact   persistence.Transactioner
	opSvc      OperationService
	webhookSvc WebhookService
	appSvc     ApplicationService
}

// Create lists all webhooks of type "OPEN_RESOURCE_DISCOVERY" and for every application creates corresponding operation
func (oc *ORDOperationCreator) Create(ctx context.Context) error {
	operations, err := oc.buildOperationInputs(ctx)
	if err != nil {
		return errors.Wrap(err, "while building operation inputs")
	}

	tx, err := oc.transact.Begin()
	if err != nil {
		return err
	}
	defer oc.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := oc.opSvc.CreateMultiple(ctx, operations); err != nil {
		return errors.Wrap(err, "while creating multiple operations")
	}

	return tx.Commit()
}

func (oc *ORDOperationCreator) buildOperationInputs(ctx context.Context) ([]*model.OperationInput, error) {
	ordWebhooks, err := oc.getWebhooksWithOrdType(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while getting webhooks of type OPEN_RESOURCE_DISCOVERY")
	}

	operations := make([]*model.OperationInput, 0)
	for _, webhook := range ordWebhooks {
		if webhook.ObjectType == model.ApplicationTemplateWebhookReference {
			ops, err := oc.appTemplateWebhookToOperations(ctx, webhook)
			if err != nil {
				return nil, errors.Wrapf(err, "while creating operations from application template webhook")
			}
			operations = append(operations, ops...)
		} else if webhook.ObjectType == model.ApplicationWebhookReference {
			opData := NewOrdOperationData(webhook.ObjectID, "")
			data, err := opData.GetData()
			if err != nil {
				return nil, err
			}
			operations = append(operations, buildORDOperationInput(data))
		}
	}

	return operations, nil
}

func (oc *ORDOperationCreator) appTemplateWebhookToOperations(ctx context.Context, webhook *model.Webhook) ([]*model.OperationInput, error) {
	operations := make([]*model.OperationInput, 0)
	if webhook.ObjectType != model.ApplicationTemplateWebhookReference {
		return operations, nil
	}

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

func (oc *ORDOperationCreator) getWebhooksWithOrdType(ctx context.Context) ([]*model.Webhook, error) {
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

func (oc *ORDOperationCreator) getApplicationsForAppTemplate(ctx context.Context, appTemplateID string) ([]*model.Application, error) {
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
		OpType:     OrdAggregationOpType,
		Status:     scheduledOpStatus,
		Data:       json.RawMessage(data),
		Error:      nil,
		Priority:   1,
		CreatedAt:  &now,
		FinishedAt: nil,
	}
}

// NewOperationCreator creates OperationCreator based on kind
func NewOperationCreator(kind string, transact persistence.Transactioner, opSvc OperationService, webhookSvc WebhookService, appSvc ApplicationService) OperationCreator {
	if kind == OrdCreatorType {
		return &ORDOperationCreator{
			transact:   transact,
			opSvc:      opSvc,
			webhookSvc: webhookSvc,
			appSvc:     appSvc,
		}
	}
	return nil
}
