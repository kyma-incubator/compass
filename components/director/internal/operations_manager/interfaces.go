package operationsmanager

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// OperationPriority used fro operation priority
type OperationPriority int

const (
	// LowOperationPriority represents low priority for operations
	LowOperationPriority OperationPriority = 1
	// HighOperationPriority represents high priority for operations
	HighOperationPriority OperationPriority = 100
)

// OperationService is responsible for the service-layer Operation operations.
//
//go:generate mockery --name=OperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationService interface {
	CreateMultiple(ctx context.Context, in []*model.OperationInput) error
	DeleteMultiple(ctx context.Context, ids []string) error
	MarkAsCompleted(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id, errorMsg string) error
	ListAllByType(ctx context.Context, opType model.OperationType) ([]*model.Operation, error)
	ListPriorityQueue(ctx context.Context, queueLimit int, opType model.OperationType) ([]*model.Operation, error)
	LockOperation(ctx context.Context, operationID string) (bool, error)
	Get(ctx context.Context, operationID string) (*model.Operation, error)
	Update(ctx context.Context, input *model.Operation) error
	RescheduleOperations(ctx context.Context, operationType model.OperationType, reschedulePeriod time.Duration) error
	RescheduleHangedOperations(ctx context.Context, operationType model.OperationType, hangPeriod time.Duration) error
	RescheduleOperation(ctx context.Context, operationID string, priority int) error
}

// WebhookService is responsible for the service-layer Webhook operations.
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error)
}

// ApplicationService is responsible for the service-layer Application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	ListAllByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Application, error)
}

// OperationsProcessor is responsible for processing of scheduled operations.
//
//go:generate mockery --name=OperationsProcessor --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationsProcessor interface {
	Process(ctx context.Context, operation *model.Operation) error
}
