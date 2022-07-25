package claims

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
)

// RuntimeService is used to interact with runtimes.
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	GetLabel(context.Context, string, string) (*model.Label, error)
	GetByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error)
}

// RuntimeCtxService is used to interact with runtime contexts.
//go:generate mockery --name=RuntimeCtxService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeCtxService interface {
	ListByFilter(ctx context.Context, runtimeID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error)
}

// ApplicationTemplateService is used to interact with application templates.
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
}

// ApplicationService is used to interact with runtime contexts.
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	ListAll(ctx context.Context) ([]*model.Application, error)
}

// IntegrationSystemService is used to check if integration system with a given ID exists.
//go:generate mockery --name=IntegrationSystemService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationSystemService interface {
	Exists(context.Context, string) (bool, error)
}

type validator struct {
	transact                     persistence.Transactioner
	runtimesSvc                  RuntimeService
	runtimeCtxSvc                RuntimeCtxService
	appTemplateSvc               ApplicationTemplateService
	applicationSvc               ApplicationService
	intSystemSvc                 IntegrationSystemService
	subscriptionProviderLabelKey string
	consumerSubaccountLabelKey   string
	tokenPrefix                  string
}

// NewValidator creates new claims validator
func NewValidator(transact persistence.Transactioner, runtimesSvc RuntimeService, runtimeCtxSvc RuntimeCtxService, appTemplateSvc ApplicationTemplateService, applicationSvc ApplicationService, intSystemSvc IntegrationSystemService, subscriptionProviderLabelKey, consumerSubaccountLabelKey, tokenPrefix string) *validator {
	return &validator{
		transact:                     transact,
		runtimesSvc:                  runtimesSvc,
		runtimeCtxSvc:                runtimeCtxSvc,
		appTemplateSvc:               appTemplateSvc,
		applicationSvc:               applicationSvc,
		intSystemSvc:                 intSystemSvc,
		subscriptionProviderLabelKey: subscriptionProviderLabelKey,
		consumerSubaccountLabelKey:   consumerSubaccountLabelKey,
		tokenPrefix:                  tokenPrefix,
	}
}

// Validate validates given id_token claims
func (v *validator) Validate(ctx context.Context, claims Claims) error {
	if err := claims.Valid(); err != nil {
		return errors.Wrapf(err, "while validating claims")
	}

	if claims.Tenant[tenantmapping.ConsumerTenantKey] == "" && claims.Tenant[tenantmapping.ExternalTenantKey] != "" {
		return apperrors.NewTenantNotFoundError(claims.Tenant[tenantmapping.ExternalTenantKey])
	}

	if claims.OnBehalfOf == "" {
		return nil
	}

	log.C(ctx).Infof("Consumer-Provider call by %s on behalf of %s. Proceeding with double authentication crosscheck...", claims.Tenant[tenantmapping.ProviderTenantKey], claims.Tenant[tenantmapping.ConsumerTenantKey])
	switch claims.ConsumerType {
	case consumer.Runtime, consumer.ExternalCertificate, consumer.SuperAdmin: // SuperAdmin consumer is needed only for testing purposes
		errRuntimeConsumer := v.validateRuntimeConsumer(ctx, claims)
		if errRuntimeConsumer == nil {
			return nil
		}
		errAppProvider := v.validateApplicationProvider(ctx, claims)
		if errAppProvider == nil {
			return nil
		}
		return apperrors.NewUnauthorizedError(fmt.Sprintf("subscription record not found neither for application: %q nor for runtime: %q", errAppProvider.Error(), errRuntimeConsumer.Error()))
	case consumer.IntegrationSystem:
		return v.validateIntegrationSystemConsumer(ctx, claims)
	default:
		return apperrors.NewUnauthorizedError(fmt.Sprintf("consumer with type %s is not supported", claims.ConsumerType))
	}
}

func (v *validator) validateRuntimeConsumer(ctx context.Context, claims Claims) error {
	tx, err := v.transact.Begin()
	if err != nil {
		log.C(ctx).Errorf("An error has occurred while opening transaction: %v", err)
		return errors.Wrapf(err, "An error has occurred while opening transaction")
	}
	defer v.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if len(claims.TokenClientID) == 0 {
		log.C(ctx).Errorf("Could not find consumer token client ID")
		return apperrors.NewUnauthorizedError("could not find consumer token client ID")
	}
	if len(claims.Region) == 0 {
		log.C(ctx).Errorf("Could not determine consumer token's region")
		return apperrors.NewUnauthorizedError("could not determine token's region")
	}

	tokenClientID := strings.TrimPrefix(claims.TokenClientID, v.tokenPrefix)
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(v.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", tokenClientID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", claims.Region)),
	}

	providerInternalTenantID := claims.Tenant[tenantmapping.ProviderTenantKey]
	providerExternalTenantID := claims.Tenant[tenantmapping.ProviderExternalTenantKey]
	ctxWithProviderTenant := tenant.SaveToContext(ctx, providerInternalTenantID, providerExternalTenantID)

	log.C(ctx).Infof("Getting runtime in provider tenant %s for labels %s: %s and %s: %s", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)
	runtime, err := v.runtimesSvc.GetByFilters(ctxWithProviderTenant, filters)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error while getting runtime in provider tenant %s for labels %s: %s and %s: %s: %v", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID, err)
		return errors.Wrapf(err, "failed to get runtime in tenant %s for labels %s: %s and %s: %s", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)
	}
	log.C(ctx).Infof("Found runtime with ID: %s in provider tenant %s for labels %s: %s and %s: %s", runtime.ID, providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)

	consumerInternalTenantID := claims.Tenant[tenantmapping.ConsumerTenantKey]
	consumerExternalTenantID := claims.Tenant[tenantmapping.ExternalTenantKey]
	ctxWithConsumerTenant := tenant.SaveToContext(ctx, consumerInternalTenantID, consumerExternalTenantID)

	rtmCtxFilter := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(v.consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", consumerExternalTenantID)),
	}

	log.C(ctx).Infof("Listing runtime context(s) in the consumer tenant %q for runtime with ID: %q and label with key: %q and value: %q", consumerExternalTenantID, runtime.ID, v.consumerSubaccountLabelKey, consumerExternalTenantID)
	rtmCtxPage, err := v.runtimeCtxSvc.ListByFilter(ctxWithConsumerTenant, runtime.ID, rtmCtxFilter, 100, "")
	if err != nil {
		log.C(ctx).Errorf("An error occurred while listing runtime context for runtime with ID: %q and filter with key: %q and value: %q", runtime.ID, v.consumerSubaccountLabelKey, consumerExternalTenantID)
		return errors.Wrapf(err, "while listing runtime context for runtime with ID: %q and filter with key: %q and value: %q", runtime.ID, v.consumerSubaccountLabelKey, consumerExternalTenantID)
	}
	log.C(ctx).Infof("Found %d runtime context(s) for runtime with ID: %q", len(rtmCtxPage.Data), runtime.ID)

	if len(rtmCtxPage.Data) == 0 {
		log.C(ctx).Errorf("Consumer's external tenant %s was not found as subscription record in the runtime context table for the runtime with ID: %s in the provider tenant %s", consumerExternalTenantID, runtime.ID, providerInternalTenantID)
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the runtime context table for the runtime in the provider tenant %s", consumerExternalTenantID, providerInternalTenantID))
	}

	return tx.Commit()
}

