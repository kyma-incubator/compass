package subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/subscription/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

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

	regionalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}

	emptyLabelSvcFn = func() *automock.LabelService { return &automock.LabelService{} }
	emptyUIDSvcFn   = func() *automock.UidService { return &automock.UidService{} }

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
		RuntimeServiceFn          func() *automock.RuntimeService
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UidService
		TenantSvcFn               func() *automock.TenantService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
	}{
		{
			Name:   "Succeeds",
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
			Name:   "Returns error when could not get tenant for runtime",
			Region: tenantRegion,
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
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &createLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
		},
		{
			Name: "Returns error when could not parse label value",
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
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		},
		{
			Name: "Succeeds and updates label",
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
		{
			Name: "Succeeds and updates label on second try",
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
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &updateLabelInput).Return(testError).Once()
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
			runtimeSvc := testCase.RuntimeServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			service := NewService(runtimeSvc, tenantSvc, labelSvc, uuidSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			_, err := service.SubscribeTenant(context.TODO(), subscriptionProviderID, subaccountTenantExtID, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, runtimeSvc, labelSvc, uuidSvc, tenantSvc)
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
		UIDServiceFn              func() *automock.UidService
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
				labelSvc.On("CreateLabel", context.TODO(), tenantID, fixUUID(), &createLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
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
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
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
		{
			Name: "Succeeds and updates label on second try",
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
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &removeLabelInput).Return(testError).Once()
				labelSvc.On("UpdateLabel", context.TODO(), tenantID, testLabel.ID, &removeLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
		}}

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

			service := NewService(runtimeSvc, tenantSvc, labelSvc, uidSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			_, err := service.UnsubscribeTenant(context.TODO(), subscriptionProviderID, subaccountTenantExtID, testCase.Region)

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
