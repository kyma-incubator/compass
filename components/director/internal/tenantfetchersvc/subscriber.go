package tenantfetchersvc

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantProvisioner interface {
	ProvisionTenants(context.Context, *TenantSubscriptionRequest) error
}

type subscriptionFunc func(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error

type subscriber struct {
	gqlClient                       DirectorGraphQLClient
	provisioner                     TenantProvisioner
	selfRegisterDistinguishLabelKey string
}

// NewSubscriber creates new subscriber
func NewSubscriber(directorClient DirectorGraphQLClient, provisioner TenantProvisioner, selfRegisterDistinguishLabelKey string) *subscriber {
	return &subscriber{
		gqlClient:                       directorClient,
		provisioner:                     provisioner,
		selfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
	}
}

// Subscribe subscribes tenant to runtime. If the tenant does not exist it will be created
func (s *subscriber) Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error {
	if err := s.provisioner.ProvisionTenants(ctx, tenantSubscriptionRequest); err != nil {
		return err
	}

	appTemplates, runtimes, err := s.fetchAppTemplatesAndRuntimes(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.Region)
	if err != nil {
		return err
	}

	flowType, err := s.determineFlow(appTemplates.TotalCount, runtimes.TotalCount)
	if err != nil {
		return errors.Wrap(err, "while determining subscription flow")
	}

	if flowType == resource.ApplicationTemplate {
		return s.applyApplicationsSubscriptionChange(ctx, appTemplates.Data[0], tenantSubscriptionRequest.SubaccountTenantID, tenantSubscriptionRequest.SubscriptionAppName, true)
	}

	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, tenantSubscriptionRequest.ProviderSubaccountID, tenantSubscriptionRequest.Region, true)
}

// Unsubscribe unsubscribes tenant from runtime.
func (s *subscriber) Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error {
	return s.applyRuntimesSubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, tenantSubscriptionRequest.ProviderSubaccountID, tenantSubscriptionRequest.Region, false)
}

func (s *subscriber) applyRuntimesSubscriptionChange(ctx context.Context, subscriptionProviderID, subaccountTenantID, providerSubaccountID, region string, subscribe bool) error {
	var err error

	if subscribe {
		err = s.gqlClient.SubscribeTenantToRuntime(ctx, subscriptionProviderID, subaccountTenantID, providerSubaccountID, region)
	} else {
		err = s.gqlClient.UnsubscribeTenantFromRuntime(ctx, subscriptionProviderID, subaccountTenantID, providerSubaccountID, region)
	}
	return err
}

func (s *subscriber) applyApplicationsSubscriptionChange(ctx context.Context, appTemplate *graphql.ApplicationTemplate, subaccountTenantID, subscriptionAppName string, subscribe bool) error {
	var err error
	if appTemplate == nil {
		errors.New("no application template provided")
	}

	if subscribe {
		err = s.gqlClient.RegisterApplicationFromTemplate(ctx, appTemplate.Name, subaccountTenantID, subscriptionAppName)
	} else {
		// TODO: delete app from app tmpl
	}
	return err
}

func (s *subscriber) determineFlow(appTemplateCount, runtimeCount int) (resource.Type, error) {
	if runtimeCount == 1 && appTemplateCount == 0 {
		return resource.Runtime, nil
	}

	if runtimeCount == 0 && appTemplateCount == 1 {
		return resource.ApplicationTemplate, nil
	}

	if runtimeCount == 0 && appTemplateCount == 0 {
		return "", errors.Errorf("no runtime or application template exists with such labels and cannot determine the flow")
	}

	if runtimeCount > 0 && appTemplateCount > 0 {
		return "", errors.Errorf("runtimes and application templates exist with such labels and cannot determine the flow")
	}

	return "", errors.Errorf("could not determine flow")
}

func (s *subscriber) fetchAppTemplatesAndRuntimes(ctx context.Context, subscriptionProviderID, region string) (graphql.ApplicationTemplatePage, graphql.RuntimePage, error) {
	log.C(ctx).Infof("fetching application templates with labels: { %s: %s, %s: %s }", tenant.RegionLabelKey, region, s.selfRegisterDistinguishLabelKey, subscriptionProviderID)
	applicationTemplates, err := s.gqlClient.GetApplicationTemplates(ctx, region, s.selfRegisterDistinguishLabelKey, subscriptionProviderID)
	if err != nil {
		return graphql.ApplicationTemplatePage{}, graphql.RuntimePage{}, errors.Wrapf(err, "while quering application templates")
	}

	log.C(ctx).Infof("fetching runtimes with labels: { %s: %s, %s: %s }", tenant.RegionLabelKey, region, s.selfRegisterDistinguishLabelKey, subscriptionProviderID)
	runtimes, err := s.gqlClient.GetRuntimes(ctx, region, s.selfRegisterDistinguishLabelKey, subscriptionProviderID)
	if err != nil {
		return graphql.ApplicationTemplatePage{}, graphql.RuntimePage{}, errors.Wrapf(err, "while quering runtimes")
	}

	return applicationTemplates, runtimes, nil
}
