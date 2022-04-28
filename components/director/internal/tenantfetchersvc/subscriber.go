package tenantfetchersvc

import (
	"context"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore
type TenantProvisioner interface {
	ProvisionTenants(context.Context, *TenantSubscriptionRequest) error
}

type subscriptionFunc func(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error

type subscriber struct {
	gqlClient   DirectorGraphQLClient
	provisioner TenantProvisioner
}

// NewSubscriber creates new subscriber
func NewSubscriber(directorClient DirectorGraphQLClient, provisioner TenantProvisioner) *subscriber {
	return &subscriber{
		gqlClient:   directorClient,
		provisioner: provisioner,
	}
}

// Subscribe subscribes tenant to runtime. If the tenant does not exist it will be created
func (s *subscriber) Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error {
	if err := s.provisioner.ProvisionTenants(ctx, tenantSubscriptionRequest); err != nil {
		return err
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
