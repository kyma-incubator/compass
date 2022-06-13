package subscription

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/internal/subscription/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	tenantExtID  = "tenant-external-id"
	tenantRegion = "myregion"

	subscriptionAppName        = "subscription-app-name-value"
	regionalTenantSubdomain    = "myregionaltenant"
	subaccountTenantExtID      = "32468d6e-f4cc-453c-beca-28c7bf55e0dd"
	subaccountTenantInternalID = "b6fe5f45-d8ba-4aa8-a949-5f8edb00d4a3"
	subscriptionProviderID     = "id-value!t12345"
	providerSubaccountID       = "9cdbe10c-778c-432e-bf0e-e9686d04c679"
	providerInternalID         = "aa6301aa-7cf5-4335-82ae-35d078f8a2ed"
	consumerTenantID           = "ddf290d5-31c2-457e-a14d-461d3df95ac9"
	runtimeCtxID               = "cf226ea8-31f8-475d-bd28-5cba3df9c199"
	providerRuntimeID          = "96c85d13-22ee-4555-9b41-f5e364070c20"
	runtimeM2MTableName        = "tenant_runtimes"

	uuid = "647af599-7f2d-485c-a63b-615b5ff6daf1"
)

var (
	testError                      = errors.New("test error")
	notFoundErr                    = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	subscriptionProviderIDLabelKey = "subscriptionProviderId"
	consumerSubaccountLabelKey     = "consumer_subaccount_id"
	subscriptionLabelKey           = "subscription"
	subscriptionAppNameLabelKey    = "runtimeType"

	emptyLabelSvcFn   = func() *automock.LabelService { return &automock.LabelService{} }
	emptyRuntimeSvcFn = func() *automock.RuntimeService { return &automock.RuntimeService{} }
	emptyRtmCtxSvcFn  = func() *automock.RuntimeCtxService { return &automock.RuntimeCtxService{} }
	emptyUIDSvcFn     = func() *automock.UidService { return &automock.UidService{} }

	regionalFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionProviderIDLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}

	regionalTenant = tenantfetchersvc.TenantSubscriptionRequest{
		SubaccountTenantID:     subaccountTenantExtID,
		AccountTenantID:        tenantExtID,
		Subdomain:              regionalTenantSubdomain,
		Region:                 tenantRegion,
		SubscriptionProviderID: subscriptionProviderID,
	}

	providerRuntimes = []*model.Runtime{
		{
			ID:   providerRuntimeID,
			Name: "provider-runtime-1",
		},
	}

	providerTenant = &model.BusinessTenantMapping{
		ID:             providerInternalID,
		ExternalTenant: providerSubaccountID,
	}

	consumerTenant = &model.BusinessTenantMapping{
		ID:             subaccountTenantInternalID,
		ExternalTenant: subaccountTenantExtID,
	}

	providerAppNameLabelInput = &model.LabelInput{
		Key:        subscriptionAppNameLabelKey,
		Value:      subscriptionAppName,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   providerRuntimeID,
	}

	runtimeCtxInput = model.RuntimeContextInput{
		Key:       subscriptionLabelKey,
		Value:     consumerTenantID,
		RuntimeID: providerRuntimeID,
	}

	consumerSubaccountLabelInput = &model.LabelInput{
		Key:        consumerSubaccountLabelKey,
		Value:      subaccountTenantExtID,
		ObjectID:   runtimeCtxID,
		ObjectType: model.RuntimeContextLabelableObject,
	}
)

