package subscription

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// todo:: delete the whole file!! Double check the changes
// Config is configuration for the tenant-runtime subscription flow
type Config struct {
	ProviderLabelKey                    string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountLabelKey          string `envconfig:"APP_CONSUMER_SUBACCOUNT_LABEL_KEY,default=consumer_subaccount_id"`
	SubscriptionLabelKey                string `envconfig:"APP_SUBSCRIPTION_LABEL_KEY,default=subscription"`
	SubscriptionProviderAppNameLabelKey string `envconfig:"APP_SUBSCRIPTION_PROVIDER_APP_NAME_LABEL_KEY,default=runtimeType"`
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	ListByFilters(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// RuntimeCtxService provide functionality to interact with the runtime contexts(create, list, delete).
//go:generate mockery --name=RuntimeCtxService --output=automock --outpkg=automock --case=underscore
type RuntimeCtxService interface {
	Create(ctx context.Context, in model.RuntimeContextInput) (string, error)
	Delete(ctx context.Context, id string) error
	ListByFilter(ctx context.Context, runtimeID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error)
}

// TenantService provides functionality for retrieving, and creating tenants.
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --unroll-variadic=False --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetTenantByExternalID(ctx context.Context, externalTenantID string) (*model.BusinessTenantMapping, error)
}

// LabelService is responsible updating already existing labels, and their label definitions.
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

type service struct {
	runtimeSvc                   RuntimeService
	runtimeCtxSvc                RuntimeCtxService
	tenantSvc                    TenantService
	labelSvc                     LabelService
	uidSvc                       uidService
	consumerSubaccountLabelKey   string
	subscriptionLabelKey         string
	appNameLabelKey              string
	subscriptionProviderLabelKey string
}

// NewService missing godoc
func NewService(runtimeSvc RuntimeService, runtimeCtxSvc RuntimeCtxService, tenantSvc TenantService, labelSvc LabelService, uidService uidService,
	consumerSubaccountLabelKey, subscriptionLabelKey, appNameLabelKey, subscriptionProviderLabelKey string) *service {
	return &service{
		runtimeSvc:                   runtimeSvc,
		runtimeCtxSvc:                runtimeCtxSvc,
		tenantSvc:                    tenantSvc,
		labelSvc:                     labelSvc,
		uidSvc:                       uidService,
		consumerSubaccountLabelKey:   consumerSubaccountLabelKey,
		subscriptionLabelKey:         subscriptionLabelKey,
		appNameLabelKey:              appNameLabelKey,
		subscriptionProviderLabelKey: subscriptionProviderLabelKey,
	}
}

func (s *service) SubscribeTenant(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, appNameLabel string) (bool, error) {
	log.C(ctx).Infof("Subscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q", consumerTenantID, subaccountTenantID, providerSubaccountID)
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", providerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	providerTenant, err := s.tenantSvc.GetTenantByExternalID(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID from external ID: %q", providerSubaccountID)
	}
	ctx = tenant.SaveToContext(ctx, providerTenant.ID, providerSubaccountID)

	log.C(ctx).Infof("Listing runtimes in provider tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	runtimes, err := s.runtimeSvc.ListByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %q: %q and %q: %q", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}
	log.C(ctx).Infof("Found %d provider runtime(s) during subscription", len(runtimes))

	for _, runtime := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, runtime.ID)
		if err != nil {
			log.C(ctx).Errorf("An error occurred while getting lowest owner for resource type: %q with ID: %q: %v", resource.Runtime, runtime.ID, err)
			return false, err
		}

		if err := s.labelSvc.UpsertLabel(ctx, tnt, &model.LabelInput{
			Key:        s.appNameLabelKey,
			Value:      appNameLabel,
			ObjectType: model.RuntimeLabelableObject,
			ObjectID:   runtime.ID,
		}); err != nil {
			log.C(ctx).Errorf("An error occurred while upserting label with key: %q and value: %q for object type: %q and ID: %q: %v", s.appNameLabelKey, appNameLabel, model.RuntimeLabelableObject, runtime.ID, err)
			return false, err
		}

		tenantMapping, err := s.tenantSvc.GetTenantByExternalID(ctx, subaccountTenantID)
		if err != nil {
			log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during subscription: %v", subaccountTenantID, err)
			return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subaccountTenantID)
		}

		ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)

		rtmCtxID, err := s.runtimeCtxSvc.Create(ctx, model.RuntimeContextInput{
			Key:       s.subscriptionLabelKey,
			Value:     consumerTenantID,
			RuntimeID: runtime.ID,
		})
		if err != nil {
			log.C(ctx).Errorf("An error occurred while creating runtime context with key: %q and value: %q, and runtime ID: %q: %v", s.subscriptionLabelKey, consumerTenantID, runtime.ID, err)
			return false, errors.Wrapf(err, "while creating runtime context with value: %q and runtime ID: %q during subscription", consumerTenantID, runtime.ID)
		}

		m2mTable, ok := resource.Runtime.TenantAccessTable()
		if !ok {
			return false, errors.Errorf("entity %s does not have access table", resource.RuntimeContext)
		}

		if err := repo.CreateTenantAccessRecursively(ctx, m2mTable, &repo.TenantAccess{
			TenantID:   tenantMapping.ID,
			ResourceID: runtime.ID,
			Owner:      false,
		}); err != nil {
			return false, err
		}

		if err := s.labelSvc.CreateLabel(ctx, tenantMapping.ID, s.uidSvc.Generate(), &model.LabelInput{
			Key:        s.consumerSubaccountLabelKey,
			Value:      subaccountTenantID,
			ObjectID:   rtmCtxID,
			ObjectType: model.RuntimeContextLabelableObject,
		}); err != nil {
			log.C(ctx).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q: %v", s.consumerSubaccountLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID, err)
			return false, errors.Wrap(err, fmt.Sprintf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", s.consumerSubaccountLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID))
		}
	}
	return true, nil
}

