package tenantfetchersvc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	tenantExtID     = "tenant-external-id"
	tenantSubdomain = "mytenant"
	tenantRegion    = "myregion"

	regionalTenantSubdomain = "myregionaltenant"
	subaccountTenantExtID   = "subaccount-tenant-external-id"
	subscriptionProviderID  = "123"

	parentTenantExtID = "parent-tenant-external-id"

	tenantProviderTenantIDProperty           = "tenantId"
	tenantProviderCustomerIDProperty         = "customerId"
	tenantProviderSubdomainProperty          = "subdomain"
	tenantProviderSubaccountTenantIDProperty = "subaccountTenantId"
	subscriptionProviderIDProperty           = "subscriptionProviderId"

	compassURL = "https://github.com/kyma-incubator/compass"
)

var (
	testError                     = errors.New("test error")
	notFoundErr                   = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	subscriptionConsumerLabelKey  = "SubscriptionProviderLabelKey"
	consumerSubaccountIDsLabelKey = "ConsumerSubaccountIDsLabelKey"
	ctxWithTenant                 = tenant.SaveToContext(context.TODO(), testRuntime.Tenant, "")

	globalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", "")),
	}
	regionalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}
	testRuntime = model.Runtime{
		ID:                "321",
		Name:              "test",
		Description:       nil,
		Tenant:            "test-tenant",
		Status:            nil,
		CreationTimestamp: time.Time{},
	}
	invalidTestLabel = model.Label{
		ID:         "456",
		Tenant:     "test-tenant",
		Key:        consumerSubaccountIDsLabelKey,
		Value:      "",
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
	}
	testLabel = model.Label{
		ID:         "456",
		Tenant:     "test-tenant",
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []interface{}{"789"},
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
	}
	updateLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{"789", subaccountTenantExtID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	createLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{subaccountTenantExtID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	removeLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{"789"},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
	emptyLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
)

func TestSubscribeGlobalTenant(t *testing.T) {
	//GIVEN

	accountProvisioningRequest := tenantfetchersvc.TenantSubscriptionRequest{
		AccountTenantID:        tenantExtID,
		CustomerTenantID:       parentTenantExtID,
		Subdomain:              tenantSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}

	testCases := []struct {
		Name                      string
		Region                    string
		TenantProvisionerFn       func() *automock.TenantProvisioner
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelUpsertServiceFn      func() *automock.LabelUpsertService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name:   "Succeeds",
			Region: "",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &accountProvisioningRequest, "").Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), globalFilters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: accountProvisioningRequest,
		},
		{
			Name:   "Returns error when tenant creation fails",
			Region: "",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &accountProvisioningRequest, "").Return(testError).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: accountProvisioningRequest,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name:   "Returns error when listing runtime fails",
			Region: "",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &accountProvisioningRequest, "").Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), globalFilters).Return(nil, testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: accountProvisioningRequest,
			ExpectedErrorOutput:       testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			runtimeSvc := testCase.RuntimeServiceFn()
			labelUpsertSvc := testCase.LabelUpsertServiceFn()
			defer mock.AssertExpectationsForObjects(t, provisioner, runtimeSvc)

			subscriber := tenantfetchersvc.NewSubscriber(provisioner, runtimeSvc, labelUpsertSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Subscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

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

func TestSubscribeRegionalTenant(t *testing.T) {
	//GIVEN

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
		Region                    string
		TenantProvisionerFn       func() *automock.TenantProvisioner
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelUpsertServiceFn      func() *automock.LabelUpsertService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name:   "Returns error when tenant creation fails",
			Region: tenantRegion,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(testError).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Succeeds when can't find runtimes",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Returns error when could not list runtimes",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not get label for runtime",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(nil, testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not parse label value",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&invalidTestLabel, nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       "Failed to parse label values for label ",
		},
		{
			Name: "Returns error when could not set label for runtime",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &updateLabelInput).Return(testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &updateLabelInput).Return(testError).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Succeeds and creates label",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(nil, notFoundErr).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &createLabelInput).Return(nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &createLabelInput).Return(nil).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Succeeds and updates label",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &updateLabelInput).Return(nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &updateLabelInput).Return(nil).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			runtimeSvc := testCase.RuntimeServiceFn()
			labelUpsertSvc := testCase.LabelUpsertServiceFn()

			defer mock.AssertExpectationsForObjects(t, provisioner, runtimeSvc)

			subscriber := tenantfetchersvc.NewSubscriber(provisioner, runtimeSvc, labelUpsertSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Subscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

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
	//GIVEN

	regionalTenant := tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionProviderID: subscriptionProviderID,
	}

	testCases := []struct {
		Name                      string
		Region                    string
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelUpsertServiceFn      func() *automock.LabelUpsertService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name:   "Succeeds when no runtime is found",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Succeeds when can't find runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Returns error when could not list runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not get label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(nil, testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not parse label value",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&invalidTestLabel, nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				return &automock.LabelUpsertService{}
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       "Failed to parse label values for label ",
		},
		{
			Name: "Returns error when could not set label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &removeLabelInput).Return(testError).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &removeLabelInput).Return(testError).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Succeeds and creates empty label",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(nil, notFoundErr).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &emptyLabelInput).Return(nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &emptyLabelInput).Return(nil).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Succeeds and updates label",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				provisioner.On("GetLabel", ctxWithTenant, testRuntime.ID, consumerSubaccountIDsLabelKey).Return(&testLabel, nil).Once()
				//provisioner.On("SetLabel", ctxWithTenant, &removeLabelInput).Return(nil).Once()
				return provisioner
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				labelSvc := &automock.LabelUpsertService{}
				labelSvc.On("UpsertLabel", ctxWithTenant, testRuntime.Tenant, &removeLabelInput).Return(nil).Once()
				return labelSvc
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := testCase.RuntimeServiceFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			defer mock.AssertExpectationsForObjects(t, runtimeSvc)

			subscriber := tenantfetchersvc.NewSubscriber(&automock.TenantProvisioner{}, runtimeSvc, labelSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Unsubscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

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
