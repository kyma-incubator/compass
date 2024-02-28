package tenantfetchersvc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/assert"
)

const (
	autogeneratedProviderName   = "autogenerated"
	testProviderName            = "test-provider"
	parentTenantExtID           = "parent-tenant-external-id"
	parentCostObjectTenantExtID = "parent-cost-object-external-id"
)

var (
	testLicenseType = "LICENSETYPE"
	customerTenant  = model.BusinessTenantMappingInput{
		Name:           parentTenantExtID,
		ExternalTenant: parentTenantExtID,
		Parents:        []string{},
		Type:           tenantEntity.TypeToStr(tenantEntity.Customer),
		Provider:       autogeneratedProviderName,
		LicenseType:    &testLicenseType,
	}
	costObjectTenant = model.BusinessTenantMappingInput{
		Name:           parentCostObjectTenantExtID,
		ExternalTenant: parentCostObjectTenantExtID,
		Parents:        []string{},
		Type:           tenantEntity.TypeToStr(tenantEntity.CostObject),
		Provider:       autogeneratedProviderName,
		LicenseType:    &testLicenseType,
	}
	parentAccountTenant = model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parents:        []string{parentTenantExtID},
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      "",
		Region:         "",
		LicenseType:    &testLicenseType,
	}
	parentCostObjectAccountTenant = model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parents:        []string{parentCostObjectTenantExtID},
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      "",
		Region:         "",
		LicenseType:    &testLicenseType,
	}
	subaccountTenant = model.BusinessTenantMappingInput{
		Name:           subaccountTenantExtID,
		ExternalTenant: subaccountTenantExtID,
		Parents:        []string{tenantExtID},
		Type:           tenantEntity.TypeToStr(tenantEntity.Subaccount),
		Provider:       testProviderName,
		Subdomain:      regionalTenantSubdomain,
		Region:         tenantRegion,
		LicenseType:    &testLicenseType,
	}

	requestWithSubaccountTenant = &tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		CustomerTenantID:       parentTenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionLcenseType: testLicenseType,
	}

	requestWithCostObject = &tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		CostObjectTenantID:     parentCostObjectTenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionLcenseType: testLicenseType,
	}
)

func TestProvisioner_CreateRegionalTenant(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testCases := []struct {
		Name                string
		DirectorClient      func() *automock.DirectorGraphQLClient
		TenantConverter     func() *automock.TenantConverter
		Request             *tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput string
	}{
		{
			Name: "Succeeds when parent account tenant already exists",
			TenantConverter: func() *automock.TenantConverter {
				tenantSvc := &automock.TenantConverter{}
				return tenantSvc
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				tenantSvc := &automock.DirectorGraphQLClient{}
				tenantSvc.On("ExistsTenantByExternalID", ctx, parentTenantExtID).Return(true, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, tenantExtID).Return(true, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, subaccountTenantExtID).Return(true, nil).Once()
				return tenantSvc
			},
			Request: requestWithSubaccountTenant,
		},
		{
			Name: "Succeeds when a cost object is present",
			TenantConverter: func() *automock.TenantConverter {
				tenantSvc := &automock.TenantConverter{}
				return tenantSvc
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				tenantSvc := &automock.DirectorGraphQLClient{}
				tenantSvc.On("ExistsTenantByExternalID", ctx, parentCostObjectTenantExtID).Return(true, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, tenantExtID).Return(true, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, subaccountTenantExtID).Return(true, nil).Once()
				return tenantSvc
			},
			Request: requestWithCostObject,
		},
		{
			Name: "Returns error when checking for tenant fails",
			TenantConverter: func() *automock.TenantConverter {
				tenantSvc := &automock.TenantConverter{}
				return tenantSvc
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				tenantSvc := &automock.DirectorGraphQLClient{}
				tenantSvc.On("ExistsTenantByExternalID", ctx, parentTenantExtID).Return(true, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, tenantExtID).Return(false, testError).Once()
				return tenantSvc
			},
			Request:             requestWithSubaccountTenant,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Returns error when tenant creation fails",
			TenantConverter: func() *automock.TenantConverter {
				tenantSvc := &automock.TenantConverter{}
				expectedTenants := []model.BusinessTenantMappingInput{customerTenant, parentAccountTenant, subaccountTenant}
				expectedTenantsConverted := convertTenantsToGQLInput(expectedTenants)
				tenantSvc.On("MultipleInputToGraphQLInput", expectedTenants).Return(expectedTenantsConverted).Once()
				return tenantSvc
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				expectedTenants := []model.BusinessTenantMappingInput{customerTenant, parentAccountTenant, subaccountTenant}
				expectedTenantsConverted := convertTenantsToGQLInput(expectedTenants)
				tenantSvc := &automock.DirectorGraphQLClient{}
				tenantSvc.On("ExistsTenantByExternalID", ctx, parentTenantExtID).Return(false, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, tenantExtID).Return(false, nil).Once()
				tenantSvc.On("ExistsTenantByExternalID", ctx, subaccountTenantExtID).Return(false, nil).Once()
				tenantSvc.On("WriteTenants", ctx, expectedTenantsConverted).Return(testError).Once()
				return tenantSvc
			},
			Request:             requestWithSubaccountTenant,
			ExpectedErrorOutput: testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantConverter := testCase.TenantConverter()
			directorClient := testCase.DirectorClient()
			defer mock.AssertExpectationsForObjects(t, tenantConverter, directorClient)

			provisioner := tenantfetchersvc.NewTenantProvisioner(directorClient, tenantConverter, testProviderName)

			// WHEN
			err := provisioner.ProvisionTenants(ctx, testCase.Request)

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

func convertTenantsToGQLInput(tenants []model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput {
	return tenant.NewConverter().MultipleInputToGraphQLInput(tenants)
}
