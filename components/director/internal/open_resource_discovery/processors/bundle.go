package processors

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
)

// BundleService is responsible for the service-layer Bundle operations.
//
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	CreateBundle(ctx context.Context, resourceType resource.Type, resourceID string, in model.BundleCreateInput, bndlHash uint64) (string, error)
	UpdateBundle(ctx context.Context, resourceType resource.Type, id string, in model.BundleUpdateInput, bndlHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
	ListByApplicationTemplateVersionIDNoPaging(ctx context.Context, appTemplateVersionID string) ([]*model.Bundle, error)
}

// WebhookConverter is responsible for converting webhook structs
//
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
}

// WebhookService is responsible for the service-layer Webhook operations.
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	GetByIDAndWebhookTypeGlobal(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
	ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error)
	ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	ListForApplicationGlobal(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	EnrichWebhooksWithTenantMappingWebhooks(in []*graphql.WebhookInput) ([]*graphql.WebhookInput, error)
	Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error)
	Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error
}

const (
	customTypeProperty  = "customType"
	callbackURLProperty = "callbackUrl"

	// TenantMappingCustomTypeIdentifier represents an identifier for tenant mapping webhooks in Credential exchange strategies
	TenantMappingCustomTypeIdentifier = "sap.ucl:tenant-mapping"
)

// CredentialExchangeStrategyTenantMapping contains tenant mappings configuration
type CredentialExchangeStrategyTenantMapping struct {
	Mode    model.WebhookMode
	Version string
}

type BundleProcessor struct {
	transact                                 persistence.Transactioner
	bundleSvc                                BundleService
	webhookSvc                               WebhookService
	webhookConverter                         WebhookConverter
	credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping
}

// NewBundleProcessor creates new instance of BundleProcessor
func NewBundleProcessor(transact persistence.Transactioner, bundleSvc BundleService, webhookSvc WebhookService, webhookConverter WebhookConverter, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) *BundleProcessor {
	return &BundleProcessor{
		transact:                                 transact,
		bundleSvc:                                bundleSvc,
		webhookSvc:                               webhookSvc,
		webhookConverter:                         webhookConverter,
		credentialExchangeStrategyTenantMappings: credentialExchangeStrategyTenantMappings,
	}
}

func (bp *BundleProcessor) Process(ctx context.Context, resourceType directorresource.Type, resourceID string, bundles []*model.BundleCreateInput, resourceHashes map[string]uint64) ([]*model.Bundle, error) {
	bundlesFromDB, err := bp.listBundlesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	credentialExchangeStrategyHashCurrent := uint64(0)
	var credentialExchangeStrategyJSON gjson.Result
	for _, bndl := range bundles {
		bndlHash := resourceHashes[str.PtrStrToStr(bndl.OrdID)]
		if err := bp.resyncBundleInTx(ctx, resourceType, resourceID, bundlesFromDB, bndl, bndlHash); err != nil {
			return nil, err
		}

		credentialExchangeStrategies, err := bndl.CredentialExchangeStrategies.MarshalJSON()
		if err != nil {
			return nil, errors.Wrapf(err, "while marshalling credential exchange strategies for %s with ID %s", resourceType, resourceID)
		}

		for _, credentialExchangeStrategy := range gjson.ParseBytes(credentialExchangeStrategies).Array() {
			customType := credentialExchangeStrategy.Get(customTypeProperty).String()
			isTenantMappingType := strings.Contains(customType, TenantMappingCustomTypeIdentifier)

			if !isTenantMappingType {
				continue
			}

			currentHash, err := HashObject(credentialExchangeStrategy)
			if err != nil {
				return nil, errors.Wrapf(err, "while hasing credential exchange strategy for application with ID %s", resourceID)
			}

			if credentialExchangeStrategyHashCurrent != 0 && currentHash != credentialExchangeStrategyHashCurrent {
				return nil, errors.Errorf("There are differences in the Credential Exchange Strategies for Tenant Mappings for application with ID %s. They should be the same.", resourceID)
			}

			credentialExchangeStrategyHashCurrent = currentHash
			credentialExchangeStrategyJSON = credentialExchangeStrategy
		}
	}

	if err = bp.resyncTenantMappingWebhooksInTx(ctx, credentialExchangeStrategyJSON, resourceID); err != nil {
		return nil, err
	}

	bundlesFromDB, err = bp.listBundlesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	return bundlesFromDB, nil
}

