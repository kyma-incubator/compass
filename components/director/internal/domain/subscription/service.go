package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/avast/retry-go"
	labelPkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Config is configuration for the tenant subscription flow
type Config struct {
	ProviderLabelKey              string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountIDsLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY,default=consumer_subaccount_ids"`
}

// RuntimeService is responsible for Runtime operations
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	ListByFilters(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	GetByFiltersGlobal(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error)
}

// TenantService provides functionality for retrieving, and creating tenants.
//go:generate mockery --name=TenantService --output=automock --outpkg=automock --case=underscore --unroll-variadic=False --disable-version-string
type TenantService interface {
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

// LabelService is responsible updating already existing labels, and their label definitions.
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
}

//go:generate mockery --exported --name=uidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uidService interface {
	Generate() string
}

// ApplicationTemplateService is responsible for Application Template operations
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Exists(ctx context.Context, id string) (bool, error)
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

// ApplicationConverter is converting graphql and model Applications
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error)
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
}

// ApplicationService is responsible for Application operations
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error)
	ListAll(ctx context.Context) ([]*model.Application, error)
	Delete(ctx context.Context, id string) error
}

const (
	retryAttempts          = 2
	retryDelayMilliseconds = 100
)

type service struct {
	runtimeSvc                    RuntimeService
	tenantSvc                     TenantService
	labelSvc                      LabelService
	appTemplateSvc                ApplicationTemplateService
	appConv                       ApplicationConverter
	appSvc                        ApplicationService
	uidSvc                        uidService
	subscriptionProviderLabelKey  string
	consumerSubaccountIDsLabelKey string
}

// NewService returns a new object responsible for service-layer Subscription operations.
func NewService(runtimeSvc RuntimeService, tenantSvc TenantService, labelSvc LabelService, appTemplateSvc ApplicationTemplateService, appConv ApplicationConverter, appSvc ApplicationService, uidService uidService,
	subscriptionProviderLabelKey string, consumerSubaccountIDsLabelKey string) *service {
	return &service{
		runtimeSvc:                    runtimeSvc,
		tenantSvc:                     tenantSvc,
		labelSvc:                      labelSvc,
		appTemplateSvc:                appTemplateSvc,
		appConv:                       appConv,
		appSvc:                        appSvc,
		uidSvc:                        uidService,
		subscriptionProviderLabelKey:  subscriptionProviderLabelKey,
		consumerSubaccountIDsLabelKey: consumerSubaccountIDsLabelKey,
	}
}

// SubscribeTenantToRuntime subscribes a tenant to runtimes by labeling the runtime
func (s *service) SubscribeTenantToRuntime(ctx context.Context, providerID, subaccountTenantID, providerSubaccountID, region string) (bool, error) {
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrap(err, "while getting provider subaccount internal ID")
	}
	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	runtimes, err := s.runtimeSvc.ListByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	for _, provider := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, provider.ID)
		if err != nil {
			return false, err
		}

		label, err := s.labelSvc.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			ObjectID:   provider.ID,
			ObjectType: model.RuntimeLabelableObject,
		})

		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", provider.ID, s.consumerSubaccountIDsLabelKey))
			}
			if err := s.createLabel(ctx, tnt, provider, subaccountTenantID); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to create label with key: %s", s.consumerSubaccountIDsLabelKey))
			}
		} else {
			labelOldValue, err := labelPkg.ValueToStringsSlice(label.Value)
			if err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
			labelNewValue := append(labelOldValue, subaccountTenantID)

			if err := s.updateLabelWithRetry(ctx, tnt, provider, label, labelNewValue); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", provider.ID))
			}
		}
	}
	return true, nil
}

// UnsubscribeTenantFromRuntime unsubscribes a tenant from runtimes by removing labels from runtime
func (s *service) UnsubscribeTenantFromRuntime(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrap(err, "while getting provider subaccount internal ID")
	}
	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	runtimes, err := s.runtimeSvc.ListByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", tenant.RegionLabelKey, region, s.subscriptionProviderLabelKey, providerID))
	}

	for _, runtime := range runtimes {
		tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Runtime, runtime.ID)
		if err != nil {
			return false, err
		}

		label, err := s.labelSvc.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			ObjectID:   runtime.ID,
			ObjectType: model.RuntimeLabelableObject,
		})

		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", runtime.ID, s.consumerSubaccountIDsLabelKey))
			}
			return true, nil
		} else {
			labelOldValue, err := labelPkg.ValueToStringsSlice(label.Value)
			if err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
			labelNewValue := removeElement(labelOldValue, subaccountTenantID)

			if err := s.updateLabelWithRetry(ctx, tnt, runtime, label, labelNewValue); err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", runtime.ID))
			}
		}
	}

	return true, nil
}

