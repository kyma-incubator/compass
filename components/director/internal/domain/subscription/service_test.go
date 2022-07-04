package subscription_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	uuid                = "647af599-7f2d-485c-a63b-615b5ff6daf1"
	tenantRegion        = "eu-1"
	subscriptionAppName = "subscription-app-name-value"

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

	subscriptionProviderIDLabelKey = "subscriptionProviderId"
	consumerSubaccountLabelKey     = "consumer_subaccount_id"
	subscriptionLabelKey           = "subscription"
	subscriptionAppNameLabelKey    = "runtimeType"
)

var (
	testError              = errors.New("test error")
	notFoundErr            = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	notFoundAppTemplateErr = apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate)

	regionalAndSubscriptionFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionProviderIDLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}

	providerRuntimes = []*model.Runtime{
		{
			ID:   providerRuntimeID,
			Name: "provider-runtime-1",
		},
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
		Name                string
		Region              string
		RuntimeServiceFn    func() *automock.RuntimeService
		RuntimeCtxServiceFn func() *automock.RuntimeCtxService
		LabelServiceFn      func() *automock.LabelService
		UIDServiceFn        func() *automock.UidService
		TenantSvcFn         func() *automock.TenantService
		TenantAccessMock    func(DBMock testdb.DBMock)
		ExpectedErrorOutput string
		IsSuccessful        bool
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			IsSuccessful: true,
		},
		{
			Name:                "Returns an error when getting internal provider tenant",
			Region:              tenantRegion,
			RuntimeServiceFn:    unusedRuntimeSvc,
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when can't find runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   false,
		},
		{
			Name:   "Returns an error when could not list runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, testError).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when could not get lowest owner of the runtime",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when could not upsert provider app name label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				return tenantSvc
			},

			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when getting consumer tenant",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return("", testError).Once()
				return tenantSvc
			},

			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating runtime context",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return("", testError).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating tenant access recursively",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: unusedUUIDSvc,
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
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", providerCtx, resource.Runtime, providerRuntimeID).Return(providerSubaccountID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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

			service := subscription.NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, nil, nil, nil, uuidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isSubscribeSuccessful, err := service.SubscribeTenantToRuntime(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region, subscriptionAppName)

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
		Name                string
		Region              string
		RuntimeServiceFn    func() *automock.RuntimeService
		RuntimeCtxServiceFn func() *automock.RuntimeCtxService
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
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
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
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   true,
		},
		{
			Name:                "Returns an error when getting internal provider tenant",
			Region:              tenantRegion,
			RuntimeServiceFn:    unusedRuntimeSvc,
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when can't find runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, notFoundErr).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   false,
		},
		{
			Name:   "Returns an error when could not list runtimes",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, testError).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when could not list runtime contexts",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(nil, testError).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when getting consumer tenant",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: unusedRuntimeContextSvc,
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return("", testError).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when deleting runtime context",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("ListByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntimes, nil).Once()
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
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
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

			service := subscription.NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, nil, nil, nil, uidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isUnsubscribeSuccessful, err := service.UnsubscribeTenantFromRuntime(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region)

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

