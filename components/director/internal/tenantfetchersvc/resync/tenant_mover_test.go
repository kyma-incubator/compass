package resync_test

import (
	"context"
	"errors"
	"testing"

	domaintenant "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantMover_TenantsToMove(t *testing.T) {
	const (
		timestamp = "1234567899987"

		sourceParentTenantID = "d65e725e-f9a2-4a2a-b349-6487ab084ebc"
		targetParentTenantID = "7fb04de7-b397-4b86-98f5-f3af8ac1da0c"
	)

	ctx := context.TODO()
	tenantConverter := domaintenant.NewConverter()
	jobConfig := configForTenantType(tenant.Subaccount)

	movedSubaccount1 := model.MovedSubaccountMappingInput{
		TenantMappingInput: fixBusinessTenantMappingInput("1", provider, "subdomain-1", "", sourceParentTenantID, tenant.Subaccount),
		SubaccountID:       "1",
		SourceTenant:       sourceParentTenantID,
		TargetTenant:       targetParentTenantID,
	}
	movedSubaccount2 := model.MovedSubaccountMappingInput{
		TenantMappingInput: fixBusinessTenantMappingInput("2", provider, "subdomain-2", "", sourceParentTenantID, tenant.Subaccount),
		SubaccountID:       "2",
		SourceTenant:       sourceParentTenantID,
		TargetTenant:       targetParentTenantID,
	}

	event1 := fixEvent(t, "Subaccount", movedSubaccount1.TenantMappingInput.Parent, movedEventFieldsFromTenant(jobConfig.TenantFieldMapping, jobConfig.MovedSubaccountsFieldMapping, movedSubaccount1))
	event2 := fixEvent(t, "Subaccount", movedSubaccount2.TenantMappingInput.Parent, movedEventFieldsFromTenant(jobConfig.TenantFieldMapping, jobConfig.MovedSubaccountsFieldMapping, movedSubaccount2))

	pageOneQueryParams := QueryParams{
		jobConfig.PageSizeField:                        "1",
		jobConfig.PageNumField:                         "1",
		jobConfig.TimestampField:                       timestamp,
		jobConfig.EventsConfig.QueryConfig.RegionField: region,
	}

	testCases := []struct {
		name               string
		jobConfigFn        func() JobConfig
		directorClientFn   func() *automock.DirectorGraphQLClient
		apiClientFn        func() *automock.EventAPIClient
		runtimeSvcFn       func() *automock.RuntimeService
		labelRepoFn        func() *automock.LabelRepo
		tenantStorageSvcFn func() *automock.TenantStorageService
		expectedTenants    []model.MovedSubaccountMappingInput
		expectedErrMsg     string
	}{
		{
			name: "Success when only one page is returned for moved tenants events",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			apiClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				return client
			},
			expectedTenants: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when two pages are returned for moved tenants events",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			apiClientFn: func() *automock.EventAPIClient {
				pageTwoQueryParams := QueryParams{
					"pageSize":  "1",
					"pageNum":   "2",
					"region":    region,
					"timestamp": timestamp,
				}

				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 2, 2), nil).Once()
				client.On("FetchTenantEventsPage", MovedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 2, 2), nil).Once()

				return client
			},
			expectedTenants: []model.MovedSubaccountMappingInput{movedSubaccount1, movedSubaccount2},
		},
		{
			name: "Fail when fetching moved tenants events returns an error",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			apiClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", MovedSubaccountType, pageOneQueryParams).Return(nil, errors.New("failed to get moved")).Once()
				return client
			},
			expectedErrMsg: "while fetching moved tenants",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.jobConfigFn()
			directorClient := tc.directorClientFn()
			eventAPIClient := tc.apiClientFn()
			storageSvc := &automock.TenantStorageService{}
			runtimeSvc := &automock.RuntimeService{}
			labelRepo := &automock.LabelRepo{}
			txGen := txtest.NewTransactionContextGenerator(nil)
			persist, transact := txGen.ThatDoesntStartTransaction()
			defer mock.AssertExpectationsForObjects(t, directorClient, eventAPIClient, storageSvc, runtimeSvc, labelRepo, persist, transact)

			mover := NewSubaccountsMover(cfg, transact, directorClient, eventAPIClient, tenantConverter, storageSvc, runtimeSvc, labelRepo)
			res, err := mover.TenantsToMove(ctx, region, timestamp)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.expectedTenants, res)
			}
		})
	}
}

