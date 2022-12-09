package subscription

import (
	"context"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

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
	ConsumerSubaccountLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_LABEL_KEY,default=global_subaccount_id"`
	SubscriptionLabelKey       string `envconfig:"APP_SUBSCRIPTION_LABEL_KEY,default=subscription"`
	RuntimeTypeLabelKey        string `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
}

const (
	// SubdomainLabelKey is the key of the tenant label for subdomain.
	SubdomainLabelKey = "subdomain"
	// RegionPrefix a prefix to be trimmed from the region placeholder value when creating an app from template
	RegionPrefix = "cf-"
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

// ApplicationConverter is converting graphql and model Applications
//
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error)
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
	appSvc                       ApplicationService
	uidSvc                       uidService
	consumerSubaccountLabelKey   string
	subscriptionLabelKey         string
	runtimeTypeLabelKey          string
	subscriptionProviderLabelKey string
}

// NewService returns a new object responsible for service-layer Subscription operations.
func NewService(runtimeSvc RuntimeService, runtimeCtxSvc RuntimeCtxService, tenantSvc TenantService, labelSvc LabelService, appTemplateSvc ApplicationTemplateService, appConv ApplicationConverter, appSvc ApplicationService, uidService uidService,
	consumerSubaccountLabelKey, subscriptionLabelKey, runtimeTypeLabelKey, subscriptionProviderLabelKey string) *service {
	return &service{
		runtimeSvc:                   runtimeSvc,
		runtimeCtxSvc:                runtimeCtxSvc,
		tenantSvc:                    tenantSvc,
		labelSvc:                     labelSvc,
		appTemplateSvc:               appTemplateSvc,
		appConv:                      appConv,
		appSvc:                       appSvc,
		uidSvc:                       uidService,
		consumerSubaccountLabelKey:   consumerSubaccountLabelKey,
		subscriptionLabelKey:         subscriptionLabelKey,
		runtimeTypeLabelKey:          runtimeTypeLabelKey,
		subscriptionProviderLabelKey: subscriptionProviderLabelKey,
	}
}

// SubscribeTenantToRuntime subscribes a tenant to runtimes by labeling the runtime
func (s *service) SubscribeTenantToRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName string) (bool, error) {
	log.C(ctx).Infof("Subscribe request is triggerred between consumer with tenant: %q and subaccount: %q and provider with subaccount: %q", consumerTenantID, subaccountTenantID, providerSubaccountID)
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
	log.C(ctx).Infof("Listing runtime context(s) in the consumer tenant %q for label with key: %q and value: %q", subaccountTenantID, s.consumerSubaccountLabelKey, subaccountTenantID)
	rtmCtxPage, err := s.runtimeCtxSvc.ListByFilter(tenant.SaveToContext(ctx, consumerInternalTenant, subaccountTenantID), runtimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(s.consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantID))}, 100, "")
	if err != nil {
		log.C(ctx).Errorf("An error occurred while listing runtime contexts with key: %q and value: %q for runtime with ID: %q: %v", s.consumerSubaccountLabelKey, subaccountTenantID, runtimeID, err)
		return false, err
	}
	log.C(ctx).Infof("Found %d runtime context(s) with key: %q and value: %q for runtime with ID: %q", len(rtmCtxPage.Data), s.consumerSubaccountLabelKey, subaccountTenantID, runtimeID)

	for _, rtmCtx := range rtmCtxPage.Data {
		if rtmCtx.Value == consumerTenantID {
			// Already subscribed
			log.C(ctx).Infof("Consumer %q is already subscribed", consumerTenantID)
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

	if err := repo.UpsertTenantAccessRecursively(ctx, m2mTable, &repo.TenantAccess{
		TenantID:   consumerInternalTenant,
		ResourceID: runtime.ID,
		Owner:      false,
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

	log.C(ctx).Infof("Creating label for runtime context with ID: %q with key: %q and value: %q", rtmCtxID, s.consumerSubaccountLabelKey, subaccountTenantID)
	if err := s.labelSvc.CreateLabel(ctx, consumerInternalTenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        s.consumerSubaccountLabelKey,
		Value:      subaccountTenantID,
		ObjectID:   rtmCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}); err != nil {
		log.C(ctx).Errorf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q: %v", s.consumerSubaccountLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID, err)
		return false, errors.Wrap(err, fmt.Sprintf("An error occurred while creating label with key: %q and value: %q for object type: %q and ID: %q", s.consumerSubaccountLabelKey, subaccountTenantID, model.RuntimeContextLabelableObject, rtmCtxID))
	}

	return true, nil
}

// UnsubscribeTenantFromRuntime unsubscribes a tenant from runtimes by removing labels from runtime
func (s *service) UnsubscribeTenantFromRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region string) (bool, error) {
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

	return true, nil
}

// SubscribeTenantToApplication fetches model.ApplicationTemplate by region and provider and registers an Application from that template
func (s *service) SubscribeTenantToApplication(ctx context.Context, providerID, subscribedSubaccountID, consumerTenantID, region, subscribedAppName string, subscriptionPayload *string) (bool, error) {
	filters := s.buildLabelFilters(providerID, region)
	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrapf(err, "while getting application template with filter labels %q and %q", providerID, region)
	}

	consumerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, subscribedSubaccountID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting tenant by external ID: %q during application subscription: %v", subscribedSubaccountID, err)
		return false, errors.Wrapf(err, "while getting tenant with external ID: %q", subscribedSubaccountID)
	}

	ctx = tenant.SaveToContext(ctx, consumerInternalTenant, subscribedSubaccountID)

	applications, err := s.appSvc.ListAll(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while listing applications")
	}

	for _, app := range applications {
		if str.PtrStrToStr(app.ApplicationTemplateID) == appTemplate.ID {
			// Already subscribed
			return true, nil
		}
	}

	subdomainLabel, err := s.labelSvc.GetByKey(ctx, consumerInternalTenant, model.TenantLabelableObject, consumerInternalTenant, SubdomainLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return false, errors.Wrapf(err, "while getting label %q for %q with id %q", SubdomainLabelKey, model.TenantLabelableObject, consumerInternalTenant)
		}
	}

	subdomainValue := ""
	if subdomainLabel != nil && subdomainLabel.Value != nil {
		if subdomainLabelValue, ok := subdomainLabel.Value.(string); ok {
			subdomainValue = subdomainLabelValue
		}
	}

	if err := s.createApplicationFromTemplate(ctx, appTemplate, subscribedSubaccountID, consumerTenantID, subscribedAppName, subdomainValue, region, subscriptionPayload); err != nil {
		return false, err
	}

	return true, nil
}

// UnsubscribeTenantFromApplication fetches model.ApplicationTemplate by region and provider, lists all applications for
// the subscribedSubaccountID tenant and deletes them synchronously
func (s *service) UnsubscribeTenantFromApplication(ctx context.Context, providerID, subscribedSubaccountID, region string) (bool, error) {
	filters := s.buildLabelFilters(providerID, region)
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

	if err := s.deleteApplicationsByAppTemplateID(ctx, appTemplate.ID); err != nil {
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

func (s *service) createApplicationFromTemplate(ctx context.Context, appTemplate *model.ApplicationTemplate, subscribedSubaccountID, consumerTenantID, subscribedAppName, subdomain, region string, subscriptionPayload *string) error {
	values := []*model.ApplicationTemplateValueInput{
		{Placeholder: "subdomain", Value: subdomain},
		{Placeholder: "region", Value: strings.TrimPrefix(region, RegionPrefix)},
	}
	values = processAdditionalValues(subscriptionPayload, appTemplate, values, ctx, subscribedAppName)
	log.C(ctx).Debugf("Preparing ApplicationCreateInput JSON from Application Template with name %q", appTemplate.Name)
	appCreateInputJSON, err := s.appTemplateSvc.PrepareApplicationCreateInputJSON(appTemplate, values)
	if err != nil {
		return errors.Wrapf(err, "while preparing ApplicationCreateInput JSON from Application Template with name %q", appTemplate.Name)
	}

	log.C(ctx).Debugf("Converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	appCreateInputGQL, err := s.appConv.CreateInputJSONToGQL(appCreateInputJSON)
	if err != nil {
		return errors.Wrapf(err, "while converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	}

	log.C(ctx).Infof("Validating GraphQL ApplicationRegistrationInput from Application Template with name %q", appTemplate.Name)
	if err := inputvalidation.Validate(appCreateInputGQL); err != nil {
		return errors.Wrapf(err, "while validating application input from Application Template with name %q", appTemplate.Name)
	}

	appCreateInputModel, err := s.appConv.CreateInputFromGraphQL(ctx, appCreateInputGQL)
	if err != nil {
		return errors.Wrap(err, "while converting ApplicationFromTemplate input")
	}

	if appCreateInputModel.Labels == nil {
		appCreateInputModel.Labels = make(map[string]interface{})
	}
	appCreateInputModel.Labels["managed"] = "false"
	appCreateInputModel.Labels[s.consumerSubaccountLabelKey] = subscribedSubaccountID
	appCreateInputModel.LocalTenantID = &consumerTenantID

	log.C(ctx).Infof("Creating an Application with name %q from Application Template with name %q", subscribedAppName, appTemplate.Name)
	_, err = s.appSvc.CreateFromTemplate(ctx, appCreateInputModel, &appTemplate.ID)
	if err != nil {
		return errors.Wrapf(err, "while creating an Application with name %s from Application Template with name %s", subscribedAppName, appTemplate.Name)
	}

	return nil
}

func processAdditionalValues(subscriptionPayload *string, appTemplate *model.ApplicationTemplate, values []*model.ApplicationTemplateValueInput, ctx context.Context, subscribedAppName string) []*model.ApplicationTemplateValueInput {
	if subscriptionPayload != nil && len(*subscriptionPayload) > 0 {
		for _, placeholder := range appTemplate.Placeholders {
			if len(*placeholder.JSONPath) > 0 {
				value := gjson.Get(*subscriptionPayload, *placeholder.JSONPath)
				if value.Exists() {
					values = append(values, &model.ApplicationTemplateValueInput{Placeholder: placeholder.Name, Value: value.String()})
				} else {
					log.C(ctx).Errorf("while parsing the callback payload with the Application template %s the value for placeholder %s with jsonPath %s, do not exists in payload.", appTemplate.Name, placeholder.Name, *placeholder.JSONPath)
				}
			} else {
				log.C(ctx).Errorf("while parsing the callback payload with the Application template %s the placeholder %s, do not have JSONPath.", appTemplate.Name, placeholder.Name)
			}
		}
	} else {
		additionalValues := []*model.ApplicationTemplateValueInput{
			{Placeholder: "name", Value: subscribedAppName},
			{Placeholder: "display-name", Value: subscribedAppName},
		}
		values = append(additionalValues, values...)
	}
	return values
}

func (s *service) deleteApplicationsByAppTemplateID(ctx context.Context, appTemplateID string) error {
	applications, err := s.appSvc.ListAll(ctx)
	if err != nil {
		return errors.Wrapf(err, "while listing applications")
	}

	for _, app := range applications {
		if str.PtrStrToStr(app.ApplicationTemplateID) == appTemplateID {
			if err := s.appSvc.Delete(ctx, app.ID); err != nil {
				return errors.Wrapf(err, "while trying to delete Application with ID: %q", app.ID)
			}
		}
	}

	return nil
}

func (s *service) buildLabelFilters(subscriptionProviderID, region string) []*labelfilter.LabelFilter {
	return []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}
}
