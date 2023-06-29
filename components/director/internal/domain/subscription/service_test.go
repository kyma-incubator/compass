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
	uuid                   = "647af599-7f2d-485c-a63b-615b5ff6daf1"
	tenantRegion           = "eu-1"
	tenantRegionWithPrefix = subscription.RegionPrefix + tenantRegion
	subscriptionAppName    = "subscription-app-name-value"

	regionalTenantSubdomain    = "myregionaltenant"
	subaccountTenantExtID      = "32468d6e-f4cc-453c-beca-28c7bf55e0dd"
	subaccountTenantInternalID = "b6fe5f45-d8ba-4aa8-a949-5f8edb00d4a3"
	subscriptionProviderID     = "id-value!t12345"
	subscriptionID             = "320e9b8b-41b2-4492-b429-b570873f3041"
	subscriptionID2            = "2c606605-ed29-4840-b307-659f21dcba41"
	subscriptionID3            = "a436829f-e3cd-431b-847c-f4b792f55a14"
	providerSubaccountID       = "9cdbe10c-778c-432e-bf0e-e9686d04c679"
	providerInternalID         = "aa6301aa-7cf5-4335-82ae-35d078f8a2ed"
	consumerTenantID           = "ddf290d5-31c2-457e-a14d-461d3df95ac9"
	runtimeCtxID               = "cf226ea8-31f8-475d-bd28-5cba3df9c199"
	providerRuntimeID          = "96c85d13-22ee-4555-9b41-f5e364070c20"
	subscriptionsLabelID       = "123-123"
	runtimeM2MTableName        = "tenant_runtimes"

	subscriptionProviderIDLabelKey = "subscriptionProviderId"
	consumerSubaccountLabelKey     = "global_subaccount_id"
	subscriptionLabelKey           = "subscription"
	subscriptionAppNameLabelKey    = "runtimeType"
	tntSubdomain                   = "subdomain1"
)