func (v *validator) validateApplicationProvider(ctx context.Context, claims Claims) error {
	tx, err := v.transact.Begin()
	if err != nil {
		log.C(ctx).Errorf("An error has occurred while opening transaction: %v", err)
		return errors.Wrapf(err, "An error has occurred while opening transaction")
	}
	defer v.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if len(claims.TokenClientID) == 0 {
		log.C(ctx).Errorf("Could not find consumer token client ID")
		return apperrors.NewUnauthorizedError("could not find consumer token client ID")
	}
	if len(claims.Region) == 0 {
		log.C(ctx).Errorf("Could not determine consumer token's region")
		return apperrors.NewUnauthorizedError("could not determine token's region")
	}

	tokenClientID := strings.TrimPrefix(claims.TokenClientID, v.tokenPrefix)
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(v.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", tokenClientID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", claims.Region)),
	}

	providerInternalTenantID := claims.Tenant[tenantmapping.ProviderTenantKey]
	providerExternalTenantID := claims.Tenant[tenantmapping.ProviderExternalTenantKey]
	ctxWithProviderTenant := tenant.SaveToContext(ctx, providerInternalTenantID, providerExternalTenantID)

	log.C(ctx).Infof("Get application template in provider tenant %s for labels %s: %s and %s: %s", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)
	applicationTemplate, err := v.appTemplateSvc.GetByFilters(ctxWithProviderTenant, filters)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error while getting application template in provider tenant %s for labels %s: %s and %s: %s: %v", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID, err)
		return errors.Wrapf(err, "failed to get application template in tenant %s for labels %s: %s and %s: %s", providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)
	}
	log.C(ctx).Infof("Found application template with ID %q in provider tenant %s for labels %s: %s and %s: %s", applicationTemplate.ID, providerInternalTenantID, tenant.RegionLabelKey, claims.Region, v.subscriptionProviderLabelKey, tokenClientID)

	consumerInternalTenantID := claims.Tenant[tenantmapping.ConsumerTenantKey]
	consumerExternalTenantID := claims.Tenant[tenantmapping.ExternalTenantKey]
	ctxWithConsumerTenant := tenant.SaveToContext(ctx, consumerInternalTenantID, consumerExternalTenantID)

	found := false

	log.C(ctx).Infof("Listing applications in the consumer tenant %q for application template with ID: %q and label with key: %q and value: %q", consumerExternalTenantID, applicationTemplate.ID, v.consumerSubaccountLabelKey, consumerExternalTenantID)
	applications, err := v.applicationSvc.ListAll(ctxWithConsumerTenant)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while listing applications for filter with key: %q and value: %q", v.consumerSubaccountLabelKey, consumerExternalTenantID)
		return errors.Wrapf(err, "while listing applications for filter with key: %q and value: %q", v.consumerSubaccountLabelKey, consumerExternalTenantID)
	}

	log.C(ctx).Infof("Found %d applications in consumer tenant using label: %q and external tenant ID: %q", len(applications), v.consumerSubaccountLabelKey, consumerExternalTenantID)

	for _, application := range applications {
		if *application.ApplicationTemplateID == applicationTemplate.ID {
			found = true
			break
		}
	}

	if !found {
		log.C(ctx).Errorf("Consumer's external tenant %s was not found as subscription record in the applications table for any application templates in the provider tenant %s", consumerExternalTenantID, providerInternalTenantID)
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the applications table for any application templates in the provider tenant %s", consumerExternalTenantID, providerInternalTenantID))
	}

	return tx.Commit()
}

func (v *validator) validateIntegrationSystemConsumer(ctx context.Context, claims Claims) error {
	if claims.Tenant[tenantmapping.ProviderExternalTenantKey] == claims.ConsumerID {
		return nil // consumer ID is a subaccount tenant
	}

	exists, err := v.intSystemSvc.Exists(ctx, claims.ConsumerID)
	if err != nil {
		return errors.Wrapf(err, "while checking if integration system with ID %s exists", claims.ConsumerID)
	}
	if !exists {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("integration system with ID %s does not exist", claims.ConsumerID))
	}

	return nil
}
