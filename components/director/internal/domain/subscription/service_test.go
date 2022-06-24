package subscription_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription/automock"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	tenantID            = "tenantID"
	runtimeID           = "runtimeID"
	tenantExtID         = "tenant-external-id"
	tenantRegion        = "myregion"
	subscriptionAppName = "my-app"

	regionalTenantSubdomain = "myregionaltenant"
	subaccountTenantExtID   = "subaccount-tenant-external-id"
	subscriptionProviderID  = "123"
	subscribedSubaccountID  = "333-444"
	providerSubaccountID    = "123-456"
	providerInternalID      = "456-789"

	uuid = "647af599-7f2d-485c-a63b-615b5ff6daf1"
)

var (
	testError                     = errors.New("test error")
	notFoundErr                   = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	notFoundAppTemplateErr        = apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate)
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

	getLabelInput = model.LabelInput{
		Key:        consumerSubaccountIDsLabelKey,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   testRuntime.ID,
	}
)

func CtxWithTenantMatcher(expectedTenantID string) interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		tenantID, err := tenant.LoadFromContext(ctx)
		return err == nil && tenantID == expectedTenantID
	})
}

func TestSubscribeRegionalTenant(t *testing.T) {
	// Subscribe flow
	testCases := []struct {
		Name                string
		Region              string
		RuntimeServiceFn    func() *automock.RuntimeService
		LabelServiceFn      func() *automock.LabelService
		UIDServiceFn        func() *automock.UidService
		TenantSvcFn         func() *automock.TenantService
		ExpectedErrorOutput string
		IsSuccessful        bool
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: emptyLabelSvcFn,
			UIDServiceFn:   emptyUIDSvcFn,
			IsSuccessful:   true,
		},
		{
			Name: "Returns an error when can't internal provider tenant",
			RuntimeServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			Region:              tenantRegion,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name: "Returns an error when can't find runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: emptyLabelSvcFn,
			UIDServiceFn:   emptyUIDSvcFn,
			Region:         tenantRegion,
			IsSuccessful:   false,
		},
		{
			Name: "Returns an error when could not list runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			Region:              tenantRegion,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when could not get tenant for runtime",
			Region: tenantRegion,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return("", testError).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name: "Returns an error when could not get label for runtime",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(nil, testError).Once()
				return labelSvc
			},
			UIDServiceFn:        emptyUIDSvcFn,
			Region:              tenantRegion,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name: "Returns an error when could not create label for runtime",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", mock.AnythingOfType("*context.valueCtx"), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, uuid, &createLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(uuid)
				return uidService
			},
			Region:              tenantRegion,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name: "Returns an error when could not parse label value",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&invalidTestLabel, nil).Once()
				return labelSvc
			},
			UIDServiceFn:        emptyUIDSvcFn,
			Region:              tenantRegion,
			ExpectedErrorOutput: "Failed to parse label values for label ",
			IsSuccessful:        false,
		},
		{
			Name: "Returns an error when could not update label for runtime",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &updateLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn:        emptyUIDSvcFn,
			Region:              tenantRegion,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name: "Succeeds and creates label",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(nil, notFoundErr).Once()
				labelSvc.On("CreateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, uuid, &createLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidService := &automock.UidService{}
				uidService.On("Generate").Return(uuid)
				return uidService
			},
			Region:       tenantRegion,
			IsSuccessful: true,
		},
		{
			Name: "Succeeds and updates label",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &updateLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: emptyUIDSvcFn,
			Region:       tenantRegion,
			IsSuccessful: true,
		},
		{
			Name: "Succeeds and updates label on second try",
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &updateLabelInput).Return(testError).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &updateLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: emptyUIDSvcFn,
			Region:       tenantRegion,
			IsSuccessful: true,
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

			service := subscription.NewService(runtimeSvc, tenantSvc, labelSvc, nil, nil, nil, uuidSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			isSubscribeSuccessful, err := service.SubscribeTenantToRuntime(context.TODO(), subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, isSubscribeSuccessful)

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
		IsSuccessful              bool
	}{
		{
			Name:   "Succeeds when no runtime is found",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              true,
		},
		{
			Name: "Returns an error when can't internal provider tenant",
			RuntimeServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name: "Returns an error when can't find runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              false,
		},
		{
			Name: "Returns an error when could not list runtimes",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name: "Returns an error when could not get tenant for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name: "Returns an error when could not get label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(nil, testError).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name: "Succeeds if label for runtime is not found",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(nil, apperrors.NewNotFoundError(resource.Label, getLabelInput.ObjectID)).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              true,
		},
		{
			Name: "Returns an error when could not parse label value",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&invalidTestLabel, nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       "Failed to parse label values for label ",
			IsSuccessful:              false,
		},
		{
			Name: "Returns an error when could not update label for runtime",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &removeLabelInput).Return(testError).Twice()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name: "Succeeds and updates label",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &removeLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              true,
		},
		{
			Name: "Succeeds and updates label on second try",
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return([]*model.Runtime{&testRuntime}, nil).Once()
				return provisioner
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", mock.AnythingOfType("*context.valueCtx"), resource.Runtime, runtimeID).Return(tenantID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("GetLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, &getLabelInput).Return(&testLabel, nil).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &removeLabelInput).Return(testError).Once()
				labelSvc.On("UpdateLabel", mock.AnythingOfType("*context.valueCtx"), tenantID, testLabel.ID, &removeLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			Region:                    tenantRegion,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              true,
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
			defer mock.AssertExpectationsForObjects(t, runtimeSvc, labelSvc, uidSvc, tenantSvc)

			service := subscription.NewService(runtimeSvc, tenantSvc, labelSvc, nil, nil, nil, uidSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			isUnsubscribeSuccessful, err := service.UnsubscribeTenantFromRuntime(context.TODO(), subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, isUnsubscribeSuccessful)
		})
	}
}

