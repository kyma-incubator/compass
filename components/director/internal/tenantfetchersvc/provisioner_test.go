package tenantfetchersvc_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/assert"
)

const (
	autogeneratedProviderName = "autogenerated"
	testProviderName          = "test-provider"
)

var (
	customerTenant = model.BusinessTenantMappingInput{
		Name:           parentTenantExtID,
		ExternalTenant: parentTenantExtID,
		Parent:         "",
		Type:           tenantEntity.TypeToStr(tenantEntity.Customer),
		Provider:       autogeneratedProviderName,
	}
	accountTenant = model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parent:         parentTenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      tenantSubdomain,
		Region:         tenantRegion,
	}
	parentAccountTenant = model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Parent:         parentTenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      "",
		Region:         "",
	}
	subaccountTenant = model.BusinessTenantMappingInput{
		Name:           subaccountTenantExtID,
		ExternalTenant: subaccountTenantExtID,
		Parent:         tenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Subaccount),
		Provider:       testProviderName,
		Subdomain:      tenantSubdomain,
		Region:         tenantRegion,
	}
	accountTenantWithoutParent = model.BusinessTenantMappingInput{
		Name:           tenantExtID,
		ExternalTenant: tenantExtID,
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       testProviderName,
		Subdomain:      tenantSubdomain,
		Region:         tenantRegion,
	}

	requestWithAccountTenant = &tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:  tenantExtID,
		CustomerTenantID: parentTenantExtID,
		Subdomain:        tenantSubdomain,
		Region:           tenantRegion,
	}

	requestWithAccountTenantWithoutParent = &tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID: tenantExtID,
		Subdomain:       tenantSubdomain,
		Region:          tenantRegion,
	}

	requestWithSubaccountTenant = &tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID: subaccountTenantExtID,
		AccountTenantID:    tenantExtID,
		CustomerTenantID:   parentTenantExtID,
		Subdomain:          tenantSubdomain,
		Region:             tenantRegion,
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
				expectedTenants := []model.BusinessTenantMappingInput{customerTenant, parentAccountTenant, subaccountTenant}
				expectedTenantsConverted := convertTenantsToGQLInput(expectedTenants)
				tenantSvc.On("MultipleInputToGraphQLInput", expectedTenants).Return(expectedTenantsConverted).Once()
				return tenantSvc
			},
			DirectorClient: func() *automock.DirectorGraphQLClient {
				expectedTenants := []model.BusinessTenantMappingInput{customerTenant, parentAccountTenant, subaccountTenant}
				expectedTenantsConverted := convertTenantsToGQLInput(expectedTenants)
				tenantSvc := &automock.DirectorGraphQLClient{}
				tenantSvc.On("WriteTenants", ctx, expectedTenantsConverted).Return(nil).Once()
				return tenantSvc
			},
			Request: requestWithSubaccountTenant,
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

			provisioner := tenantfetchersvc.NewTenantProvisioner(directorClient, tenantConverter, testProviderName)

			// WHEN
			err := provisioner.ProvisionTenants(ctx, testCase.Request, "asd")

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, tenantConverter, directorClient)
		})
	}
}

func convertTenantsToGQLInput(tenants []model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput {
	return tenant.NewConverter().MultipleInputToGraphQLInput(tenants)
}
