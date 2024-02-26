package tenantfetchersvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"
	pkgwebhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

const (
	// RegistryLabelKey is the label key for registry label
	RegistryLabelKey = "registry"
	// SaaSRegistryLabelValue is the label value for saas registry label
	SaaSRegistryLabelValue = "saas-registry"
)

// WebhookService is responsible for the service-layer Webhook operations.
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListByTypeAndLabelFilter(ctx context.Context, webhookType model.WebhookType, filter *labelfilter.LabelFilter) ([]*model.Webhook, error)
	Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error
}

// TenantService is responsible for the service-layer Tenant operations.
//
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType directorresource.Type, objectID string) (string, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationService is responsible for the service-layer application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	UpdateBaseURLAndReadyState(ctx context.Context, appID, baseURL string, ready bool) error
}

// Client is responsible for the service-layer application operations.
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	UpdateBaseURLAndReadyState(ctx context.Context, appID, baseURL string, ready bool) error
}

// WebhookProcessor represents webhook processor
type WebhookProcessor struct {
	transact                       persistence.Transactioner
	webhookSvc                     WebhookService
	tenantSvc                      TenantService
	appSvc                         ApplicationService
	webhookClient                  *http.Client
	webhookProcessorElectionConfig cronjob.ElectionConfig
	webhookProcessorJobInterval    time.Duration

	systemFieldDiscoveryWebhookPartialProcessing     bool
	systemFieldDiscoveryWebhookPartialProcessMaxDays int
}

// NewWebhookProcessor creates new webhook processor
func NewWebhookProcessor(transact persistence.Transactioner, webhookSvc WebhookService, tenantSvc TenantService, appSvc ApplicationService, webhookClient *http.Client,
	webhookProcessorElectionConfig cronjob.ElectionConfig, webhookProcessorJobInterval time.Duration, systemFieldDiscoveryWebhookPartialProcessing bool, systemFieldDiscoveryWebhookPartialProcessMaxDays int) *WebhookProcessor {
	return &WebhookProcessor{
		transact:                       transact,
		webhookSvc:                     webhookSvc,
		tenantSvc:                      tenantSvc,
		appSvc:                         appSvc,
		webhookClient:                  webhookClient,
		webhookProcessorElectionConfig: webhookProcessorElectionConfig,
		webhookProcessorJobInterval:    webhookProcessorJobInterval,
		systemFieldDiscoveryWebhookPartialProcessing:     systemFieldDiscoveryWebhookPartialProcessing,
		systemFieldDiscoveryWebhookPartialProcessMaxDays: systemFieldDiscoveryWebhookPartialProcessMaxDays,
	}
}

// StartWebhookProcessorJob starts webhook processor job.
func (w *WebhookProcessor) StartWebhookProcessorJob(ctx context.Context, registryName string) error {
	resyncJob := cronjob.CronJob{
		Name: "WebhookProcessorJob",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Info("Starting WebhookProcessorJob...")

			if err := w.ProcessWebhooks(ctx, registryName); err != nil {
				log.C(jobCtx).Errorf("Error during execution of WebhookProcessorJob %v", err)
			}

			log.C(jobCtx).Infof("WebhookProcessorJob finished.")
		},
		SchedulePeriod: w.webhookProcessorJobInterval,
	}
	return cronjob.RunCronJob(ctx, w.webhookProcessorElectionConfig, resyncJob)
}