func TestSubscribeTenantToApplication(t *testing.T) {
	appTmplName := "test-app-tmpl"
	appTmplAppName := "{{name}}"
	appTmplID := "123-456-789"

	jsonAppCreateInput := fixJSONApplicationCreateInput(appTmplAppName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(appTmplID, appTmplName, jsonAppCreateInput)
	modelAppFromTemplateInput := fixModelApplicationFromTemplateInput(appTmplName, subscriptionAppName)
	gqlAppCreateInput := fixGQLApplicationCreateInput(appTmplName)
	modelAppCreateInput := fixModelApplicationCreateInput(appTmplName)
	modelAppCreateInputWithLabels := fixModelApplicationCreateInputWithLabels(appTmplName, subscribedSubaccountID)

	testCases := []struct {
		Name                   string
		Region                 string
		SubscriptionAppName    string
		SubscriptionProviderID string
		AppTemplateServiceFn   func() *automock.ApplicationTemplateService
		LabelServiceFn         func() *automock.LabelService
		UIDServiceFn           func() *automock.UidService
		TenantSvcFn            func() *automock.TenantService
		AppConverterFn         func() *automock.ApplicationConverter
		AppSvcFn               func() *automock.ApplicationService
		ExpectedErrorOutput    string
		IsSuccessful           bool
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", CtxWithTenantMatcher(providerInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			LabelServiceFn: emptyLabelSvcFn,
			UIDServiceFn:   emptyUIDSvcFn,
			IsSuccessful:   true,
		},
		{
			Name:   "Returns an error when can't find internal provider tenant",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "GetByFilters")
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when can't find app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, notFoundErr).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: emptyLabelSvcFn,
			UIDServiceFn:   emptyUIDSvcFn,
			IsSuccessful:   false,
		},
		{
			Name:   "Returns an error when fails finding app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, testError).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when preparing application input json",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return("", testError).Once()

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating graphql input from json",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, testError).Once()
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating input from graphql",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", CtxWithTenantMatcher(providerInternalID), gqlAppCreateInput).Return(model.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.AssertNotCalled(t, "CreateFromTemplate")
				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating app from app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", CtxWithTenantMatcher(providerInternalID), gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("CreateFromTemplate", CtxWithTenantMatcher(providerInternalID), modelAppCreateInputWithLabels, &appTmplID).Return("", testError).Once()
				return appSvc
			},
			LabelServiceFn:      emptyLabelSvcFn,
			UIDServiceFn:        emptyUIDSvcFn,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateSvc := testCase.AppTemplateServiceFn()
			labelSvc := testCase.LabelServiceFn()
			appConv := testCase.AppConverterFn()
			appSvc := testCase.AppSvcFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			service := subscription.NewService(nil, tenantSvc, labelSvc, appTemplateSvc, appConv, appSvc, uuidSvc, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			isSubscribeSuccessful, err := service.SubscribeTenantToApplication(context.TODO(), subscriptionProviderID, subscribedSubaccountID, providerSubaccountID, testCase.Region, subscriptionAppName)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, isSubscribeSuccessful)

			mock.AssertExpectationsForObjects(t, appTemplateSvc, labelSvc, uuidSvc, tenantSvc, appConv, appSvc)
		})
	}
}