var (
	testError              = errors.New("test error")
	notFoundErr            = apperrors.NewNotFoundErrorWithType(resource.Runtime)
	notFoundLabelErr       = apperrors.NewNotFoundErrorWithType(resource.Label)
	notFoundAppTemplateErr = apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate)

	regionalAndSubscriptionFiltersWithPrefix = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionProviderIDLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegionWithPrefix)),
	}
	regionalAndSubscriptionFilters = []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(subscriptionProviderIDLabelKey, fmt.Sprintf("\"%s\"", subscriptionProviderID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", tenantRegion)),
	}

	providerRuntime = &model.Runtime{ID: providerRuntimeID, Name: "provider-runtime-1"}

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

	subdomainLabel = &model.Label{
		Key:        subscription.SubdomainLabelKey,
		Value:      tntSubdomain,
		ObjectType: model.TenantLabelableObject,
	}

	subscriptionsLabelWithOneSubscription = &model.Label{
		ID:    subscriptionsLabelID,
		Key:   subscription.SubscriptionsLabelKey,
		Value: []interface{}{subscriptionID},
	}

	subscriptionsLabelWithTwoSubscriptions = &model.Label{
		ID:    subscriptionsLabelID,
		Key:   subscription.SubscriptionsLabelKey,
		Value: []interface{}{subscriptionID, subscriptionID2},
	}

	subscriptionsLabelInputWithOneSubscription = &model.LabelInput{
		Key:   subscription.SubscriptionsLabelKey,
		Value: []interface{}{subscriptionID},
	}

	subscriptionsLabelInputWithTwoSubscriptions = &model.LabelInput{
		Key:   subscription.SubscriptionsLabelKey,
		Value: []interface{}{subscriptionID, subscriptionID2},
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
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
				labelInput := &model.LabelInput{
					ObjectType: model.RuntimeContextLabelableObject,
					ObjectID:   runtimeCtxID,
					Key:        subscription.SubscriptionsLabelKey,
					Value:      []string{subscriptionID2},
				}
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, consumerSubaccountLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, labelInput).Return(nil).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Twice()
				return uidSvc
			},
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents) ON CONFLICT ( tenant_id, id ) DO NOTHING", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			IsSuccessful: true,
		},
		{
			Name:   "Succeeds when consumer is already subscribed",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithTwoSubscriptions).Return(nil)

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
		},
		{
			Name:   "Succeeds when consumer is already subscribed - missing subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				labelInput := &model.LabelInput{
					Key:        subscription.SubscriptionsLabelKey,
					ObjectID:   "id",
					ObjectType: model.RuntimeContextLabelableObject,
					Value:      []string{subscription.PreviousSubscriptionID, subscriptionID2},
				}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				lblSvc.On("CreateLabel", mock.Anything, subaccountTenantInternalID, uuid, labelInput).Return(nil)

				return lblSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Once()
				return uidSvc
			},
			IsSuccessful: true,
		},
		{
			Name:   "Returns error when consumer listing runtime contexts by filter fails",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(nil, testError).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, notFoundErr).Once()
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
			Name:   "Returns an error when could not get runtime",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, testError).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				svc := &automock.RuntimeCtxService{}
				svc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				svc := &automock.RuntimeCtxService{}
				svc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
			Name:   "Returns an error when creating runtime context",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return("", testError).Once()
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
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
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents) ON CONFLICT ( tenant_id, id ) DO NOTHING", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Error when getting subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(nil, testError)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name:   "Error when updating subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithTwoSubscriptions).Return(testError)

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name:   "Error when casting subscriptions label value",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				subscriptionsLabel := &model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: "invalid-value",
				}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(subscriptionsLabel, nil)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: errors.New("cannot cast \"subscriptions\" label value").Error(),
		},
		{
			Name:   "Error when creating subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{
					Data: []*model.RuntimeContext{{
						ID:        "id",
						RuntimeID: providerRuntimeID,
						Key:       consumerSubaccountLabelKey,
						Value:     consumerTenantID,
					}},
					PageInfo:   nil,
					TotalCount: 1,
				}, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				labelInput := &model.LabelInput{
					Key:        subscription.SubscriptionsLabelKey,
					ObjectID:   "id",
					ObjectType: model.RuntimeContextLabelableObject,
					Value:      []string{subscription.PreviousSubscriptionID, subscriptionID2},
				}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, "id", subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				lblSvc.On("CreateLabel", mock.Anything, subaccountTenantInternalID, uuid, labelInput).Return(testError)

				return lblSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Once()
				return uidSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name:   "Returns an error when creating tenant access recursively",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
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
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents) ON CONFLICT ( tenant_id, id ) DO NOTHING", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
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
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents) ON CONFLICT ( tenant_id, id ) DO NOTHING", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
					WithArgs(subaccountTenantInternalID, providerRuntimeID, false).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when creating subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("Create", consumerCtx, runtimeCtxInput).Return(runtimeCtxID, nil).Once()
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(consumerSubaccountLabelKey, fmt.Sprintf("\"%s\"", subaccountTenantExtID))}, 100, "").Return(&model.RuntimeContextPage{}, nil).Once()
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
				labelInput := &model.LabelInput{
					ObjectType: model.RuntimeContextLabelableObject,
					ObjectID:   runtimeCtxID,
					Key:        subscription.SubscriptionsLabelKey,
					Value:      []string{subscriptionID2},
				}
				labelSvc := &automock.LabelService{}
				labelSvc.On("UpsertLabel", providerCtx, providerSubaccountID, providerAppNameLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, consumerSubaccountLabelInput).Return(nil).Once()
				labelSvc.On("CreateLabel", consumerCtx, subaccountTenantInternalID, uuid, labelInput).Return(testError).Once()
				return labelSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Twice()
				return uidSvc
			},
			TenantAccessMock: func(dbMock testdb.DBMock) {
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents) ON CONFLICT ( tenant_id, id ) DO NOTHING", runtimeM2MTableName, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
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

			service := subscription.NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, nil, nil, nil, nil, uuidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isSubscribeSuccessful, err := service.SubscribeTenantToRuntime(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region, subscriptionAppName, subscriptionID2)

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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
		},
		{
			Name:   "Succeeds - missing subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
		},
		{
			Name:   "Succeeds - skip deletion because of multiple subscriptions",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithTwoSubscriptions, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithOneSubscription).Return(nil)

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, notFoundErr).Once()
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
			Name:   "Returns an error when could not get runtime",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(nil, testError).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Returns an error when deleting runtime context with missing subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
		},
		{
			Name:   "Error when getting subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(nil, testError)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name:   "error when casting subscriptions label value",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				subscriptionsLabel := &model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: "invalid-value",
				}
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabel, nil)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: "cannot cast \"subscriptions\" label value",
		},
		{
			Name:   "error while updating subscriptions label",
			Region: tenantRegion,
			RuntimeServiceFn: func() *automock.RuntimeService {
				provisioner := &automock.RuntimeService{}
				provisioner.On("GetByFilters", providerCtx, regionalAndSubscriptionFilters).Return(providerRuntime, nil).Once()
				return provisioner
			},
			RuntimeCtxServiceFn: func() *automock.RuntimeCtxService {
				rtmCtxSvc := &automock.RuntimeCtxService{}
				rtmCtxSvc.On("ListByFilter", consumerCtx, providerRuntimeID, runtimeCtxFilter, 100, "").Return(runtimeCtxPage, nil).Once()
				return rtmCtxSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", ctx, providerSubaccountID).Return(providerInternalID, nil).Once()
				tenantSvc.On("GetInternalTenant", providerCtx, subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.RuntimeContextLabelableObject, runtimeCtxID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithTwoSubscriptions, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithOneSubscription).Return(testError)

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
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

			service := subscription.NewService(runtimeSvc, runtimeCtxSvc, tenantSvc, labelSvc, nil, nil, nil, nil, uidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			isUnsubscribeSuccessful, err := service.UnsubscribeTenantFromRuntime(ctx, subscriptionProviderID, subaccountTenantExtID, providerSubaccountID, consumerTenantID, testCase.Region, subscriptionID2)

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
	subscriptionPayload := `{"name":"subscription-app-name-value", "display-name":"subscription-app-name-value"}`

	jsonAppCreateInput := fixJSONApplicationCreateInput(appTmplAppName)
	modelAppTemplate := fixModelAppTemplateWithAppInputJSON(appTmplID, appTmplName, jsonAppCreateInput)
	modelAppTemplateWithPlaceholders := fixModelAppTemplateWithPlaceholdersWithAppInputJSON(appTmplID, appTmplName, jsonAppCreateInput)
	modelAppFromTemplateInput := fixModelApplicationFromTemplateInput(appTmplName, subscriptionAppName, tntSubdomain, tenantRegion)
	modelAppFromTemplateSimplifiedInput := fixModelApplicationFromTemplateSimplifiedInput(appTmplName, subscriptionAppName, tntSubdomain, tenantRegion)
	gqlAppFromTemplateInput := fixGQLApplicationFromTemplateWithPayloadInput(appTmplName, subscriptionAppName, tntSubdomain, tenantRegion)
	gqlAppFromTemplateWithPayloadInput := fixGQLApplicationFromTemplateWithPayloadInput(appTmplName, subscriptionAppName, tntSubdomain, tenantRegion)
	modelAppFromTemplateInputWithPlaceholders := fixModelApplicationFromTemplateInputWithPlaceholders(appTmplName, subscriptionAppName, tntSubdomain, tenantRegion)
	modelAppFromTemplateInputWithEmptySubdomain := fixModelApplicationFromTemplateInput(appTmplName, subscriptionAppName, "", tenantRegion)
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
		SubscriptionPayload    string
		AppTemplateServiceFn   func() *automock.ApplicationTemplateService
		LabelServiceFn         func() *automock.LabelService
		UIDServiceFn           func() *automock.UidService
		TenantSvcFn            func() *automock.TenantService
		AppConverterFn         func() *automock.ApplicationConverter
		AppTemplConverterFn    func() *automock.ApplicationTemplateConverter
		AppSvcFn               func() *automock.ApplicationService
		ExpectedErrorOutput    string
		IsSuccessful           bool
		Repeats                int
	}{
		{
			Name:                "Succeeds",
			Region:              tenantRegionWithPrefix,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFiltersWithPrefix).Return(modelAppTemplate, nil).Once()
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      1,
		},
		{
			Name:                "Succeeds with subscription payload",
			Region:              tenantRegionWithPrefix,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFiltersWithPrefix).Return(modelAppTemplateWithPlaceholders, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplateWithPlaceholders, modelAppFromTemplateInputWithPlaceholders.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()

				return appConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateWithPlaceholders, gqlAppFromTemplateWithPayloadInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      1,
		},
		{
			Name:                "Succeeds",
			Region:              tenantRegionWithPrefix,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFiltersWithPrefix).Return(modelAppTemplate, nil).Once()
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
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      1,
		},
		{
			Name:                "Returns an error when can't find internal consumer tenant",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateWithPlaceholders, gqlAppFromTemplateWithPayloadInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
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
			Name:                "Returns an error when can't find app template",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateWithPlaceholders, gqlAppFromTemplateWithPayloadInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
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
			Name:                "Returns an error when fails to find an app template",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateWithPlaceholders, gqlAppFromTemplateWithPayloadInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
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
			Name:                "Returns an error when fails to get subdomain label",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplateWithPlaceholders, gqlAppFromTemplateWithPayloadInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(nil, testError).Once()

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:                "Success with empty subdomain value when it is missing",
			Region:              tenantRegionWithPrefix,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFiltersWithPrefix).Return(modelAppTemplate, nil).Once()
				appTemplateSvc.On("PrepareApplicationCreateInputJSON", modelAppTemplate, modelAppFromTemplateInputWithEmptySubdomain.Values).Return(jsonAppCreateInput, nil).Once()
				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", mock.Anything, gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return(appTmplID, nil).Once()

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(nil, notFoundLabelErr).Once()

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      1,
		},
		{
			Name:                "Returns an error when preparing application input json",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:                "Returns an error when creating graphql input from json",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(graphql.ApplicationRegisterInput{}, testError).Once()
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:                "Returns an error when creating input from graphql",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", ctxWithTenantMatcher(subaccountTenantInternalID), gqlAppCreateInput).Return(model.ApplicationRegisterInput{}, testError).Once()
				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")
				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:                "Returns an error when creating app from app template",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.On("CreateRegisterInputJSONToGQL", jsonAppCreateInput).Return(gqlAppCreateInput, nil).Once()
				appConv.On("CreateInputFromGraphQL", ctxWithTenantMatcher(subaccountTenantInternalID), gqlAppCreateInput).Return(modelAppCreateInput, nil).Once()
				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Once()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return([]*model.Application{}, nil).Once()
				appSvc.On("CreateFromTemplate", ctxWithTenantMatcher(subaccountTenantInternalID), modelAppCreateInputWithLabels, &appTmplID).Return("", testError).Once()
				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.TenantLabelableObject, subaccountTenantInternalID, subscription.SubdomainLabelKey).Return(subdomainLabel, nil).Once()

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			ExpectedErrorOutput: testError.Error(),
			IsSuccessful:        false,
			Repeats:             1,
		},
		{
			Name:                "Succeeds on multiple calls",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Times(repeats)

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Times(repeats)
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("UpdateLabel", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithTwoSubscriptions).Return(nil)

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      repeats,
		},
		{
			Name:                "Succeeds on multiple calls - subscription already exists",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Times(repeats)

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Times(repeats)
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(&model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: []interface{}{subscriptionID2},
				}, nil)

				return lblSvc
			},
			UIDServiceFn: unusedUUIDSvc,
			IsSuccessful: true,
			Repeats:      repeats,
		},
		{
			Name:                "Succeeds on multiple calls - subscriptions label not found",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
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
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Times(repeats)

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Times(repeats)
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelInput := &model.LabelInput{
					Key:        subscription.SubscriptionsLabelKey,
					ObjectID:   appTmplID,
					ObjectType: model.ApplicationLabelableObject,
					Value:      []string{subscription.PreviousSubscriptionID, subscriptionID2},
				}
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				lblSvc.On("CreateLabel", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, uuid, labelInput).Return(nil)

				return lblSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Times(repeats)
				return uidSvc
			},
			IsSuccessful: true,
			Repeats:      repeats,
		},
		{
			Name:                "Error when getting subscriptions label",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Twice()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Twice()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Twice()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Twice()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(nil, testError)
				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
			Repeats:             2,
		},
		{
			Name:                "Error when creating subscriptions label",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Twice()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Twice()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Twice()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Twice()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				labelInput := &model.LabelInput{
					Key:        subscription.SubscriptionsLabelKey,
					ObjectID:   appTmplID,
					ObjectType: model.ApplicationLabelableObject,
					Value:      []string{subscription.PreviousSubscriptionID, subscriptionID2},
				}
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				lblSvc.On("CreateLabel", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, uuid, labelInput).Return(testError)

				return lblSvc
			},
			UIDServiceFn: func() *automock.UidService {
				uidSvc := &automock.UidService{}
				uidSvc.On("Generate").Return(uuid).Twice()
				return uidSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
			Repeats:             2,
		},
		{
			Name:                "Error when updating subscriptions label",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Twice()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Twice()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Twice()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Twice()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("UpdateLabel", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithTwoSubscriptions).Return(testError)

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
			Repeats:             2,
		},
		{
			Name:                "Error when casting subscriptions label value",
			Region:              tenantRegion,
			SubscriptionPayload: subscriptionPayload,
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				appTemplateSvc := &automock.ApplicationTemplateService{}
				appTemplateSvc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Twice()
				appTemplateSvc.AssertNotCalled(t, "PrepareApplicationCreateInputJSON")

				return appTemplateSvc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Twice()
				return tenantSvc
			},
			AppConverterFn: func() *automock.ApplicationConverter {
				appConv := &automock.ApplicationConverter{}
				appConv.AssertNotCalled(t, "CreateRegisterInputJSONToGQL")
				appConv.AssertNotCalled(t, "CreateInputFromGraphQL")

				return appConv
			},
			AppTemplConverterFn: func() *automock.ApplicationTemplateConverter {
				appTemplateConv := &automock.ApplicationTemplateConverter{}
				appTemplateConv.On("ApplicationFromTemplateInputFromGraphQL", modelAppTemplate, gqlAppFromTemplateInput).Return(modelAppFromTemplateSimplifiedInput, nil).Twice()

				return appTemplateConv
			},
			AppSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Twice()
				appSvc.AssertNotCalled(t, "CreateFromTemplate")

				return appSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				subscirptionsLabel := &model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: "invalid-value",
				}
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", ctxWithTenantMatcher(subaccountTenantInternalID), subaccountTenantInternalID, model.ApplicationLabelableObject, appTmplID, subscription.SubscriptionsLabelKey).Return(subscirptionsLabel, nil)

				return lblSvc
			},
			UIDServiceFn:        unusedUUIDSvc,
			IsSuccessful:        false,
			ExpectedErrorOutput: errors.New("cannot cast \"subscriptions\" label value").Error(),
			Repeats:             2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateSvc := testCase.AppTemplateServiceFn()
			labelSvc := testCase.LabelServiceFn()
			appConv := testCase.AppConverterFn()
			appTemplConv := testCase.AppTemplConverterFn()
			appSvc := testCase.AppSvcFn()
			uuidSvc := testCase.UIDServiceFn()
			tenantSvc := &automock.TenantService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}

			service := subscription.NewService(nil, nil, tenantSvc, labelSvc, appTemplateSvc, appConv, appTemplConv, appSvc, uuidSvc, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			for count := 0; count < testCase.Repeats; count++ {
				// WHEN
				isSubscribeSuccessful, err := service.SubscribeTenantToApplication(context.TODO(), subscriptionProviderID, subaccountTenantExtID, consumerTenantID, testCase.Region, subscriptionAppName, subscriptionID2, testCase.SubscriptionPayload)

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
		LabelServiceFn       func() *automock.LabelService
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				return lblSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Success - missing subscriptions label",
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(nil, notFoundLabelErr)
				return lblSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Success - skipping deletion because of more than one subscriptions",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithTwoSubscriptions, nil)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithTwoSubscriptions, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithOneSubscription).Return(nil).Twice()
				return lblSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Success - skipping deletion because of more than one subscriptions and provided subscriptionID does not exist, but previous subscription id exists",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(&model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: []interface{}{subscription.PreviousSubscriptionID, subscriptionID},
				}, nil)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(&model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: []interface{}{subscription.PreviousSubscriptionID, subscriptionID},
				}, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithOneSubscription).Return(nil).Twice()
				return lblSvc
			},
			IsSuccessful: true,
		},
		{
			Name: "Success - skipping deletion because of more than one subscriptions and provided subscriptionID and previous subscription do not exist",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(&model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: []interface{}{subscriptionID, subscriptionID3},
				}, nil)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(&model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: []interface{}{subscriptionID, subscriptionID3},
				}, nil)
				return lblSvc
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				return lblSvc
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				return lblSvc
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				return lblSvc
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				return lblSvc
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, nil)
				return lblSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when deleting application - missing subscriptions label",
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
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, notFoundLabelErr)
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appSecondID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithOneSubscription, notFoundLabelErr)
				return lblSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when getting subscriptions label",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(nil, testError)
				return lblSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: testError.Error(),
		},
		{
			Name: "Error when casting subscriptions label",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				subscriptionsLabel := &model.Label{
					ID:    subscriptionsLabelID,
					Key:   subscription.SubscriptionsLabelKey,
					Value: "invalid-value",
				}
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabel, nil)
				return lblSvc
			},
			IsSuccessful:        false,
			ExpectedErrorOutput: errors.New("cannot cast \"subscriptions\" label value").Error(),
		},
		{
			Name: "Error when updating subscriptions label",
			AppTemplateServiceFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("GetByFilters", context.TODO(), regionalAndSubscriptionFilters).Return(modelAppTemplate, nil).Once()
				return svc
			},
			AppSvcFn: func() *automock.ApplicationService {
				svc := &automock.ApplicationService{}
				svc.On("ListAll", ctxWithTenantMatcher(subaccountTenantInternalID)).Return(modelApps, nil).Once()
				return svc
			},
			TenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetInternalTenant", context.TODO(), subaccountTenantExtID).Return(subaccountTenantInternalID, nil).Once()
				return tenantSvc
			},
			LabelServiceFn: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("GetByKey", mock.Anything, subaccountTenantInternalID, model.ApplicationLabelableObject, appFirstID, subscription.SubscriptionsLabelKey).Return(subscriptionsLabelWithTwoSubscriptions, nil)
				lblSvc.On("UpdateLabel", mock.Anything, subaccountTenantInternalID, subscriptionsLabelID, subscriptionsLabelInputWithOneSubscription).Return(testError)
				return lblSvc
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
			lblSvc := testCase.LabelServiceFn()
			service := subscription.NewService(nil, nil, tenantSvc, lblSvc, appTemplateSvc, nil, nil, appSvc, nil, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

			// WHEN
			successful, err := service.UnsubscribeTenantFromApplication(context.TODO(), subscriptionProviderID, subaccountTenantExtID, tenantRegion, subscriptionID2)

			// THEN
			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.IsSuccessful, successful)

			mock.AssertExpectationsForObjects(t, appTemplateSvc, tenantSvc, appSvc, lblSvc)
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

			service := subscription.NewService(rtmService, nil, nil, nil, appTemplateSvc, nil, nil, nil, nil, consumerSubaccountLabelKey, subscriptionLabelKey, subscriptionAppNameLabelKey, subscriptionProviderIDLabelKey)

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