func TestSubscribeRegionalTenant(t *testing.T) {
	// GIVEN
	db, DBMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	providerCtx := tenant.SaveToContext(ctx, providerInternalID, providerSubaccountID)
	consumerCtx := tenant.SaveToContext(providerCtx, subaccountTenantInternalID, subaccountTenantExtID)

	// Subscribe flow
	testCases := []struct {
		Name                      string
		Region                    string
		RuntimeServiceFn          func() *automock.RuntimeService
		RuntimeCtxServiceFn       func() *automock.RuntimeCtxService
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UidService
		TenantSvcFn               func() *automock.TenantService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		TenantAccessMock          func(DBMock testdb.DBMock)
		ExpectedErrorOutput       string
		IsSuccessful              bool
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, consumerSubaccountLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Once()
				return uidSvc
			},
			TenantSubscriptionRequest: regionalTenant,
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			IsSuccessful: true,
		},
		{
			Name:                "Returns an error when getting internal provider tenant",
			Region:              tenantRegion,
			RuntimeServiceFn:    emptyRuntimeSvcFn,
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(nil, testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when can't find runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when could not list runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when could not get lowest owner of the runtime",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when could not upsert provider app name label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				return tenantSvc
			},

			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when getting consumer tenant",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(nil, testError).Once()
				return tenantSvc
			},

			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when creating runtime context",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return("", testError).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when creating tenant access recursively",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnError(testError)
			},
			ExpectedErrorOutput: "Unexpected error while executing SQL query",
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating consumer subaccount label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, consumerSubaccountLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Once()
				return uidSvc
			},
			TenantSubscriptionRequest: regionalTenant,
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := testCase.RuntimeServiceFn()
			runtimeCtxSvc := testCase.RuntimeCtxServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			if testCase.TenantAccessMock != nil {
				testCase.TenantAccessMock(DBMock)
				defer DBMock.AssertExpectations(t)
			}

			service := NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, uuidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isSubscribeSuccessful, err := service.SubscribeTenant(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region, subscriptionAppName)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, isSubscribeSuccessful)

			mock.AssertExpectationsForObjects(t, runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, uuidSvc)
		})
	}
}

func TestUnSubscribeRegionalTenant(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	providerCtx := tenant.SaveToContext(ctx, providerInternalID, providerSubaccountID)
	consumerCtx := tenant.SaveToContext(providerCtx, subaccountTenantInternalID, subaccountTenantExtID)

	runtimeCtxPage := &model.RuntimeContextPage{
		Data: []*model.RuntimeContext{
			{
				ID:        runtimeCtxID,
				RuntimeID: providerRuntimeID,
				Key:       subscriptionLabelKey,
				Value:     consumerTenantID,
			},
		},
	}

	runtimeCtxFilter := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID)),
	}

	testCases := []struct {
		Name                      string
		Region                    string
		RuntimeServiceFn          func() *automock.RuntimeService
		RuntimeCtxServiceFn       func() *automock.RuntimeCtxService
		LabelServiceFn            func() *automock.LabelService
		UIDServiceFn              func() *automock.UidService
		TenantSvcFn               func() *automock.TenantService
		TenantSubscriptionRequest tenantfetchersvc.TenantSubscriptionRequest
		ExpectedErrorOutput       string
		IsSuccessful              bool
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				rtmCtxSvc.On("Delete", consumerCtx, runtimeCtxID).Return(nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              true,
		},
		{
			Name:                "Returns an error when getting internal provider tenant",
			Region:              tenantRegion,
			RuntimeServiceFn:    emptyRuntimeSvcFn,
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(nil, testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when can't find runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when could not list runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(nil, testError).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when could not list runtime contexts",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(nil, testError).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when getting consumer tenant",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: emptyRtmCtxSvcFn,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(nil, testError).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
		{
			Name:   "Returns an error when deleting runtime context",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				rtmCtxSvc.On("Delete", consumerCtx, runtimeCtxID).Return(testError).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetTenantByExternalID", ctx, providerSubaccountID).Return(providerTenant, nil).Once()
				tenantSvc.On("GetTenantByExternalID", providerCtx, subaccountTenantExtID).Return(consumerTenant, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:            emptyLabelSvcFn,
			UIDServiceFn:              emptyUIDSvcFn,
			TenantSubscriptionRequest: regionalTenant,
			ExpectedErrorOutput:       testError.Error(),
			IsSuccessful:              false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := testCase.RuntimeServiceFn()
			runtimeCtxSvc := testCase.RuntimeCtxServiceFn()
			labelSvc := testCase.LabelServiceFn()
			uidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			defer mock.AssertExpectationsForObjects(t, runtimeSvc, labelSvc, uidSvc, tenantSvc)

			service := NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, uidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isUnsubscribeSuccessful, err := service.UnsubscribeTenant(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, isUnsubscribeSuccessful)
			mock.AssertExpectationsForObjects(t, runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, uidSvc)
		})
	}
}