func (s *service) UnsubscribeTenant(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region string) (bool, error) {
	log.C(ctx).Infof("Unsubscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q", consumerTenantID, subaccountTenantID, providerSubaccountID)
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", providerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	providerTenant, err := s.tenantSvc.GetTenantByExternalID(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID from external ID: %q", providerSubaccountID)
	}
	ctx = tenant.SaveToContext(ctx, providerTenant.ID, providerSubaccountID)

	log.C(ctx).Infof("Listing runtimes in provider tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	runtimes, err := s.runtimeSvc.ListByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %q: %q and %q: %q", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}
	log.C(ctx).Infof("Found %d provider runtime(s) during unsubscribe", len(runtimes))

	tenantMapping, err := s.tenantSvc.GetTenantByExternalID(ctx, subaccountTenantID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during subscription: %v", subaccountTenantID, err)
		return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subaccountTenantID)
	}

	ctx = tenant.SaveToContext(ctx, tenantMapping.ID, tenantMapping.ExternalTenant)

	for _, runtime := range runtimes {
		runtimeID := runtime.ID
		log.C(ctx).Infof("Listing runtime context(s) in the consumer tenant %q for label with key: %q and value: %q", subaccountTenantID, s.consumerSubaccountLabelKey, subaccountTenantID)
		rtmCtxPage, err := s.runtimeCtxSvc.ListByFilter(ctx, runtimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(s.consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantID))}, 100, "")
		if err != nil {
			log.C(ctx).Errorf("An error occurred while listing runtime contexts with key: %q and value: %q for runtime with ID: %q: %v", s.consumerSubaccountLabelKey, subaccountTenantID, runtimeID, err)
			return false, err
		}
		log.C(ctx).Infof("Found %d runtime context(s) with key: %q and value: %q for runtime with ID: %q", len(rtmCtxPage.Data), s.consumerSubaccountLabelKey, subaccountTenantID, runtimeID)

		for _, rtmCtx := range rtmCtxPage.Data {
			// if the current subscription(runtime context) is the one for which the unsubscribe request is initiated, delete the record from the DB
			if rtmCtx.Value == consumerTenantID {
				log.C(ctx).Infof("Deleting runtime context with key: %q and value: %q for runtime ID: %q", rtmCtx.Key, rtmCtx.Value, runtimeID)
				if err := s.runtimeCtxSvc.Delete(ctx, rtmCtx.ID); err != nil {
					log.C(ctx).Errorf("An error occurred while deleting runtime context with key: %q and value: %q for runtime ID: %q", rtmCtx.Key, rtmCtx.Value, runtimeID)
					return false, err
				}
				log.C(ctx).Infof("Successfully deleted runtime context with key: %q and value: %q for runtime ID: %q", rtmCtx.Key, rtmCtx.Value, runtimeID)
				break
			}
		}
	}

	return true, nil
}