func TestUnsubscribeTenantFromApplication(t *testing.T) {
	appTmplAppName := "app-name"
	appTmplName := "app-tmpl-name"
	appTmplID := "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appFirstID := "b91b59f7-2563-40b2-aba9-fef726037bb4"
	appSecondID := "b91b59e6-2563-40b2-aba9-fef726037cc5"
	jsonAppCreateInput := fixJSONApplicationCreateInput(appTmplAppName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(appTmplID, appTmplName, jsonAppCreateInput)
	modelApps := []*model.Application{
		fixModelApplication(appFirstID, "first", appTmplID),
		fixModelApplication(appSecondID, "second", appTmplID),
	}
	testCases := []struct {
		Name                 string
		AppTemplateServiceFn func() *automock.ApplicationTemplateService
		AppSvcFn             func() *automock.ApplicationService
		TenantSvcFn          func() *automock.TenantService
		ExpectedErrorOutput  string
		IsSuccessful         bool
	}{
		{
			Name: "Success",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", CtxWithTenantMatcher(providerInternalID)).Return(modelApps, nil).Once()
				svc.On("Delete", CtxWithTenantMatcher(providerInternalID), appFirstID).Return(nil).Once()
				svc.On("Delete", CtxWithTenantMatcher(providerInternalID), appSecondID).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Error when fails to get internal tenant",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.AssertNotCalled(t, "GetByFilters")
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "ListAll")
				svc.AssertNotCalled(t, "Delete")
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when getting app template by filters",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, testError).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "ListAll")
				svc.AssertNotCalled(t, "Delete")
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when app template by filters is not found",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(nil, notFoundAppTemplateErr).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.AssertNotCalled(t, "ListAll")
				svc.AssertNotCalled(t, "Delete")
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful: false,
		},
		{
			Name: "Error when listing applications",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", CtxWithTenantMatcher(providerInternalID)).Return(nil, testError).Once()
				svc.AssertNotCalled(t, "Delete")
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when deleting application",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", CtxWithTenantMatcher(providerInternalID), regionalFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", CtxWithTenantMatcher(providerInternalID)).Return(modelApps, nil).Once()
				svc.On("Delete", CtxWithTenantMatcher(providerInternalID), appFirstID).Return(nil).Once()
				svc.On("Delete", CtxWithTenantMatcher(providerInternalID), appSecondID).Return(testError).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateSvc := testCase.AppTemplateServiceFn()
			appSvc := testCase.AppSvcFn()
			tenantSvc := testCase.TenantSvcFn()
			service := subscription.NewService(nil, tenantSvc, nil, appTemplateSvc, nil, appSvc, nil, subscriptionConsumerLabelKey, consumerSubaccountIDsLabelKey)

			// WHEN
			successful, err := service.UnsubscribeTenantFromApplication(context.TODO(), subscriptionProviderID, providerSubaccountID, tenantRegion)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, successful)

			mock.AssertExpectationsForObjects(t, appTemplateSvc, tenantSvc, appSvc)
		})
	}
}

func TestDetermineSubscriptionFlow(t *testing.T) {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery("subscriptionProviderId", fmt.Sprintf("\"%s\"", providerSubaccountID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", regionalTenantSubdomain)),
	}
	appTemplateID := "app-tmpl-ID"
	appTemplateName := "app-tmpl-name"
	rtmName := "rtm-name"
	modelAppTemplate := fixModelApplicationTemplate(appTemplateID, appTemplateName)
	modelRuntime := fixModelRuntime(rtmName)

	testCases := []struct {
		Name                string
		AppTemplateFn       func() *automock.ApplicationTemplateService
		RuntimeFn           func() *automock.RuntimeService
		Output              resource.Type
		ExpectedErrorOutput string
	}{
		{
			Name: "Success for runtime",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), filters).Return(nil, nil).Once()
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(modelRuntime, nil).Once()
				return runtimeSvc
			},
			Output: resource.Runtime,
		},
		{
			Name: "Success for application template",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), filters).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(nil, nil).Once()
				return runtimeSvc
			},
			Output: resource.ApplicationTemplate,
		},
		{
			Name: "Error for runtime fetch by filters",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.AssertNotCalled(t, "GetByFilters")
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(nil, testError).Once()
				return runtimeSvc
			},
			Output:              "",
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error for application template fetch by filters",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), filters).Return(nil, testError).Once()
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(nil, nil).Once()
				return runtimeSvc
			},
			Output:              "",
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when a runtime and app template exist",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), filters).Return(modelAppTemplate, nil).Once()
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(modelRuntime, nil).Once()
				return runtimeSvc
			},
			Output:              "",
			ExpectedErrorOutput: fmt.Sprintf("both a runtime (%+v) and application template (%+v) exist with filter labels provider (%q) and region (%q)", modelRuntime, modelAppTemplate, providerSubaccountID, regionalTenantSubdomain),
		},
		{
			Name: "Success when no runtime and app template exists",
			AppTemplateFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), filters).Return(nil, nil).Once()
				return appTemplateSvc
			},
			RuntimeFn: func() *automock.RuntimeService {
				runtimeSvc := &automock.RuntimeService{}
				runtimeSvc.On("GetByFiltersGlobal", context.TODO(), filters).Return(nil, nil).Once()
				return runtimeSvc
			},
			Output: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateSvc := testCase.AppTemplateFn()
			rtmService := testCase.RuntimeFn()
			defer mock.AssertExpectationsForObjects(t, appTemplateSvc, rtmService)

			service := subscription.NewService(rtmService, nil, nil, appTemplateSvc, nil, nil, nil, "subscriptionProviderId", "")

			output, err := service.DetermineSubscriptionFlow(context.TODO(), providerSubaccountID, regionalTenantSubdomain)
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.Output, output)
		})
	}
}
