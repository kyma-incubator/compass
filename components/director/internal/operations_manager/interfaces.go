package operationsmanager

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// OperationService is responsible for the service-layer Operation operations.
//
//go:generate mockery --name=OperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationService interface {
	CreateMultiple(ctx context.Context, in []*model.OperationInput) error
	MarkAsCompleted(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id, error string) error
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