// ProcessWebhooks processes webhooks
func (w *WebhookProcessor) ProcessWebhooks(ctx context.Context, registryName string) error {
	log.C(ctx).Infof("Starting to process webhooks with type %q", model.WebhookTypeSystemFieldDiscovery)

	webhooks, err := w.listWebhooksBySystemFieldDiscoveryTypeAndLabelFilter(ctx, registryName)
	if err != nil {
		return errors.Wrapf(err, "failed listing webhooks by type %q and label", model.WebhookTypeSystemFieldDiscovery)
	}

	if w.systemFieldDiscoveryWebhookPartialProcessing {
		log.C(ctx).Infof("Partial system field discovery webhook processing is enabled for webhooks which are not older than %d days", w.systemFieldDiscoveryWebhookPartialProcessMaxDays)
	}

	date := time.Now().AddDate(0, 0, -1*w.systemFieldDiscoveryWebhookPartialProcessMaxDays)
	for _, wh := range webhooks {
		if w.systemFieldDiscoveryWebhookPartialProcessing && wh.CreatedAt.Before(date) {
			continue
		}

		ctx, err = saveCredentialsToContext(ctx, wh)
		if err != nil {
			log.C(ctx).Errorf(errors.Wrap(err, "failed saving credentials to context").Error())
			continue
		}

		respBody, err := pkgwebhook.ExecuteSystemFieldDiscoveryWebhook(ctx, w.webhookClient, wh)
		if err != nil {
			log.C(ctx).Errorf(errors.Wrapf(err, "failed executing webhook with id %q and type %q", wh.ID, model.WebhookTypeSystemFieldDiscovery).Error())
			continue
		}

		switch registryName {
		case SaaSRegistryLabelValue:
			var response pkgwebhook.SubscriptionsResponse
			if err = json.Unmarshal(respBody, &response); err != nil {
				log.C(ctx).Errorf(errors.Wrap(err, "failed to unmarshal subscriptions response").Error())
				continue
			}
			for _, subscription := range response.Subscriptions {
				processed, err := w.processSubscription(ctx, subscription, wh)
				if err != nil {
					log.C(ctx).Errorf(errors.Wrapf(err, "failed processing subscription for webhook with id %q and type %q", wh.ID, model.WebhookTypeSystemFieldDiscovery).Error())
					break
				}
				if processed {
					log.C(ctx).Infof("Successfully processed webhook with id %q", wh.ID)
					break
				}
				log.C(ctx).Infof("Response for webhook with ID %q does not contain app URL", wh.ID)
			}
		}

	}
	return nil
}

func (w *WebhookProcessor) listWebhooksBySystemFieldDiscoveryTypeAndLabelFilter(ctx context.Context, registryName string) ([]*model.Webhook, error) {
	tx, err := w.transact.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to begin a transaction for listing webhooks by type %q and label", model.WebhookTypeSystemFieldDiscovery)
	}
	defer w.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := w.webhookSvc.ListByTypeAndLabelFilter(ctx, model.WebhookTypeSystemFieldDiscovery, labelfilter.NewForKeyWithQuery(RegistryLabelKey, fmt.Sprintf("\"%s\"", registryName)))
	if err != nil {
		return nil, err
	}

	return webhooks, tx.Commit()
}

func (w *WebhookProcessor) processSubscription(ctx context.Context, subscription pkgwebhook.Subscription, webhook *model.Webhook) (bool, error) {
	if subscription.AppURL == "" {
		return false, nil
	}

	tx, err := w.transact.Begin()
	if err != nil {
		return false, errors.Wrapf(err, "failed to begin a transaction for processing subscriptions for webhook with id %q and type %q and label", webhook.ID, model.WebhookTypeSystemFieldDiscovery)
	}
	defer w.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	internalTntID, err := w.tenantSvc.GetLowestOwnerForResource(ctx, directorresource.Application, webhook.ObjectID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get lowest owner for %q resource with id %q", directorresource.Application, webhook.ObjectID)
	}

	tnt, err := w.tenantSvc.GetTenantByID(ctx, internalTntID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get tenant by id for internal tenant with id %q", internalTntID)
	}

	ctx = tenant.SaveToContext(ctx, internalTntID, tnt.ExternalTenant)

	if err := w.appSvc.UpdateBaseURLAndReadyState(ctx, webhook.ObjectID, subscription.AppURL, true); err != nil {
		return false, errors.Wrapf(err, "failed to update base url and ready state for webhook with id %q and app with id %q", webhook.ID, webhook.ObjectID)
	}

	if err := w.webhookSvc.Delete(ctx, webhook.ID, model.ApplicationWebhookReference); err != nil {
		return false, errors.Wrapf(err, "failed to delete application webhook with id %q for app with id %q", webhook.ID, webhook.ObjectID)
	}

	return true, tx.Commit()
}

func saveCredentialsToContext(ctx context.Context, webhook *model.Webhook) (context.Context, error) {
	if webhook.Auth != nil && webhook.Auth.Credential.Oauth != nil {
		credentials := &pkgAuth.OAuthCredentials{
			ClientID:     webhook.Auth.Credential.Oauth.ClientID,
			ClientSecret: webhook.Auth.Credential.Oauth.ClientSecret,
			TokenURL:     webhook.Auth.Credential.Oauth.URL,
		}
		ctx = pkgAuth.SaveToContext(ctx, credentials)
		return ctx, nil
	}
	return ctx, errors.Errorf("webhook credentials are missing for webhook with id %q", webhook.ID)
}
