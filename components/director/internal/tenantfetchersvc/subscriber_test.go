package tenantfetchersvc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	tenantID        = "tenantID"
	runtimeID       = "runtimeID"
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

	globalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", "")),
	}
	regionalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}

	emptyLabelSvcFn = func() *automock.LabelService { return &automock.LabelService{} }
	emptyUIDSvcFn   = func() *automock.UIDService { return &automock.UIDService{} }

	testRuntime = model.Runtime{
		ID:                runtimeID,
		Name:              "test",
		Description:       nil,
		Status:            nil,
		CreationTimestamp: time.Time{},
	}

	invalidTestLabel = model.Label{
		ID:         "456",
		Key:        consumerSubaccountIDsLabelKey,
		Value:      "",
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testLabel = model.Label{
		ID:         "456",
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []interface{}{"789"},
		ObjectID:   testRuntime.ID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    1,
	}

	updateLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{"789", subaccountTenantExtID},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
		Version:    testLabel.Version,
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
		Version:    testLabel.Version,
	}

	emptyLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		Value:      []string{""},
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}

	getLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
)

func TestSubscribeGlobalTenant(t *testing.T) {
	// GIVEN

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
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UIDService
		TenantSvcFn               func() *automock.TenantService
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			RuntimeServiceFn:          func() *automock.RuntimeService { return &automock.RuntimeService{} },
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: accountProvisioningRequest,
			ExpectedErrorOutput:       testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			runtimeSvc := testCase.RuntimeServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			subscriber := tenantfetchersvc.NewSubscriber(provisioner, runtimeSvc, labelSvc, uuidSvc, tenantSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Subscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, provisioner, runtimeSvc, labelSvc, uuidSvc, tenantSvc)
		})
	}
}

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
		Region                    string
		TenantProvisionerFn       func() *automock.TenantProvisioner
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UIDService
		TenantSvcFn               func() *automock.TenantService
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			RuntimeServiceFn:          func() *automock.RuntimeService { return &automock.RuntimeService{} },
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name:   "Returns error when could not get tenant for runtime",
			Region: tenantRegion,
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return("", testError).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, testError).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not create label for runtime",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &createLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UIDService {
				uidService := &automock.UIDService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
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
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&invalidTestLabel, nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       "Failed to parse label values for label ",
		},
		{
			Name: "Returns error when could not update label for runtime",
			TenantProvisionerFn: func() *automock.TenantProvisioner {
				provisioner := &automock.TenantProvisioner{}
				provisioner.On("ProvisionTenants", context.TODO(), &regionalTenant, tenantRegion).Return(nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &updateLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
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
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &createLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UIDService {
				uidService := &automock.UIDService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
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
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &updateLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			provisioner := testCase.TenantProvisionerFn()
			runtimeSvc := testCase.RuntimeServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			subscriber := tenantfetchersvc.NewSubscriber(provisioner, runtimeSvc, labelSvc, uuidSvc, tenantSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Subscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, provisioner, runtimeSvc, labelSvc, uuidSvc, tenantSvc)
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
		Region                    string
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UIDService
		TenantSvcFn               func() *automock.TenantService
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Succeeds when can't find runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
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
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not get tenant for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not get label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, testError).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not create label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &createLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UIDService {
				uidService := &automock.UIDService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
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
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&invalidTestLabel, nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       "Failed to parse label values for label ",
		},
		{
			Name: "Returns error when could not update label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &removeLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Succeeds and creates empty label",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &emptyLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UIDService {
				uidService := &automock.UIDService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			Region: tenantRegion,
			TenantSubscriptionRequest: tenantfetchersvc.TenantSubscriptionRequest{
				SubaccountTenantID:     "",
				AccountTenantID:        tenantExtID,
				Subdomain:              regionalTenantSubdomain,
				Region:                 tenantRegion,
				SubscriptionProviderID: subscriptionProviderID,
			},
		},
		{
			Name: "Succeeds and updates label",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFiltersGlobal", context.TODO(), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", context.TODO(), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", context.TODO(), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &removeLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := testCase.RuntimeServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			defer mock.AssertExpectationsForObjects(t, runtimeSvc)

			subscriber := tenantfetchersvc.NewSubscriber(&automock.TenantProvisioner{}, runtimeSvc, labelSvc, uidSvc, tenantSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			err := subscriber.Unsubscribe(context.TODO(), &testCase.TenantSubscriptionRequest, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, runtimeSvc, labelSvc, uidSvc, tenantSvc)
		})
	}
}

func fixUUID() string {
	return "647af599-7f2d-485c-a63b-615b5ff6daf1"
}
