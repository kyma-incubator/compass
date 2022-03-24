package tenantfetchersvc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	tenantExtID  = "tenant-external-id"
	tenantRegion = "myregion"

	regionalTenantSubdomain = "myregionaltenant"
	subaccountTenantExtID   = "subaccount-tenant-external-id"
	subscriptionProviderID  = "123"

	tenantProviderTenantIDProperty           = "tenantId"
	tenantProviderCustomerIDProperty         = "customerId"
	tenantProviderSubdomainProperty          = "subdomain"
	tenantProviderSubaccountTenantIDProperty = "subaccountTenantId"
	subscriptionProviderIDProperty           = "subscriptionProviderId"

	compassURL = "https://github.com/kyma-incubator/compass"
)

var (
	testError = errors.New("test error")
)

func TestSubscribeRegionalTenant(t *testing.T) {
	// GIVEN
	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionProviderID: subscriptionProviderID,
	}

	// Subscribe flow
	testCases := []struct {
		Name                      string
		TenantProvisionerFn       func() *automock.TenantProvisioner
		DirectorClient            func() *automock.DirectorGraphQLClient
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name: "Succeeds",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant).Return(nil).Once()
				return provisioner
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				directorClient := &automock.DirectorGraphQLClient{}
				directorClient.On("SubscribeTenantToRuntime", context.TODO(), regionalTenant.SubscriptionProviderID, regionalTenant.SubaccountTenantID, regionalTenant.Region).Return(nil).Once()
				return directorClient
			},
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Returns error when tenant creation fails",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant).Return(testError).Once()
				return provisioner
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				return &automock.DirectorGraphQLClient{}
			},
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when cannot subscribe tenant",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant).Return(nil).Once()
				return provisioner
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				directorClient := &automock.DirectorGraphQLClient{}
				directorClient.On("SubscribeTenantToRuntime", context.TODO(), regionalTenant.SubscriptionProviderID, regionalTenant.SubaccountTenantID, regionalTenant.Region).Return(testError).Once()
				return directorClient
			},
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			directorClient := testCase.DirectorClient()
			defer mock.AssertExpectationsForObjects(t, provisioner, directorClient)

			subscriber := tenantfetchersvc.NewSubscriber(directorClient, provisioner)

			// WHEN
			err := subscriber.Subscribe(context.TODO(), &testCase.TenantSubscriptionRequest)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnSubscribeRegionalTenant(t *testing.T) {
	// GIVEN

	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionProviderID: subscriptionProviderID,
	}

	testCases := []struct {
		Name                      string
		TenantProvisionerFn       func() *automock.TenantProvisioner
		DirectorClient            func() *automock.DirectorGraphQLClient
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name: "Succeeds",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				return &automock.TenantProvisioner{}
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				directorClient := &automock.DirectorGraphQLClient{}
				directorClient.On("UnsubscribeTenantFromRuntime", context.TODO(), regionalTenant.SubscriptionProviderID, regionalTenant.SubaccountTenantID, regionalTenant.Region).Return(nil).Once()
				return directorClient
			},
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Returns error when cannot unsubscribe tenant",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				return &automock.TenantProvisioner{}
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				directorClient := &automock.DirectorGraphQLClient{}
				directorClient.On("UnsubscribeTenantFromRuntime", context.TODO(), regionalTenant.SubscriptionProviderID, regionalTenant.SubaccountTenantID, regionalTenant.Region).Return(testError).Once()
				return directorClient
			},
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			directorClient := testCase.DirectorClient()
			defer mock.AssertExpectationsForObjects(t, provisioner, directorClient)

			subscriber := tenantfetchersvc.NewSubscriber(directorClient, provisioner)
			// WHEN
			err := subscriber.Unsubscribe(context.TODO(), &testCase.TenantSubscriptionRequest)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