func (bp *BundleProcessor) listBundlesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Bundle, error) {
	tx, err := bp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer bp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var bundlesFromDB []*model.Bundle
	switch resourceType {
	case directorresource.Application:
		bundlesFromDB, err = bp.bundleSvc.ListByApplicationIDNoPaging(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
		bundlesFromDB, err = bp.bundleSvc.ListByApplicationTemplateVersionIDNoPaging(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing bundles for %s with id %q", resourceType, resourceID)
	}

	return bundlesFromDB, tx.Commit()
}

func (bp *BundleProcessor) resyncBundleInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, bundle *model.BundleCreateInput, bndlHash uint64) error {
	tx, err := bp.transact.Begin()
	if err != nil {
		return err
	}
	defer bp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := bp.resyncBundle(ctx, resourceType, resourceID, bundlesFromDB, *bundle, bndlHash); err != nil {
		return errors.Wrapf(err, "error while resyncing bundle with ORD ID %q", *bundle.OrdID)
	}
	return tx.Commit()
}

func (bp *BundleProcessor) resyncBundle(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput, bndlHash uint64) error {
	ctx = addFieldToLogger(ctx, "bundle_ord_id", *bndl.OrdID)
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		return bp.bundleSvc.UpdateBundle(ctx, resourceType, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl), bndlHash)
	}

	_, err := bp.bundleSvc.CreateBundle(ctx, resourceType, resourceID, bndl, bndlHash)
	return err
}

func bundleUpdateInputFromCreateInput(in model.BundleCreateInput) model.BundleUpdateInput {
	return model.BundleUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: in.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            in.DefaultInstanceAuth,
		OrdID:                          in.OrdID,
		ShortDescription:               in.ShortDescription,
		Links:                          in.Links,
		Labels:                         in.Labels,
		DocumentationLabels:            in.DocumentationLabels,
		CredentialExchangeStrategies:   in.CredentialExchangeStrategies,
		CorrelationIDs:                 in.CorrelationIDs,
	}
}

func (bp *BundleProcessor) resyncTenantMappingWebhooksInTx(ctx context.Context, credentialExchangeStrategyJSON gjson.Result, appID string) error {
	if !credentialExchangeStrategyJSON.IsObject() {
		log.C(ctx).Debugf("There are no tenant mappings to resync")
		return nil
	}

	tenantMappingData, err := bp.getTenantMappingData(credentialExchangeStrategyJSON, appID)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Enriching tenant mapping webhooks for application with ID %s", appID)

	enrichedWebhooks, err := bp.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks([]*graphql.WebhookInput{createWebhookInput(credentialExchangeStrategyJSON, tenantMappingData)})
	if err != nil {
		return errors.Wrapf(err, "while enriching webhooks with tenant mapping webhooks for application with ID %s", appID)
	}

	ctxWithoutTenant := context.Background()
	tx, err := bp.transact.Begin()
	if err != nil {
		return err
	}
	defer bp.transact.RollbackUnlessCommitted(ctxWithoutTenant, tx)

	ctxWithoutTenant = persistence.SaveToContext(ctxWithoutTenant, tx)
	ctxWithoutTenant = tenant.SaveToContext(ctxWithoutTenant, "", "")

	appWebhooksFromDB, err := bp.webhookSvc.ListForApplicationGlobal(ctxWithoutTenant, appID)
	if err != nil {
		return errors.Wrapf(err, "while listing webhooks from application with ID %s", appID)
	}

	tenantMappingRelatedWebhooksFromDB, enrichedWhModels, enrichedWhModelInputs, err := bp.processEnrichedWebhooks(enrichedWebhooks, appWebhooksFromDB)
	if err != nil {
		return err
	}

	isEqual, err := isWebhookDataEqual(tenantMappingRelatedWebhooksFromDB, enrichedWhModels)
	if err != nil {
		return err
	}

	if isEqual {
		log.C(ctxWithoutTenant).Infof("There are no differences in tenant mapping webhooks from the DB and the ORD document")
		return tx.Commit()
	}

	log.C(ctxWithoutTenant).Infof("There are differences in tenant mapping webhooks from the DB and the ORD document. Continuing the sync.")

	if err := bp.deleteWebhooks(ctxWithoutTenant, tenantMappingRelatedWebhooksFromDB, appID); err != nil {
		return err
	}

	if err := bp.createWebhooks(ctxWithoutTenant, enrichedWhModelInputs, appID); err != nil {
		return err
	}

	return tx.Commit()
}

func (bp *BundleProcessor) getTenantMappingData(credentialExchangeStrategyJSON gjson.Result, appID string) (CredentialExchangeStrategyTenantMapping, error) {
	tenantMappingType := credentialExchangeStrategyJSON.Get(customTypeProperty).String()
	tenantMappingData, ok := bp.credentialExchangeStrategyTenantMappings[tenantMappingType]
	if !ok {
		return CredentialExchangeStrategyTenantMapping{}, errors.Errorf("Credential Exchange Strategy has invalid %s value: %s for application with ID %s", customTypeProperty, tenantMappingType, appID)
	}
	return tenantMappingData, nil
}