func TestTenantMover_MoveTenants(t *testing.T) {
	const (
		provider = "external-service"

		subaccountExternalTenant = "3daf8389-7f3e-41f7-94d9-75576ac80bee"
		subaccountInternalTenant = "f384fa42-41f2-4490-b9b7-95d8ee03f785"

		sourceParentTenantID = "d65e725e-f9a2-4a2a-b349-6487ab084ebc"
		targetParentTenantID = "7fb04de7-b397-4b86-98f5-f3af8ac1da0c"
	)

	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(errors.New("test err"))

	// GIVEN
	tenantConverter := domaintenant.NewConverter()

	var subaccountFromDB *model.BusinessTenantMapping
	init := func() {
		subaccountFromDB = &model.BusinessTenantMapping{
			ID:             subaccountInternalTenant,
			Name:           subaccountExternalTenant,
			ExternalTenant: subaccountExternalTenant,
			Parent:         sourceParentTenantID,
			Type:           tenant.Subaccount,
			Provider:       provider,
		}

	}

	targetParent := &model.BusinessTenantMapping{
		ID:             "05a2a8ed-66b9-4978-86b9-5167fb43520f",
		ExternalTenant: targetParentTenantID,
	}
	sourceParent := &model.BusinessTenantMapping{
		ID:             "83711a81-4b07-4b21-b1b2-042135800039",
		ExternalTenant: sourceParentTenantID,
	}

	movedSubaccountInput := graphql.BusinessTenantMappingInput{
		Name:           subaccountExternalTenant,
		ExternalTenant: subaccountExternalTenant,
		Parent:         str.Ptr(targetParentTenantID),
		Subdomain:      str.Ptr(""),
		Region:         str.Ptr(""),
		Type:           string(tenant.Subaccount),
		Provider:       provider,
	}

	movedSubaccount1 := model.MovedSubaccountMappingInput{
		TenantMappingInput: fixBusinessTenantMappingInput(subaccountExternalTenant, provider, "", "", sourceParentTenantID, tenant.Subaccount),
		SubaccountID:       subaccountExternalTenant,
		SourceTenant:       sourceParentTenantID,
		TargetTenant:       targetParentTenantID,
	}

	ctxWithSubaccountMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		tenantID, err := domaintenant.LoadFromContext(ctx)
		require.NoError(t, err)
		eq := tenantID == subaccountInternalTenant
		return eq
	})

	subaccountRuntime := &model.Runtime{
		ID: "35ab420d-1c67-4d06-9567-dbba8f95ea95",
	}

	testCases := []struct {
		name               string
		jobConfigFn        func() JobConfig
		transactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		directorClientFn   func() *automock.DirectorGraphQLClient
		runtimeSvcFn       func() *automock.RuntimeService
		labelRepoFn        func() *automock.LabelRepo
		tenantStorageSvcFn func([]model.MovedSubaccountMappingInput) *automock.TenantStorageService
		tenantsInput       []model.MovedSubaccountMappingInput
		expectedErrMsg     string
	}{
		{
			name: "Success when subaccount is moved to an existing parent tenant",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn: txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient {
				client := &automock.DirectorGraphQLClient{}
				client.On("UpdateTenant", txtest.CtxWithDBMatcher(), subaccountInternalTenant, movedSubaccountInput).Return(nil)
				return client
			},
			runtimeSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				var emptyFilters []*labelfilter.LabelFilter

				svc.On("ListByFilters", ctxWithSubaccountMatcher, emptyFilters).Return(nil, nil).Once()
				return svc
			},
			labelRepoFn: func() *automock.LabelRepo { return &automock.LabelRepo{} },
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{subaccountFromDB}, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when subaccount is created in the correct parent tenant",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn: txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient {
				client := &automock.DirectorGraphQLClient{}

				client.On("WriteTenants", txtest.CtxWithDBMatcher(), []graphql.BusinessTenantMappingInput{movedSubaccountInput}).Return(nil).Once()
				return client
			},
			runtimeSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			labelRepoFn:  func() *automock.LabelRepo { return &automock.LabelRepo{} },
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{}, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when subaccount is not in a formation",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn: txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient {
				client := &automock.DirectorGraphQLClient{}
				client.On("UpdateTenant", txtest.CtxWithDBMatcher(), subaccountInternalTenant, movedSubaccountInput).Return(nil)
				return client
			},
			runtimeSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				var emptyFilters []*labelfilter.LabelFilter

				svc.On("ListByFilters", ctxWithSubaccountMatcher, emptyFilters).Return([]*model.Runtime{subaccountRuntime}, nil).Once()
				return svc
			},
			labelRepoFn: func() *automock.LabelRepo {
				repo := &automock.LabelRepo{}
				repo.On("GetScenarioLabelsForRuntimes", txtest.CtxWithDBMatcher(), sourceParent.ID, []string{subaccountRuntime.ID}).Return(nil, nil).Once()
				return repo
			},
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{subaccountFromDB}, nil).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), sourceParentTenantID).Return(sourceParent, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when subaccount is in the default formation",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn: txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient {
				client := &automock.DirectorGraphQLClient{}
				client.On("UpdateTenant", txtest.CtxWithDBMatcher(), subaccountInternalTenant, movedSubaccountInput).Return(nil)
				return client
			},
			runtimeSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				var emptyFilters []*labelfilter.LabelFilter

				svc.On("ListByFilters", ctxWithSubaccountMatcher, emptyFilters).Return([]*model.Runtime{subaccountRuntime}, nil).Once()
				return svc
			},
			labelRepoFn: func() *automock.LabelRepo {
				repo := &automock.LabelRepo{}
				defaultLabel := []model.Label{{Value: []interface{}{"DEFAULT"}}}
				repo.On("GetScenarioLabelsForRuntimes", txtest.CtxWithDBMatcher(), sourceParent.ID, []string{subaccountRuntime.ID}).Return(defaultLabel, nil).Once()
				return repo
			},
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{subaccountFromDB}, nil).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), sourceParentTenantID).Return(sourceParent, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when target parent tenant does not exist: tenant is skipped",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn:  txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			runtimeSvcFn:     func() *automock.RuntimeService { return &automock.RuntimeService{} },
			labelRepoFn:      func() *automock.LabelRepo { return &automock.LabelRepo{} },
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{}, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Success when subaccount is already moved",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn:  txGen.ThatSucceeds,
			directorClientFn: func() *automock.DirectorGraphQLClient { return &automock.DirectorGraphQLClient{} },
			runtimeSvcFn:     func() *automock.RuntimeService { return &automock.RuntimeService{} },
			labelRepoFn:      func() *automock.LabelRepo { return &automock.LabelRepo{} },
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				subaccountFromDBWithNewParent := *subaccountFromDB
				subaccountFromDBWithNewParent.Parent = targetParent.ID
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{&subaccountFromDBWithNewParent}, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
		},
		{
			name: "Fail when subaccount is in formation",
			jobConfigFn: func() JobConfig {
				return configForTenantType(tenant.Subaccount)
			},
			transactionerFn: txGen.ThatDoesntExpectCommit,
			directorClientFn: func() *automock.DirectorGraphQLClient {
				client := &automock.DirectorGraphQLClient{}
				//client.On("UpdateTenant", txtest.CtxWithDBMatcher(), subaccountInternalTenant, movedSubaccountInput).Return(nil)
				return client
			},
			runtimeSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				var emptyFilters []*labelfilter.LabelFilter

				svc.On("ListByFilters", ctxWithSubaccountMatcher, emptyFilters).Return([]*model.Runtime{subaccountRuntime}, nil).Once()
				return svc
			},
			labelRepoFn: func() *automock.LabelRepo {
				repo := &automock.LabelRepo{}
				label := []model.Label{{Value: []interface{}{"my-scenario"}}}
				repo.On("GetScenarioLabelsForRuntimes", txtest.CtxWithDBMatcher(), sourceParent.ID, []string{subaccountRuntime.ID}).Return(label, nil).Once()
				return repo
			},
			tenantStorageSvcFn: func(input []model.MovedSubaccountMappingInput) *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				tnt := input[0]
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.TargetTenant}).Return([]*model.BusinessTenantMapping{targetParent}, nil).Once()
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{tnt.SubaccountID}).Return([]*model.BusinessTenantMapping{subaccountFromDB}, nil).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), sourceParentTenantID).Return(sourceParent, nil).Once()
				return svc
			},
			tenantsInput: []model.MovedSubaccountMappingInput{movedSubaccount1},
			expectedErrMsg: "is in scenario",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.jobConfigFn()
			init()
			eventAPIClient := &automock.EventAPIClient{}
			persist, transact := tc.transactionerFn()
			directorClient := tc.directorClientFn()
			storageSvc := tc.tenantStorageSvcFn(tc.tenantsInput)
			runtimeSvc := tc.runtimeSvcFn()
			labelRepo := tc.labelRepoFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, directorClient, eventAPIClient, storageSvc, runtimeSvc, labelRepo)

			mover := NewSubaccountsMover(cfg, transact, directorClient, eventAPIClient, tenantConverter, storageSvc, runtimeSvc, labelRepo)
			err := mover.MoveTenants(ctx, tc.tenantsInput)
			if len(tc.expectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func movedEventFieldsFromTenant(tenantFieldMapping TenantFieldMapping, movedSAFieldMapping MovedSubaccountsFieldMapping, tenantInput model.MovedSubaccountMappingInput) map[string]string {
	eventFields := eventFieldsFromTenant(tenant.Subaccount, tenantFieldMapping, tenantInput.TenantMappingInput)
	eventFields[movedSAFieldMapping.SourceTenant] = tenantInput.SourceTenant
	eventFields[movedSAFieldMapping.TargetTenant] = tenantInput.TargetTenant
	eventFields[movedSAFieldMapping.SubaccountID] = tenantInput.SubaccountID
	return eventFields
}