// SubscribeTenantToApplication fetches model.ApplicationTemplate by region and provider and registers an Application from that template
func (s *service) SubscribeTenantToApplication(ctx context.Context, providerID, subscribedSubaccountID, providerSubaccountID, region, subscribedAppName string) (bool, error) {
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID: %q", providerSubaccountID)
	}
	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrapf(err, "while getting application template with filter labels %q and %q", providerID, region)
	}

	if err := s.createApplicationFromTemplate(ctx, appTemplate, subscribedSubaccountID, subscribedAppName); err != nil {
		return false, err
	}

	return true, nil
}

// UnsubscribeTenantFromApplication fetches model.ApplicationTemplate by region and provider, lists all applications for
// the providerSubaccountID tenant and deletes them synchronously
func (s *service) UnsubscribeTenantFromApplication(ctx context.Context, providerID, providerSubaccountID, region string) (bool, error) {
	providerInternalTenant, err := s.tenantSvc.GetInternalTenant(ctx, providerSubaccountID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting provider subaccount internal ID: %q", providerSubaccountID)
	}

	ctx = tenant.SaveToContext(ctx, providerInternalTenant, providerSubaccountID)

	filters := s.buildLabelFilters(providerID, region)
	appTemplate, err := s.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}

		return false, errors.Wrapf(err, "while getting application template with filter labels %q and %q", providerID, region)
	}

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
		return "", errors.Errorf("both a runtime (%q) and application template (%q) exist with filter labels provider (%q) and region (%q)", runtime, appTemplate, providerID, region)
	}

	return "", errors.Errorf("could not determine flow")
}

func (s *service) createApplicationFromTemplate(ctx context.Context, appTemplate *model.ApplicationTemplate, subscribedSubaccountID, subscribedAppName string) error {
	values := []*model.ApplicationTemplateValueInput{
		{Placeholder: "name", Value: subscribedAppName},
		{Placeholder: "display-name", Value: subscribedAppName},
	}
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
	appCreateInputModel.Labels[scenarioassignment.SubaccountIDKey] = subscribedSubaccountID

	log.C(ctx).Infof("Creating an Application with name %q from Application Template with name %q", subscribedAppName, appTemplate.Name)
	_, err = s.appSvc.CreateFromTemplate(ctx, appCreateInputModel, &appTemplate.ID)
	if err != nil {
		return errors.Wrapf(err, "while creating an Application with name %s from Application Template with name %s", subscribedAppName, appTemplate.Name)
	}

	return nil
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

func (s *service) buildLabelFilters(subscriptionLabelValue, region string) []*labelfilter.LabelFilter {
	return []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(s.subscriptionProviderLabelKey, fmt.Sprintf("\"%s\"", subscriptionLabelValue)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}
}

func (s *service) createLabel(ctx context.Context, tenant string, runtime *model.Runtime, subaccountTenantID string) error {
	return s.labelSvc.CreateLabel(ctx, tenant, s.uidSvc.Generate(), &model.LabelInput{
		Key:        s.consumerSubaccountIDsLabelKey,
		Value:      []string{subaccountTenantID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtime.ID,
	})
}

func (s *service) updateLabelWithRetry(ctx context.Context, tenant string, runtime *model.Runtime, label *model.Label, labelNewValue []string) error {
	return retry.Do(func() error {
		err := s.labelSvc.UpdateLabel(ctx, tenant, label.ID, &model.LabelInput{
			Key:        s.consumerSubaccountIDsLabelKey,
			Value:      labelNewValue,
			ObjectType: model.RuntimeLabelableObject,
			ObjectID:   runtime.ID,
			Version:    label.Version,
		})
		if err != nil {
			return errors.Wrap(err, "while updating label")
		}
		return nil
	}, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))
}

func removeElement(slice []string, elem string) []string {
	result := make([]string, 0)
	for _, e := range slice {
		if e != elem {
			result = append(result, e)
		}
	}
	return result
}