func isWebhookDataEqual(tenantMappingRelatedWebhooksFromDB, enrichedWhModels []*model.Webhook) (bool, error) {
	appWhsFromDBMarshaled, err := json.Marshal(tenantMappingRelatedWebhooksFromDB)
	if err != nil {
		return false, errors.Wrapf(err, "while marshalling webhooks from DB")
	}

	appWhsFromDBHash, err := HashObject(string(appWhsFromDBMarshaled))
	if err != nil {
		return false, errors.Wrapf(err, "while hashing webhooks from DB")
	}

	enrichedWhsMarshaled, err := json.Marshal(enrichedWhModels)
	if err != nil {
		return false, errors.Wrapf(err, "while marshalling webhooks from DB")
	}

	enrichedHash, err := HashObject(string(enrichedWhsMarshaled))
	if err != nil {
		return false, errors.Wrapf(err, "while hashing webhooks from ORD document")
	}

	if strconv.FormatUint(appWhsFromDBHash, 10) == strconv.FormatUint(enrichedHash, 10) {
		return true, nil
	}

	return false, nil
}

func (bp *BundleProcessor) processEnrichedWebhooks(enrichedWebhooks []*graphql.WebhookInput, webhooksFromDB []*model.Webhook) ([]*model.Webhook, []*model.Webhook, []*model.WebhookInput, error) {
	tenantMappingRelatedWebhooksFromDB := make([]*model.Webhook, 0)
	enrichedWebhookModels := make([]*model.Webhook, 0)
	enrichedWebhookModelInputs := make([]*model.WebhookInput, 0)

	for _, wh := range enrichedWebhooks {
		convertedIn, err := bp.webhookConverter.InputFromGraphQL(wh)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "while converting the WebhookInput")
		}

		enrichedWebhookModelInputs = append(enrichedWebhookModelInputs, convertedIn)

		webhookModel := convertedIn.ToWebhook("", "", "")

		for _, webhookFromDB := range webhooksFromDB {
			if webhookFromDB.Type == convertedIn.Type {
				webhookModel.ID = webhookFromDB.ID
				webhookModel.ObjectType = webhookFromDB.ObjectType
				webhookModel.ObjectID = webhookFromDB.ObjectID
				webhookModel.CreatedAt = webhookFromDB.CreatedAt

				tenantMappingRelatedWebhooksFromDB = append(tenantMappingRelatedWebhooksFromDB, webhookFromDB)
				break
			}
		}

		enrichedWebhookModels = append(enrichedWebhookModels, webhookModel)
	}

	return tenantMappingRelatedWebhooksFromDB, enrichedWebhookModels, enrichedWebhookModelInputs, nil
}

func createWebhookInput(credentialExchangeStrategyJSON gjson.Result, tenantMappingData CredentialExchangeStrategyTenantMapping) *graphql.WebhookInput {
	inputMode := graphql.WebhookMode(tenantMappingData.Mode)
	return &graphql.WebhookInput{
		URL: str.Ptr(credentialExchangeStrategyJSON.Get(callbackURLProperty).String()),
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode:    &inputMode,
		Version: str.Ptr(tenantMappingData.Version),
	}
}

func (bp *BundleProcessor) deleteWebhooks(ctx context.Context, webhooks []*model.Webhook, appID string) error {
	for _, webhook := range webhooks {
		log.C(ctx).Infof("Deleting webhook with ID %s for application %s", webhook.ID, appID)
		if err := bp.webhookSvc.Delete(ctx, webhook.ID, webhook.ObjectType); err != nil {
			log.C(ctx).Errorf("error while deleting webhook with ID %s", webhook.ID)
			return errors.Wrapf(err, "while deleting webhook with ID %s", webhook.ID)
		}
	}

	return nil
}

func (bp *BundleProcessor) createWebhooks(ctx context.Context, webhooks []*model.WebhookInput, appID string) error {
	for _, webhook := range webhooks {
		log.C(ctx).Infof("Creating webhook with type %s for application %s", webhook.Type, appID)
		if _, err := bp.webhookSvc.Create(ctx, appID, *webhook, model.ApplicationWebhookReference); err != nil {
			log.C(ctx).Errorf("error while creating webhook for app %s with type %s", appID, webhook.Type)
			return errors.Wrapf(err, "error while creating webhook for app %s with type %s", appID, webhook.Type)
		}
	}

	return nil
}
