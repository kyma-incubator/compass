package subscription

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Config is configuration for the tenant subscription flow
type Config struct {
	ProviderLabelKey           string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	GlobalSubaccountIDLabelKey string `envconfig:"APP_GLOBAL_SUBACCOUNT_ID_LABEL_KEY,default=global_subaccount_id"`
	SubscriptionLabelKey       string `envconfig:"APP_SUBSCRIPTION_LABEL_KEY,default=subscription"`
	RuntimeTypeLabelKey        string `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
}

const (
	// SubdomainLabelKey is the key of the tenant label for subdomain.
	SubdomainLabelKey = "subdomain"
	// RegionPrefix a prefix to be trimmed from the region placeholder value when creating an app from template
	RegionPrefix = "cf-"
	// SubscriptionsLabelKey is the key of the subscriptions label, that stores the ids of created instances.
	SubscriptionsLabelKey = "subscriptions"
	// PreviousSubscriptionID represents a previous subscription id. This is needed, because before introducing this change there might be subscriptions which we don't know that they existed.
	PreviousSubscriptionID = "00000000-0000-0000-0000-000000000000"
)

// RuntimeService is responsible for Runtime operations
//
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	GetByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error)
	GetByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error)
}

// RuntimeCtxService provide functionality to interact with the runtime contexts(create, list, delete).
//
//go:generate mockery --name=RuntimeCtxService --output=automock --outpkg=automock --case=underscore
type RuntimeCtxService interface {
	Create(ctx context.Context, in model.RuntimeContextInput) (string, error)
	Delete(ctx context.Context, id string) error
	ListByFilter(ctx context.Context, runtimeID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error)
}

// TenantService provides functionality for retrieving, and creating tenants.
//
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --unroll-variadic=False --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

// LabelService is responsible updating already existing labels, and their label definitions.
//
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

// ApplicationTemplateService is responsible for Application Template operations
//
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Exists(ctx context.Context, id string) (bool, error)
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

// ApplicationTemplateConverter missing godoc
//
//go:generate mockery --name=ApplicationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateConverter interface {
	ApplicationFromTemplateInputFromGraphQL(appTemplate *model.ApplicationTemplate, in graphql.ApplicationFromTemplateInput) (model.ApplicationFromTemplateInput, error)
}

// ApplicationConverter is converting graphql and model Applications
//
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateRegisterInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error)
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
}

// ApplicationService is responsible for Application operations
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error)
	ListAll(ctx context.Context) ([]*model.Application, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	runtimeSvc                   RuntimeService
	runtimeCtxSvc                RuntimeCtxService
	tenantSvc                    TenantService
	labelSvc                     LabelService
	appTemplateSvc               ApplicationTemplateService
	appConv                      ApplicationConverter
	appTemplateConv              ApplicationTemplateConverter
	appSvc                       ApplicationService
	uidSvc                       uidService
	globalSubaccountIDLabelKey   string
	subscriptionLabelKey         string
	runtimeTypeLabelKey          string
	subscriptionProviderLabelKey string
}

// NewService returns a new object responsible for service-layer Subscription operations.
func NewService(runtimeSvc RuntimeService, runtimeCtxSvc RuntimeCtxService, tenantSvc TenantService, labelSvc LabelService, appTemplateSvc ApplicationTemplateService, appConv ApplicationConverter, appTemplateConv ApplicationTemplateConverter, appSvc ApplicationService, uidService uidService,
	globalSubaccountIDLabelKey, subscriptionLabelKey, runtimeTypeLabelKey, subscriptionProviderLabelKey string) *service {
	return &service{
		runtimeSvc:                   runtimeSvc,
		runtimeCtxSvc:                runtimeCtxSvc,
		tenantSvc:                    tenantSvc,
		labelSvc:                     labelSvc,
		appTemplateSvc:               appTemplateSvc,
		appConv:                      appConv,
		appTemplateConv:              appTemplateConv,
		appSvc:                       appSvc,
		uidSvc:                       uidService,
		globalSubaccountIDLabelKey:   globalSubaccountIDLabelKey,
		subscriptionLabelKey:         subscriptionLabelKey,
		runtimeTypeLabelKey:          runtimeTypeLabelKey,
		subscriptionProviderLabelKey: subscriptionProviderLabelKey,
	}
}

// SubscribeTenantToRuntime subscribes a tenant to runtimes by labeling the runtime
func (s *service) SubscribeTenantToRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID string) (bool, error) {
	log.C(ctx).Infof("Subscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q and application name: %q", consumerTenantID, subaccountTenantID, providerSubaccountID, subscriptionAppName)
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID from external ID: %q", providerSubaccountID)
	}
	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	log.C(ctx).Infof("Getting provider runtime in tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	runtime, err := s.runtimeSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtime for labels %q: %q and %q: %q", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	consumerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, subaccountTenantID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during subscription: %v", subaccountTenantID, err)
		return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subaccountTenantID)
	}

	runtimeID := runtime.ID
	log.C(ctx).Infof("Listing runtime context(s) in the consumer tenant %q for label with key: %q and value: %q", subaccountTenantID, s.globalSubaccountIDLabelKey, subaccountTenantID)
	rtmCtxPage, err := s.runtimeCtxSvc.ListByFilter(tenant.SaveToContext(ctx, consumerInternalTenant, subaccountTenantID), runtimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(s.globalSubaccountIDLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantID))}, 100, "")
	if err != nil {
		log.C(ctx).Errorf("An error occurred while listing runtime contexts with key: %q and value: %q for runtime with ID: %q: %v", s.globalSubaccountIDLabelKey, subaccountTenantID, runtimeID, err)
		return false, err
	}
	log.C(ctx).Infof("Found %d runtime context(s) with key: %q and value: %q for runtime with ID: %q", len(rtmCtxPage.Data), s.globalSubaccountIDLabelKey, subaccountTenantID, runtimeID)

	for _, rtmCtx := range rtmCtxPage.Data {
		if rtmCtx.Value == consumerTenantID {
			// Already subscribed
			log.C(ctx).Infof("Consumer %q is already subscribed. Adding the new value %q to the %q label", consumerTenantID, subscriptionID, SubscriptionsLabelKey)
			if err := s.manageSubscriptionsLabelOnSubscribe(ctx, consumerInternalTenant, model.RuntimeContextLabelableObject, rtmCtx.ID, subscriptionID); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, runtime.ID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting lowest owner for resource type: %q with ID: %q: %v", resource.Runtime, runtime.ID, err)
		return false, err
	}

	log.C(ctx).Debugf("Upserting runtime label with key: %q and value: %q", s.runtimeTypeLabelKey, subscriptionAppName)
	if err := s.labelSvc.UpsertLabel(ctx, tnt, &model.LabelInput{
		Key:        s.runtimeTypeLabelKey,
		Value:      subscriptionAppName,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtime.ID,
	}); err != nil {
		log.C(ctx).Errorf("An error occurred while upserting label with key: %q and value: %q for object type: %q and ID: %q: %v", s.runtimeTypeLabelKey, subscriptionAppName, model.RuntimeLabelableObject, runtime.ID, err)
		return false, err
	}

	ctx = tenant.SaveToContext(ctx, consumerInternalTenant, subaccountTenantID)

	m2mTable, ok := resource.Runtime.TenantAccessTable()
	if !ok {
		return false, errors.Errorf("entity %s does not have access table", resource.Runtime)
	}

	if err := repo.CreateTenantAccessRecursively(ctx, m2mTable, &repo.TenantAccess{
		TenantID:   consumerInternalTenant,
		ResourceID: runtime.ID,
		Owner:      false,
		Source:     consumerInternalTenant,
	}); err != nil {
		return false, err
	}

	rtmCtxID, err := s.runtimeCtxSvc.Create(ctx, model.RuntimeContextInput{
		Key:       s.subscriptionLabelKey,
		Value:     consumerTenantID,
		RuntimeID: runtime.ID,
	})
	if err != nil {
		log.C(ctx).Errorf("An error occurred while creating runtime context with key: %q and value: %q, and runtime ID: %q: %v", s.subscriptionLabelKey, consumerTenantID, runtime.ID, err)
		return false, errors.Wrapf(err, "while creating runtime context with value: %q and runtime ID: %q during subscription", consumerTenantID, runtime.ID)
	}

	log.C(ctx).Infof("Creating label for runtime context with ID: %q with key: %q and value: %q", rtmCtxID, s.globalSubaccountIDLabelKey, subaccountTenantID)
	if err := s.labelSvc.CreateLabel(ctx, consumerInternalTenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        s.globalSubaccountIDLabelKey,
		Value:      subaccountTenantID,
		ObjectID:   rtmCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}); err != nil {
		log.C(ctx).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q: %v", s.globalSubaccountIDLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID, err)
		return false, errors.Wrap(err, fmt.Sprintf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", s.globalSubaccountIDLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID))
	}

	log.C(ctx).Infof("Creating label for runtime context with ID: %q with key: %q and value: %q", rtmCtxID, SubscriptionsLabelKey, subscriptionID)
	if err := s.labelSvc.CreateLabel(ctx, consumerInternalTenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        SubscriptionsLabelKey,
		Value:      []string{subscriptionID},
		ObjectID:   rtmCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}); err != nil {
		log.C(ctx).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q: %v", SubscriptionsLabelKey, subscriptionID, model.RuntimeContextLabelableObject, rtmCtxID, err)
		return false, errors.Wrap(err, fmt.Sprintf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", SubscriptionsLabelKey, subscriptionID, model.RuntimeContextLabelableObject, rtmCtxID))
	}

	return true, nil
}

// UnsubscribeTenantFromRuntime unsubscribes a tenant from runtimes by removing labels from runtime
func (s *service) UnsubscribeTenantFromRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID string) (bool, error) {
	log.C(ctx).Infof("Unsubscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q", consumerTenantID, subaccountTenantID, providerSubaccountID)
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID from external ID: %q", providerSubaccountID)
	}
	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	log.C(ctx).Infof("Getting provider runtime in tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	runtime, err := s.runtimeSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtime for labels %q: %q and %q: %q", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	consumerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, subaccountTenantID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during subscription: %v", subaccountTenantID, err)
		return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subaccountTenantID)
	}
	ctx = tenant.SaveToContext(ctx, consumerInternalTenant, subaccountTenantID)

	runtimeID := runtime.ID
	log.C(ctx).Infof("Listing runtime context(s) in the consumer tenant %q for label with key: %q and value: %q", subaccountTenantID, s.globalSubaccountIDLabelKey, subaccountTenantID)
	rtmCtxPage, err := s.runtimeCtxSvc.ListByFilter(ctx, runtimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(s.globalSubaccountIDLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantID))}, 100, "")
	if err != nil {
		log.C(ctx).Errorf("An error occurred while listing runtime contexts with key: %q and value: %q for runtime with ID: %q: %v", s.globalSubaccountIDLabelKey, subaccountTenantID, runtimeID, err)
		return false, err
	}
	log.C(ctx).Infof("Found %d runtime context(s) with key: %q and value: %q for runtime with ID: %q", len(rtmCtxPage.Data), s.globalSubaccountIDLabelKey, subaccountTenantID, runtimeID)

	for _, rtmCtx := range rtmCtxPage.Data {
		// if the current subscription(runtime context) is the one for which the unsubscribe request is initiated, delete the record from the DB
		if rtmCtx.Value == consumerTenantID {
			if err := s.deleteOnUnsubscribe(ctx, consumerInternalTenant, model.RuntimeContextLabelableObject, rtmCtx.ID, subscriptionID, s.runtimeCtxSvc.Delete); err != nil {
				return false, err
			}
			break
		}
	}

	return true, nil
}

// SubscribeTenantToApplication fetches model.ApplicationTemplate by region and provider and registers an Application from that template
func (s *service) SubscribeTenantToApplication(ctx context.Context, providerID, subscribedSubaccountID, providerSubaccountID, consumerTenantID, region, subscribedAppName, subscriptionID string, subscriptionPayload string) (bool, string, string, error) {
	log.C(ctx).Infof("Subscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q and application name: %q", consumerTenantID, subscribedSubaccountID, providerSubaccountID, subscribedAppName)
	filters := s.buildLabelFilters(providerID, region)
	log.C(ctx).Infof("Getting provider application template in tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, "", "", nil
		}

		return false, "", "", errors.Wrapf(err, "while getting application template with filter labels %q and %q", providerID, region)
	}

	consumerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, subscribedSubaccountID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during application subscription: %v", subscribedSubaccountID, err)
		return false, "", "", errors.Wrapf(err, "while getting tenant with external ID: %q", subscribedSubaccountID)
	}

	ctx = tenant.SaveToContext(ctx, consumerInternalTenant, subscribedSubaccountID)

	applications, err := s.appSvc.ListAll(ctx)
	if err != nil {
		return false, "", "", errors.Wrapf(err, "while listing applications")
	}

	for _, app := range applications {
		if str.PtrStrToStr(app.ApplicationTemplateID) == appTemplate.ID {
			// Already subscribed
			log.C(ctx).Infof("Consumer %q is already subscribed. Adding the new value %q to the %q label", consumerTenantID, subscriptionID, SubscriptionsLabelKey)
			if err := s.manageSubscriptionsLabelOnSubscribe(ctx, consumerInternalTenant, model.ApplicationLabelableObject, app.ID, subscriptionID); err != nil {
				return false, "", "", err
			}
			return true, "", "", nil
		}
	}

	subdomainLabel, err := s.labelSvc.GetByKey(ctx, consumerInternalTenant, model.TenantLabelableObject, consumerInternalTenant, SubdomainLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return false, "", "", errors.Wrapf(err, "while getting label %q for %q with id %q", SubdomainLabelKey, model.TenantLabelableObject, consumerInternalTenant)
		}
	}

	subdomainValue := ""
	if subdomainLabel != nil && subdomainLabel.Value != nil {
		if subdomainLabelValue, ok := subdomainLabel.Value.(string); ok {
			subdomainValue = subdomainLabelValue
		}
	}

	appID, err := s.createApplicationFromTemplate(ctx, appTemplate, subscribedSubaccountID, consumerTenantID, subscribedAppName, subdomainValue, region, subscriptionID, subscriptionPayload)
	if err != nil {
		return false, "", "", err
	}

	return true, appID, appTemplate.ID, nil
}

// UnsubscribeTenantFromApplication fetches model.ApplicationTemplate by region and provider, lists all applications for
// the subscribedSubaccountID tenant and deletes them synchronously
func (s *service) UnsubscribeTenantFromApplication(ctx context.Context, providerID, subscribedSubaccountID, providerSubaccountID, consumerTenantID, region, subscriptionID string) (bool, error) {
	log.C(ctx).Infof("Unsubscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q", consumerTenantID, subscribedSubaccountID, providerSubaccountID)
	filters := s.buildLabelFilters(providerID, region)
	log.C(ctx).Infof("Getting provider application template in tenant %q for labels %q: %q and %q: %q", providerSubaccountID, tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID)
	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrapf(err, "while getting application template with filter labels %q and %q", providerID, region)
	}

	consumerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, subscribedSubaccountID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during application unsubscription: %v", subscribedSubaccountID, err)
		return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subscribedSubaccountID)
	}

	ctx = tenant.SaveToContext(ctx, consumerInternalTenant, subscribedSubaccountID)

	if err := s.deleteApplicationsByAppTemplateID(ctx, appTemplate.ID, subscriptionID); err != nil {
		return false, err
	}

	return true, nil
}

// DetermineSubscriptionFlow determines if the subscription flow is resource.ApplicationTemplate or resource.Runtime
// by fetching both resources by provider and region
func (s *service) DetermineSubscriptionFlow(ctx context.Context, providerID, region string) (resource.Type, error) {
	filters := s.buildLabelFilters(providerID, region)
	runtime, err := s.runtimeSvc.GetByFiltersGlobal(ctx, filters)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return "", errors.Wrapf(err, "while getting runtime with filter labels provider (%q) and region (%q)", providerID, region)
		}
	}

	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return "", errors.Wrapf(err, "while getting app template with filter labels provider (%q) and region (%q)", providerID, region)
		}
	}

	if runtime != nil && appTemplate == nil {
		return resource.Runtime, nil
	}

	if runtime == nil && appTemplate != nil {
		return resource.ApplicationTemplate, nil
	}

	if runtime == nil && appTemplate == nil {
		return "", nil
	}

	if runtime != nil && appTemplate != nil {
		return "", errors.Errorf("both a runtime (%+v) and application template (%+v) exist with filter labels provider (%q) and region (%q)", runtime, appTemplate, providerID, region)
	}

	return "", errors.Errorf("could not determine flow")
}

func (s *service) createApplicationFromTemplate(ctx context.Context, appTemplate *model.ApplicationTemplate, subscribedSubaccountID, consumerTenantID, subscribedAppName, subdomain, region, subscriptionID string, subscriptionPayload string) (string, error) {
	log.C(ctx).Debugf("Preparing Values for Application Template with name %q", appTemplate.Name)
	values, err := s.preparePlaceholderValues(appTemplate, subdomain, region, subscriptionPayload)
	if err != nil {
		return "", errors.Wrapf(err, "while preparing the values for Application template %q", appTemplate.Name)
	}

	log.C(ctx).Debugf("Preparing ApplicationCreateInput JSON from Application Template with name %q", appTemplate.Name)
	appCreateInputJSON, err := s.appTemplateSvc.PrepareApplicationCreateInputJSON(appTemplate, values)
	if err != nil {
		return "", errors.Wrapf(err, "while preparing ApplicationCreateInput JSON from Application Template with name %q", appTemplate.Name)
	}

	log.C(ctx).Debugf("Converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	appCreateInputGQL, err := s.appConv.CreateRegisterInputJSONToGQL(appCreateInputJSON)
	if err != nil {
		return "", errors.Wrapf(err, "while converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	}

	log.C(ctx).Infof("Validating GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	if err := inputvalidation.Validate(appCreateInputGQL); err != nil {
		return "", errors.Wrapf(err, "while validating application input from Application Template with name %q", appTemplate.Name)
	}

	appCreateInputModel, err := s.appConv.CreateInputFromGraphQL(ctx, appCreateInputGQL)
	if err != nil {
		return "", errors.Wrap(err, "while converting ApplicationFromTemplate input")
	}

	if appCreateInputModel.Labels == nil {
		appCreateInputModel.Labels = make(map[string]interface{})
	}
	appCreateInputModel.Labels["managed"] = "false"
	appCreateInputModel.Labels[SubscriptionsLabelKey] = []string{subscriptionID}
	appCreateInputModel.Labels[s.globalSubaccountIDLabelKey] = subscribedSubaccountID
	if appCreateInputModel.LocalTenantID == nil {
		appCreateInputModel.LocalTenantID = &consumerTenantID
	}

	log.C(ctx).Infof("Creating an Application with name %q from Application Template with name %q", subscribedAppName, appTemplate.Name)
	appID, err := s.appSvc.CreateFromTemplate(ctx, appCreateInputModel, &appTemplate.ID)
	if err != nil {
		return "", errors.Wrapf(err, "while creating an Application with name %s from Application Template with name %s", subscribedAppName, appTemplate.Name)
	}

	log.C(ctx).Infof("Successfully created an Application with id %q and name %q from Application Template with name %q", appID, subscribedAppName, appTemplate.Name)
	return appID, nil
}

func (s *service) preparePlaceholderValues(appTemplate *model.ApplicationTemplate, subdomain, region string, subscriptionPayload string) ([]*model.ApplicationTemplateValueInput, error) {
	values := []*model.ApplicationTemplateValueInput{
		{Placeholder: "subdomain", Value: subdomain},
		{Placeholder: "region", Value: strings.TrimPrefix(region, RegionPrefix)},
	}

	oldPlaceholders := appTemplate.Placeholders

	newPlaceholders := []model.ApplicationTemplatePlaceholder{}
	for _, placeholder := range oldPlaceholders {
		if placeholder.Name != "subdomain" && placeholder.Name != "region" {
			newPlaceholders = append(newPlaceholders, placeholder)
		}
	}
	appTemplate.Placeholders = newPlaceholders

	appFromTemplateInput, err := s.appTemplateConv.ApplicationFromTemplateInputFromGraphQL(appTemplate, graphql.ApplicationFromTemplateInput{
		TemplateName:        appTemplate.Name,
		PlaceholdersPayload: &subscriptionPayload,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "while parsing the callback payload with the Application template %q", appTemplate.Name)
	}

	appTemplate.Placeholders = oldPlaceholders
	values = append(appFromTemplateInput.Values, values...)
	return values, nil
}

func (s *service) deleteApplicationsByAppTemplateID(ctx context.Context, appTemplateID, subscriptionID string) error {
	applications, err := s.appSvc.ListAll(ctx)
	if err != nil {
		return errors.Wrapf(err, "while listing applications")
	}

	for _, app := range applications {
		if str.PtrStrToStr(app.ApplicationTemplateID) == appTemplateID {
			internalTenant, err := tenant.LoadFromContext(ctx)
			if err != nil {
				return errors.Wrapf(err, "An error occurred while loading tenant from context")
			}
			if err := s.deleteOnUnsubscribe(ctx, internalTenant, model.ApplicationLabelableObject, app.ID, subscriptionID, s.appSvc.Delete); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *service) manageSubscriptionsLabelOnSubscribe(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, subscriptionID string) error {
	subscriptionsLabel, err := s.labelSvc.GetByKey(ctx, tenant, objectType, objectID, SubscriptionsLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting label with key: %q for object type: %q and ID: %q", SubscriptionsLabelKey, objectType, objectID)
			return errors.Wrapf(err, "while getting label with key: %q for object type: %q and ID: %q", SubscriptionsLabelKey, objectType, objectID)
		}
		log.C(ctx).Infof("Creating label with key: %q and value: %q for %q with id: %q", subscriptionIDKey, []string{PreviousSubscriptionID, subscriptionID}, objectType, objectID)
		if err := s.labelSvc.CreateLabel(ctx, tenant, s.uidSvc.Generate(), &model.LabelInput{
			Key:        SubscriptionsLabelKey,
			Value:      []string{PreviousSubscriptionID, subscriptionID},
			ObjectID:   objectID,
			ObjectType: objectType,
		}); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", SubscriptionsLabelKey, []string{subscriptionID}, objectType, objectID)
			return errors.Wrapf(err, "while creating label with key: %q and value: %q for object type: %q and ID: %q", SubscriptionsLabelKey, []string{subscriptionID}, objectType, objectID)
		}

		log.C(ctx).Infof("%q label created, for already subscibed tenant to %q with id %q", SubscriptionsLabelKey, objectType, objectID)
		return nil
	}

	subscriptions, ok := subscriptionsLabel.Value.([]interface{})
	if !ok {
		return errors.Errorf("cannot cast %q label value of type %T to array", SubscriptionsLabelKey, subscriptionsLabel.Value)
	}
	if subscriptionExists(subscriptions, subscriptionID) {
		log.C(ctx).Infof("Subscription with id %q for %q with id %q already exists", subscriptionID, objectType, objectID)
		return nil
	}

	log.C(ctx).Infof("Adding new subscription ID: %q to the %q label. Current subscription IDs %v", subscriptionID, SubscriptionsLabelKey, subscriptions)
	subscriptions = append(subscriptions, subscriptionID)
	if err := s.labelSvc.UpdateLabel(ctx, tenant, subscriptionsLabel.ID, &model.LabelInput{
		Key:        subscriptionsLabel.Key,
		Value:      subscriptions,
		ObjectID:   subscriptionsLabel.ObjectID,
		ObjectType: subscriptionsLabel.ObjectType,
		Version:    subscriptionsLabel.Version,
	}); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while updating label with key: %q and value: %v for object type: %q and ID: %q", SubscriptionsLabelKey, subscriptions, objectType, objectID)
		return errors.Wrapf(err, "while updating label with key: %q and value: %v for object type: %q and ID: %q", SubscriptionsLabelKey, subscriptions, objectType, objectID)
	}

	log.C(ctx).Infof("Successfully added the new value %q to the label %q for %q with id %q", subscriptionID, SubscriptionsLabelKey, objectType, objectID)
	return nil
}

func (s *service) deleteOnUnsubscribe(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, subscriptionID string, deleteObject func(context.Context, string) error) error {
	subscriptionsLabel, err := s.labelSvc.GetByKey(ctx, tenant, objectType, objectID, SubscriptionsLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting label with key: %q for object type: %q and ID: %q", SubscriptionsLabelKey, objectType, objectID)
			return errors.Wrapf(err, "while getting label with key: %q for object type: %q and ID: %q", SubscriptionsLabelKey, objectType, objectID)
		}

		log.C(ctx).Debugf("Cannot find label with key %q for %q with ID %q. Triggering deletion of %q with ID %q...", SubscriptionsLabelKey, objectType, objectID, objectType, objectID)
		if err := deleteObject(ctx, objectID); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while trying to delete %q with ID: %q", objectType, objectID)
			return errors.Wrapf(err, "while trying to delete %q with ID: %q", objectType, objectID)
		}
		log.C(ctx).Infof("Successfully deleted %q with ID %q", objectType, objectID)
		return nil
	}

	subscriptions, ok := subscriptionsLabel.Value.([]interface{})
	if !ok {
		return errors.Errorf("cannot cast %q label value of type %T to array", SubscriptionsLabelKey, subscriptionsLabel.Value)
	}

	if len(subscriptions) <= 1 {
		log.C(ctx).Debugf("The number of %q for %q with ID %q is <=1. Triggering deletion of %q with ID %q...", SubscriptionsLabelKey, objectType, objectID, objectType, objectID)
		if err := deleteObject(ctx, objectID); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while deleting %q with ID: %q", objectType, objectID)
			return errors.Wrapf(err, "while deleting %q with ID: %q", objectType, objectID)
		}
		log.C(ctx).Infof("Successfully deleted %q with ID %q", objectType, objectID)
		return nil
	}

	subscriptions, removed := removeSubscription(subscriptions, subscriptionID)
	if !removed {
		log.C(ctx).Infof("Subscription with id %q does not exist. No need to update %q label value", subscriptionID, SubscriptionsLabelKey)
		return nil
	}
	if err := s.labelSvc.UpdateLabel(ctx, tenant, subscriptionsLabel.ID, &model.LabelInput{
		Key:        subscriptionsLabel.Key,
		Value:      subscriptions,
		ObjectID:   subscriptionsLabel.ObjectID,
		ObjectType: subscriptionsLabel.ObjectType,
		Version:    subscriptionsLabel.Version,
	}); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while updating label with key: %q and value: %v for object type: %q and ID: %q", SubscriptionsLabelKey, subscriptions, objectType, objectID)
		return errors.Wrapf(err, "while updating label with key: %q and value: %v for object type: %q and ID: %q", SubscriptionsLabelKey, subscriptions, objectType, objectID)
	}
	log.C(ctx).Infof("Successfully removed value %q from the label %q for %q with ID %q", subscriptionID, SubscriptionsLabelKey, objectType, objectID)

	return nil
}

func (s *service) buildLabelFilters(subscriptionProviderID, region string) []*labelfilter.LabelFilter {
	return []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}
}

func subscriptionExists(subscriptions []interface{}, subscriptionID string) bool {
	for _, id := range subscriptions {
		if id == subscriptionID {
			return true
		}
	}
	return false
}

func removeSubscription(subscriptions []interface{}, subscriptionID string) ([]interface{}, bool) {
	if subscriptionExists(subscriptions, subscriptionID) {
		return remove(subscriptions, subscriptionID), true
	}
	if subscriptionExists(subscriptions, PreviousSubscriptionID) {
		return remove(subscriptions, PreviousSubscriptionID), true
	}
	return subscriptions, false
}

// remove removes in place the subscriptionID
func remove(subscriptions []interface{}, subscriptionID string) []interface{} {
	writeIdx := 0

	for _, id := range subscriptions {
		if id != subscriptionID {
			subscriptions[writeIdx] = id
			writeIdx++
		}
	}
	return subscriptions[:writeIdx]
}