func TestSubscribeTenantToApplication(t *testing.T) {
	appTmplName := "test-app-tmpl"
	appTmplAppName := "{{name}}"
	appTmplID := "123-456-789"
	repeats := 5

	jsonAppCreateInput := fixJSONApplicationCreateInput(appTmplAppName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(appTmplID, appTmplName, jsonAppCreateInput)
	modelAppFromTemplateInput := fixModelApplicationFromTemplateInput(appTmplName, subscriptionAppName)
	gqlAppCreateInput := fixGQLApplicationCreateInput(appTmplName)
	modelAppCreateInput := fixModelApplicationCreateInput(appTmplName)
	modelAppCreateInputWithLabels := fixModelApplicationCreateInputWithLabels(appTmplName, subaccountTenantExtID)
	modelApps := []*model.Application{
		fixModelApplication(appTmplID, appTmplName, appTmplID),
	}

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
		Repeats                int
	}{
		{
			Name:   "Succeeds",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   true,
			Repeats:        1,
		},
		{
			Name:   "Returns an error when can't find internal consumer tenant",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return("", testError).Once()
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
				appSvc.AssertNotCalled(t, "ListAll")
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Returns an error when can't find app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(nil, notFoundErr).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetInternalTenant")
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
				appSvc.AssertNotCalled(t, "ListAll")
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   false,
			Repeats:        1,
		},
		{
			Name:   "Returns an error when fails finding app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(nil, testError).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.AssertNotCalled(t, "GetInternalTenant")
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
				appSvc.AssertNotCalled(t, "ListAll")
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Returns an error when preparing application input json",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return("", testError).Once()

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Returns an error when creating graphql input from json",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Returns an error when creating input from graphql",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", ctxWithTenantMatcher(subaccountTenantInternalID), gqlAppCreateInput).Return(model.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")
				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Returns an error when creating app from app template",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInput.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", ctxWithTenantMatcher(subaccountTenantInternalID), gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return("", testError).Once()
				return appSvc
			},
			LabelServiceFn:      unusedLabelSvc,
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:   "Succeeds on on multiple calls",
			Region: tenantRegion,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Times(repeats)
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Times(repeats)
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
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Times(repeats)
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: unusedLabelSvc,
			UIDServiceFn:   unusedUUIDSvc,
			IsSuccessful:   true,
			Repeats:        repeats,
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

			service := subscription.NewService(nil, nil, tenantSvc, labelSvc, appTemplateSvc, appConv, appSvc, uuidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			for count := 0; count < testCase.Repeats; count++ {
				// WHEN
				isSubscribeSuccessful, err := service.SubscribeTenantToApplication(context.TODO(), subscriptionProviderID, subaccountTenantExtID, consumerTenantID, testCase.Region, subscriptionAppName)

				// THEN
				if len(testCase.ExpectedErrorOutput) > 0 {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
				} else {
					assert.NoError(t, err)
				}

				assert.Equal(t, testCase.IsSuccessful, isSubscribeSuccessful)
			}

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
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				svc.On("Delete", ctxWithTenantMatcher(subaccountTenantInternalID), appFirstID).Return(nil).Once()
				svc.On("Delete", ctxWithTenantMatcher(subaccountTenantInternalID), appSecondID).Return(nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Error when fails to get internal tenant",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
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
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return("", testError).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when getting app template by filters",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(nil, testError).Once()
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
				tenantSvc.AssertNotCalled(t, "GetInternalTenant")
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when app template by filters is not found",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(nil, notFoundAppTemplateErr).Once()
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
				tenantSvc.AssertNotCalled(t, "GetInternalTenant")
				return tenantSvc
			},
			IsSuccessful: false,
		},
		{
			Name: "Error when listing applications",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(nil, testError).Once()
				svc.AssertNotCalled(t, "Delete")
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when deleting application",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				svc.On("Delete", ctxWithTenantMatcher(subaccountTenantInternalID), appFirstID).Return(nil).Once()
				svc.On("Delete", ctxWithTenantMatcher(subaccountTenantInternalID), appSecondID).Return(testError).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
			service := subscription.NewService(nil, nil, tenantSvc, nil, appTemplateSvc, nil, appSvc, nil, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			successful, err := service.UnsubscribeTenantFromApplication(context.TODO(), subscriptionProviderID, subaccountTenantExtID, tenantRegion)

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

			service := subscription.NewService(rtmService, nil, nil, nil, appTemplateSvc, nil, nil, nil, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

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

func ctxWithTenantMatcher(expectedTenantID string) interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		tenantID, err := tenant.LoadFromContext(ctx)
		return err == nil && tenantID == expectedTenantID
	})
}
