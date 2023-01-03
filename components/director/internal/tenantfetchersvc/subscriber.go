package tenantfetchersvc

import (
	"context"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore --disable-version-string
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

// Subscribe subscribes tenant to appTemplate/runtime. If the tenant does not exist it will be created
func (s *subscriber) Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error {
	if err := s.provisioner.ProvisionTenants(ctx, tenantSubscriptionRequest); err != nil {
		return err
	}

	return s.applySubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, tenantSubscriptionRequest.ProviderSubaccountID, tenantSubscriptionRequest.ConsumerTenantID, tenantSubscriptionRequest.Region, tenantSubscriptionRequest.SubscriptionProviderAppName, tenantSubscriptionRequest.SubscriptionPayload, true)
}

// Unsubscribe unsubscribes tenant from appTemplate/runtime.
func (s *subscriber) Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error {
	return s.applySubscriptionChange(ctx, tenantSubscriptionRequest.SubscriptionProviderID, tenantSubscriptionRequest.SubaccountTenantID, tenantSubscriptionRequest.ProviderSubaccountID, tenantSubscriptionRequest.ConsumerTenantID, tenantSubscriptionRequest.Region, tenantSubscriptionRequest.SubscriptionProviderAppName, tenantSubscriptionRequest.SubscriptionPayload, false)
}

func (s *subscriber) applySubscriptionChange(ctx context.Context, subscriptionProviderID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionProviderAppName, subscriptionPayload string, subscribe bool) error {
	var err error

	if subscribe {
		err = s.gqlClient.SubscribeTenant(ctx, subscriptionProviderID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionProviderAppName, subscriptionPayload)
	} else {
		err = s.gqlClient.UnsubscribeTenant(ctx, subscriptionProviderID, subaccountTenantID, providerSubaccountID, consumerTenantID, region)
	}
	return err
}
