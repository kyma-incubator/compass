package tenantfetcher_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	domainTenant "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	tenantInsertChunkSize = 500
)

var (
	tenantConverter = domainTenant.NewConverter()
)

func TestService_SyncSubaccountOnDemandTenants(t *testing.T) {
	// GIVEN
	subaccountID := "subaccount-1"
	provider := "default"

	var (
		tenantFieldMapping tenantfetcher.TenantFieldMapping

		busTenant1GUID string
		parentTenant   model.BusinessTenantMappingInput
		parentTenants  []model.BusinessTenantMappingInput

		busSubaccount                      model.BusinessTenantMappingInput
		busSubaccounts                     []model.BusinessTenantMappingInput
		businessSubaccount1BusinessMapping *model.BusinessTenantMapping

		createSubaccountEventFields    map[string]string
		unsuspendSubaccountEventFields map[string]string

		createSubaccountEvent    []byte
		unsuspendSubaccountEvent []byte
	)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}

	pageOneQueryParams := tenantfetcher.QueryParams{
		"entityId": subaccountID,
		"pageSize": "1",
		"pageNum":  "1",
	}

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	beforeEach := func() {
		tenantFieldMapping = tenantfetcher.TenantFieldMapping{
			DetailsField:           "eventData",
			EventsField:            "events",
			NameField:              "name",
			IDField:                "id",
			SubaccountIDField:      "entityId",
			CustomerIDField:        "customerId",
			SubdomainField:         "subdomain",
			EntityTypeField:        "type",
			RegionField:            "region",
			TotalPagesField:        "pages",
			TotalResultsField:      "total",
			GlobalAccountGUIDField: "globalAccountGUID",
		}

		busTenant1GUID = "d1f08f02-2fda-4511-962a-17fd1f1aa477"
		parentTenant = fixBusinessTenantMappingInput(busTenant1GUID, busTenant1GUID, provider, "", "", "", tenant.Account)
		parentTenants = []model.BusinessTenantMappingInput{parentTenant}

		busSubaccount = fixBusinessTenantMappingInput("foo", subaccountID, provider, "subdomain", "test-region", parentTenant.ExternalTenant, tenant.Subaccount)
		busSubaccounts = []model.BusinessTenantMappingInput{busSubaccount}
		businessSubaccount1BusinessMapping = busSubaccount.ToBusinessTenantMapping(busTenant1GUID)

		createSubaccountEventFields = map[string]string{
			tenantFieldMapping.IDField:         subaccountID,
			tenantFieldMapping.NameField:       "foo",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain",
			tenantFieldMapping.EntityTypeField: "Subaccount",
		}
		unsuspendSubaccountEventFields = map[string]string{
			tenantFieldMapping.SubaccountIDField: subaccountID,
			tenantFieldMapping.NameField:         "foo",
			tenantFieldMapping.RegionField:       "test-region",
			tenantFieldMapping.SubdomainField:    "subdomain",
			tenantFieldMapping.EntityTypeField:   "Subaccount",
		}

		createSubaccountEvent = fixEvent(t, "CreateSubaccount", busTenant1GUID, createSubaccountEventFields)
		unsuspendSubaccountEvent = fixEvent(t, "UnsuspendSubaccount", busTenant1GUID, unsuspendSubaccountEventFields)
	}

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantStorageSvcFn func() *automock.TenantStorageService
		APIClientFn        func() *automock.EventAPIClient
		GqlClientFn        func() *automock.DirectorGraphQLClient
		ExpectedErrorMsg   error
	}{
		{
			Name:            "Success processing create a tenant for a subaccount request",
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenant.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, parentTenant.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(createSubaccountEvent), 1, 1), nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccounts[:1]...)
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedErrorMsg: nil,
		},
		{
			Name:            "Success processing unsuspended tenant for a subaccount fetch request",
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenant.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, parentTenant.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(createSubaccountEvent, unsuspendSubaccountEvent), 1, 1), nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccounts[:1]...)
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedErrorMsg: nil,
		},
		{
			Name:            "Success when tenant already exists",
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), businessSubaccount1BusinessMapping.ExternalTenant).Return(businessSubaccount1BusinessMapping, nil).Once()
				return svc
			},
			APIClientFn:      UnusedEventAPIClient,
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: nil,
		},
		{
			Name:            "Error when getting subaccount from database throws unexpected error",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).Return(nil, testErr).Once()
				return svc
			},
			APIClientFn:      UnusedEventAPIClient,
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: testErr,
		},
		{
			Name:            "Error when parent retrieval fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenant.ExternalTenant).
					Return(nil, testErr).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(createSubaccountEvent), 1, 1), nil).Once()
				return client
			},
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: testErr,
		},
		{
			Name:            "Error when no event found for subaccount",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), businessSubaccount1BusinessMapping.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, businessSubaccount1BusinessMapping.ExternalTenant)).Once()
				return svc
			}, APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType)
				return client
			},
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: errors.New("no create events for subaccount with ID subaccount-1 were found"),
		},
		{
			Name:            "Error when events page is empty",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), businessSubaccount1BusinessMapping.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, businessSubaccount1BusinessMapping.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()
				return client
			},
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: errors.New("no create events for subaccount with ID subaccount-1 were found"),
		},
		{
			Name:            "Error when cannot fetch events page for a subaccount creation",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), businessSubaccount1BusinessMapping.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, businessSubaccount1BusinessMapping.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Times(tenantfetcher.RetryAttempts)
				return client
			},
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: errors.New("while fetching subaccount by ID: All attempts fail"),
		},
		{
			Name:               "Error when couldn't start transaction",
			TransactionerFn:    txGen.ThatFailsOnBegin,
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			APIClientFn:        UnusedEventAPIClient,
			GqlClientFn:        UnusedGQLClient,
			ExpectedErrorMsg:   testErr,
		},
		{
			Name:            "Success when couldn't commit empty transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenant.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, parentTenant.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(createSubaccountEvent), 1, 1), nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccount)
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)).Return(nil)
				return gqlClient
			},
		},
		{
			Name:            "Error when tenant creation fails",
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), parentTenant.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, parentTenant.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(createSubaccountEvent), 1, 1), nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccount)
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)).Return(testErr)
				return gqlClient
			},
			ExpectedErrorMsg: testErr,
		},
		{
			Name:            "Error when receiving event with wrong format",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), busSubaccount.ExternalTenant).
					Return(nil, apperrors.NewNotFoundError(resource.Tenant, busSubaccount.ExternalTenant)).Once()
				return svc
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				wrongFieldMapping := tenantfetcher.TenantFieldMapping{
					IDField:   "wrong",
					NameField: tenantFieldMapping.NameField,
				}
				wrongTenantEventFields := map[string]string{
					wrongFieldMapping.IDField:   "4",
					wrongFieldMapping.NameField: "qux",
				}
				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "Subaccount", "id992", wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				return client
			},
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: errors.New("no create events for subaccount with ID subaccount-1 were found"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			beforeEach()
			persist, transact := testCase.TransactionerFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			apiClient := testCase.APIClientFn()
			gqlClient := testCase.GqlClientFn()
			svc := tenantfetcher.NewSubaccountOnDemandService(tenantfetcher.QueryConfig{
				PageNumField:    "pageNum",
				PageSizeField:   "pageSize",
				PageStartValue:  "1",
				PageSizeValue:   "1",
				SubaccountField: "entityId",
			}, tenantFieldMapping, apiClient, transact, tenantStorageSvc, gqlClient, provider, tenantConverter)

			// WHEN
			err := svc.SyncTenant(context.TODO(), subaccountID)

			// THEN
			if testCase.ExpectedErrorMsg != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, gqlClient)
		})
	}
}

func TestService_SyncAccountTenants(t *testing.T) {
	// GIVEN
	provider := "default"
	region := "eu-1"

	tenantFieldMapping := tenantfetcher.TenantFieldMapping{
		NameField:       "name",
		IDField:         "id",
		CustomerIDField: "customerId",
		SubdomainField:  "subdomain",
		EntityTypeField: "type",
	}

	parent1 := "P1"
	parent2 := "P2"
	parent3 := "P3"

	parent1GUID := fixID()
	parent2GUID := fixID()
	parent3GUID := fixID()

	parentTenant1 := fixBusinessTenantMappingInput(parent1, parent1, provider, "", "", "", tenant.Customer)
	parentTenant2 := fixBusinessTenantMappingInput(parent2, parent2, provider, "", "", "", tenant.Customer)
	parentTenant3 := fixBusinessTenantMappingInput(parent3, parent3, provider, "", "", "", tenant.Customer)
	parentTenants := []model.BusinessTenantMappingInput{parentTenant1, parentTenant2, parentTenant3}

	parentTenant1BusinessMapping := parentTenant1.ToBusinessTenantMapping(parent1GUID)
	parentTenant2BusinessMapping := parentTenant2.ToBusinessTenantMapping(parent2GUID)
	parentTenant3BusinessMapping := parentTenant3.ToBusinessTenantMapping(parent3GUID)
	parentTenantsBusinessMappingPointers := []*model.BusinessTenantMapping{parentTenant1BusinessMapping, parentTenant2BusinessMapping, parentTenant3BusinessMapping}

	busTenant1GUID := "d1f08f02-2fda-4511-962a-17fd1f1aa477"
	busTenant2GUID := "49af7161-7dc7-472b-a969-d2f0430fc41d"
	busTenant3GUID := "72409a54-2b1a-4cbb-803b-515315c74d02"

	busTenant1 := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", region, parent1, tenant.Account)
	busTenant2 := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", region, parent2, tenant.Account)
	busTenant3 := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", region, parent3, tenant.Account)
	busTenants := []model.BusinessTenantMappingInput{busTenant1, busTenant2, busTenant3}

	busTenantForDeletion1 := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", "", parent1, tenant.Account)
	busTenantForDeletion2 := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", "", parent2, tenant.Account)
	busTenantForDeletion3 := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", "", parent3, tenant.Account)
	busTenantsForDeletion := []model.BusinessTenantMappingInput{busTenantForDeletion1, busTenantForDeletion2, busTenantForDeletion3}

	busTenant1WithParentGUID := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", region, parent1GUID, tenant.Account)
	busTenant2WithParentGUID := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", region, parent2GUID, tenant.Account)
	busTenant3WithParentGUID := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", region, parent3GUID, tenant.Account)
	busTenantsWithParentGUID := []model.BusinessTenantMappingInput{busTenant1WithParentGUID, busTenant2WithParentGUID, busTenant3WithParentGUID}

	businessTenant1BusinessMapping := busTenant1.ToBusinessTenantMapping(busTenant1GUID)
	businessTenant2BusinessMapping := busTenant2.ToBusinessTenantMapping(busTenant2GUID)
	businessTenant3BusinessMapping := busTenant3.ToBusinessTenantMapping(busTenant3GUID)
	businessTenantsBusinessMappingPointers := []*model.BusinessTenantMapping{businessTenant1BusinessMapping, businessTenant2BusinessMapping, businessTenant3BusinessMapping}

	event1Fields := map[string]string{
		tenantFieldMapping.IDField:         busTenant1.ExternalTenant,
		tenantFieldMapping.NameField:       "foo",
		tenantFieldMapping.CustomerIDField: parent1,
		tenantFieldMapping.SubdomainField:  "subdomain-1",
		tenantFieldMapping.EntityTypeField: "GlobalAccount",
	}
	event2Fields := map[string]string{
		tenantFieldMapping.IDField:         busTenant2.ExternalTenant,
		tenantFieldMapping.NameField:       "bar",
		tenantFieldMapping.CustomerIDField: parent2,
		tenantFieldMapping.SubdomainField:  "subdomain-2",
		tenantFieldMapping.EntityTypeField: "GlobalAccount",
	}
	event3Fields := map[string]string{
		tenantFieldMapping.IDField:         busTenant3.ExternalTenant,
		tenantFieldMapping.NameField:       "baz",
		tenantFieldMapping.CustomerIDField: parent3,
		tenantFieldMapping.SubdomainField:  "subdomain-3",
		tenantFieldMapping.EntityTypeField: "GlobalAccount",
	}

	event1 := fixEvent(t, "GlobalAccount", busTenant1.ExternalTenant, event1Fields)
	event2 := fixEvent(t, "GlobalAccount", busTenant2.ExternalTenant, event2Fields)
	event3 := fixEvent(t, "GlobalAccount", busTenant3.ExternalTenant, event3Fields)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}
	tenantEvents := eventsToJSONArray(event1, event2, event3)

	pageOneQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
	}
	pageTwoQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "2",
		"timestamp": "1",
	}

	pageThreeQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "3",
		"timestamp": "1",
	}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	tNowInMillis := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn        func() *automock.EventAPIClient
		TenantStorageSvcFn func() *automock.TenantStorageService
		KubeClientFn       func() *automock.KubeClient
		GqlClientFn        func() *automock.DirectorGraphQLClient
		ExpectedError      error
	}{
		{
			Name:            "Success processing all kind of account events",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event3), 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:2], busTenants[:2]...)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty db and single 'create account event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants, busTenants...)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when populated db with parents and single `create account event` page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.DeletedAccountType, tenantfetcher.UpdatedAccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				preExistingTenants := []*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
				}

				preExistingTenants = append(preExistingTenants, parentTenant2BusinessMapping, parentTenant3BusinessMapping)
				svc.On("List", txtest.CtxWithDBMatcher()).Return(preExistingTenants, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsWithParentGUID))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty db and single 'update account event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedAccountType, tenantfetcher.DeletedAccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants, busTenants...)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when all tenants already exist and single 'delete account event' page is returned",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedAccountType, tenantfetcher.UpdatedAccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantsBusinessMappingPointers, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsForDeletion))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when accounts create and accounts update events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.DeletedAccountType)

				updatedEventFields := map[string]string{
					tenantFieldMapping.IDField:         busTenant1.ExternalTenant,
					tenantFieldMapping.NameField:       "updated-name",
					tenantFieldMapping.CustomerIDField: busTenant1.Parent,
					tenantFieldMapping.EntityTypeField: busTenant1.Type,
				}

				updatedTenant := fixEvent(t, busTenant1.Type, busTenant1.ExternalTenant, updatedEventFields)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{parentTenant1BusinessMapping}, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsWithParentGUID[:1]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when db is empty and both accounts create and accounts delete events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: nil,
		},
		{
			Name:            "Success when multiple pages for accounts",
			TransactionerFn: txGen.ThatSucceeds,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsWithParentGUID[1:]))).Return(nil)
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsForDeletion[0:1]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Do not update configmap if no new account events are available",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: nil,
		},
		{
			Name:            "Should perform full resync when interval elapsed",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("11218367823", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsWithParentGUID[1:]))).Return(nil)
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsForDeletion[0:1]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Should NOT perform full resync when interval is not elapsed",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", tNowInMillis, nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsWithParentGUID[1:]))).Return(nil)
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsForDeletion[0:1]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error when expected page is empty",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageThreeQueryParams).Return(nil, nil).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: errors.New("next page was expected but response was empty"),
		},
		{
			Name:            "Error when couldn't fetch page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch updated accounts event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedAccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch deleted account event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedAccountType, tenantfetcher.UpdatedAccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch next page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when results count changed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 7, 2), nil).Once()

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: errors.New("total results number changed during fetching consecutive events pages"),
		},
		{
			Name:            "Error when couldn't start transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)

				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't commit transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(parentTenantsBusinessMappingPointers[0:1], nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when tenant creation fails",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busTenants[0])

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(testErr)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't delete",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.CreatedAccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantsBusinessMappingPointers, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busTenantsForDeletion))).Return(testErr)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Skip event when receiving event with wrong format",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				wrongFieldMapping := tenantfetcher.TenantFieldMapping{
					IDField:   "wrong",
					NameField: tenantFieldMapping.NameField,
				}
				wrongTenantEventFields := map[string]string{
					wrongFieldMapping.IDField:   "4",
					wrongFieldMapping.NameField: "qux",
				}

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "GlobalAccount", "", wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)
				return client
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: UnusedGQLClient,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			kubeClient := testCase.KubeClientFn()
			gqlClient := testCase.GqlClientFn()
			tenantConverter := domainTenant.NewConverter()
			svc := tenantfetcher.NewGlobalAccountService(tenantfetcher.QueryConfig{
				PageNumField:   "pageNum",
				PageSizeField:  "pageSize",
				TimestampField: "timestamp",
				PageSizeValue:  "1",
				PageStartValue: "1",
			}, transact, kubeClient, tenantfetcher.TenantFieldMapping{
				DetailsField:       "eventData",
				DiscriminatorField: "",
				DiscriminatorValue: "",
				EventsField:        "events",
				IDField:            "id",
				GlobalAccountKey:   "gaID",
				NameField:          "name",
				CustomerIDField:    "customerId",
				SubdomainField:     "subdomain",
				TotalPagesField:    "pages",
				TotalResultsField:  "total",
				EntityTypeField:    "type",
			}, provider, region, apiClient, tenantStorageSvc, time.Hour, gqlClient, tenantInsertChunkSize,
				tenantConverter)
			svc.SetRetryAttempts(1)

			// WHEN
			err := svc.SyncTenants()

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient, gqlClient)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(nil, nil).Once()

		tenantStorageSvc := &automock.TenantStorageService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		gqlClient := &automock.DirectorGraphQLClient{}
		tenantsToCreate := []model.BusinessTenantMappingInput{fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", region, "", tenant.Subaccount)}
		gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
		defer mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient)

		svc := tenantfetcher.NewGlobalAccountService(tenantfetcher.QueryConfig{
			PageNumField:   "pageNum",
			PageSizeField:  "pageSize",
			TimestampField: "timestamp",
			PageSizeValue:  "1",
			PageStartValue: "1",
		}, transact, kubeClient, tenantfetcher.TenantFieldMapping{
			DetailsField:       "eventData",
			DiscriminatorField: "",
			DiscriminatorValue: "",
			EventsField:        "events",
			IDField:            "id",
			NameField:          "name",
			CustomerIDField:    "customerId",
			SubdomainField:     "subdomain",
			TotalPagesField:    "pages",
			TotalResultsField:  "total",
		}, provider, region, apiClient, tenantStorageSvc, time.Hour, gqlClient, tenantInsertChunkSize, tenantConverter)

		// WHEN
		err := svc.SyncTenants()

		// THEN
		require.NoError(t, err)
	})
}

func TestService_SyncSubaccountTenants(t *testing.T) {
	// GIVEN
	provider := "default"
	testRegion := "test-region"

	runtimeID := "runtime-id"

	var (
		targetInternalTenantInput model.BusinessTenantMappingInput
		targetInternalTenant      *model.BusinessTenantMapping
		sourceInternalTenantInput model.BusinessTenantMappingInput
		sourceInternalTenant      *model.BusinessTenantMapping
		tenantFieldMapping        tenantfetcher.TenantFieldMapping
		busTenant1GUID            string
		busTenant2GUID            string
		busTenant3GUID            string
		busTenant4GUID            string

		parentTenant1 model.BusinessTenantMappingInput
		parentTenant2 model.BusinessTenantMappingInput
		parentTenant3 model.BusinessTenantMappingInput
		parentTenant4 model.BusinessTenantMappingInput
		parentTenants []model.BusinessTenantMappingInput

		parentTenant1BusinessMapping *model.BusinessTenantMapping
		parentTenant2BusinessMapping *model.BusinessTenantMapping
		parentTenant3BusinessMapping *model.BusinessTenantMapping
		parentTenant4BusinessMapping *model.BusinessTenantMapping

		busSubaccount1 model.BusinessTenantMappingInput
		busSubaccount2 model.BusinessTenantMappingInput
		busSubaccount3 model.BusinessTenantMappingInput
		busSubaccount4 model.BusinessTenantMappingInput
		busSubaccounts []model.BusinessTenantMappingInput

		businessSubaccount1BusinessMapping *model.BusinessTenantMapping
		businessSubaccount2BusinessMapping *model.BusinessTenantMapping
		businessSubaccount3BusinessMapping *model.BusinessTenantMapping
		businessSubaccount4BusinessMapping *model.BusinessTenantMapping

		businessSubaccountsMappingPointers []*model.BusinessTenantMapping

		subaccountEvent1Fields map[string]string
		subaccountEvent2Fields map[string]string
		subaccountEvent3Fields map[string]string
		subaccountEvent4Fields map[string]string

		subaccountEvent1 []byte
		subaccountEvent2 []byte
		subaccountEvent3 []byte
		subaccountEvent4 []byte

		subaccountEvents []byte
	)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}

	pageOneQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
		"region":    "test-region",
	}
	pageTwoQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "2",
		"timestamp": "1",
		"region":    "test-region",
	}
	pageThreeQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "3",
		"timestamp": "1",
		"region":    "test-region",
	}

	testErr := errors.New("test error")
	notFoundErr := errors.New("Object not found")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	beforeEach := func() {
		targetInternalTenantInput = fixBusinessTenantMappingInput("target", "target", provider, "", "", "", tenant.Account)
		targetInternalTenant = targetInternalTenantInput.ToBusinessTenantMapping("internalID")

		sourceInternalTenantInput = fixBusinessTenantMappingInput("source", "source", provider, "", "", "", tenant.Account)
		sourceInternalTenant = sourceInternalTenantInput.ToBusinessTenantMapping("sourceInternalID")

		tenantFieldMapping = tenantfetcher.TenantFieldMapping{
			NameField:       "name",
			IDField:         "id",
			CustomerIDField: "customerId",
			SubdomainField:  "subdomain",
			EntityTypeField: "type",
			RegionField:     "region",
		}

		busTenant1GUID = "d1f08f02-2fda-4511-962a-17fd1f1aa477"
		busTenant2GUID = "49af7161-7dc7-472b-a969-d2f0430fc41d"
		busTenant3GUID = "72409a54-2b1a-4cbb-803b-515315c74d02"
		busTenant4GUID = "72409a54-2b1a-4cbb-803b-515315123212"

		parentTenant1 = fixBusinessTenantMappingInput(busTenant1GUID, busTenant1GUID, provider, "", "", "", tenant.Account)
		parentTenant2 = fixBusinessTenantMappingInput(busTenant2GUID, busTenant2GUID, provider, "", "", "", tenant.Account)
		parentTenant3 = fixBusinessTenantMappingInput(busTenant3GUID, busTenant3GUID, provider, "", "", "", tenant.Account)
		parentTenant4 = fixBusinessTenantMappingInput(busTenant4GUID, busTenant4GUID, provider, "", "", "", tenant.Account)
		parentTenants = []model.BusinessTenantMappingInput{parentTenant1, parentTenant2, parentTenant3, parentTenant4}

		parentTenant1BusinessMapping = parentTenant1.ToBusinessTenantMapping(busTenant1GUID)
		parentTenant2BusinessMapping = parentTenant2.ToBusinessTenantMapping(busTenant2GUID)
		parentTenant3BusinessMapping = parentTenant3.ToBusinessTenantMapping(busTenant3GUID)
		parentTenant4BusinessMapping = parentTenant4.ToBusinessTenantMapping(busTenant4GUID)

		busSubaccount1 = fixBusinessTenantMappingInput("foo", "S1", provider, "subdomain-1", "test-region", parentTenant1.ExternalTenant, tenant.Subaccount)
		busSubaccount2 = fixBusinessTenantMappingInput("bar", "S2", provider, "subdomain-2", "test-region", parentTenant2.ExternalTenant, tenant.Subaccount)
		busSubaccount3 = fixBusinessTenantMappingInput("baz", "S3", provider, "subdomain-3", "test-region", parentTenant3.ExternalTenant, tenant.Subaccount)
		busSubaccount4 = fixBusinessTenantMappingInput("bsk", "S4", provider, "subdomain-4", "test-region", parentTenant4.ExternalTenant, tenant.Subaccount)
		busSubaccounts = []model.BusinessTenantMappingInput{busSubaccount1, busSubaccount2, busSubaccount3, busSubaccount4}

		businessSubaccount1BusinessMapping = busSubaccount1.ToBusinessTenantMapping(busTenant1GUID)
		businessSubaccount2BusinessMapping = busSubaccount2.ToBusinessTenantMapping(busTenant2GUID)
		businessSubaccount3BusinessMapping = busSubaccount3.ToBusinessTenantMapping(busTenant3GUID)
		businessSubaccount4BusinessMapping = busSubaccount4.ToBusinessTenantMapping(busTenant4GUID)

		businessSubaccountsMappingPointers = []*model.BusinessTenantMapping{businessSubaccount1BusinessMapping, businessSubaccount2BusinessMapping, businessSubaccount3BusinessMapping}

		subaccountEvent1Fields = map[string]string{
			tenantFieldMapping.IDField:         "S1",
			tenantFieldMapping.NameField:       "foo",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain-1",
			tenantFieldMapping.EntityTypeField: "Subaccount",
		}
		subaccountEvent2Fields = map[string]string{
			tenantFieldMapping.IDField:         "S2",
			tenantFieldMapping.NameField:       "bar",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain-2",
			tenantFieldMapping.EntityTypeField: "Subaccount",
		}
		subaccountEvent3Fields = map[string]string{
			tenantFieldMapping.IDField:         "S3",
			tenantFieldMapping.NameField:       "baz",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain-3",
			tenantFieldMapping.EntityTypeField: "Subaccount",
		}
		subaccountEvent4Fields = map[string]string{
			tenantFieldMapping.IDField:         "S4",
			tenantFieldMapping.NameField:       "bsk",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain-4",
			tenantFieldMapping.EntityTypeField: "Subaccount",
			"source_tenant":                    "source",
			"target_tenant":                    "target",
		}

		subaccountEvent1 = fixEvent(t, "Subaccount", busTenant1GUID, subaccountEvent1Fields)
		subaccountEvent2 = fixEvent(t, "Subaccount", busTenant2GUID, subaccountEvent2Fields)
		subaccountEvent3 = fixEvent(t, "Subaccount", busTenant3GUID, subaccountEvent3Fields)
		subaccountEvent4 = fixEvent(t, "Subaccount", busTenant4GUID, subaccountEvent4Fields)

		subaccountEvents = eventsToJSONArray(subaccountEvent1, subaccountEvent2, subaccountEvent3, subaccountEvent4)
	}

	tNowInMillis := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	ctxWithSubaccountMatcher := mock.MatchedBy(func(ctx context.Context) bool {
		tenantID, err := domainTenant.LoadFromContext(ctx)
		require.NoError(t, err)
		return tenantID == businessSubaccount4BusinessMapping.ID
	})

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn         func() *automock.EventAPIClient
		TenantStorageSvcFn  func() *automock.TenantStorageService
		RuntimeStorageSvcFn func() *automock.RuntimeService
		LabelRepoFn         func() *automock.LabelRepo
		KubeClientFn        func() *automock.KubeClient
		GqlClientFn         func() *automock.DirectorGraphQLClient
		ExpectedError       error
	}{
		{
			Name:            "Success processing all kind of subaccount events",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent2), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent3), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent4), 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:2], busSubaccounts[:2]...)
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty db and single 'create subaccounts event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccounts[:1]...)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty db and single 'update subaccounts event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent2), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := []model.BusinessTenantMappingInput{parentTenant2, busSubaccount2}
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when all tenants already exist and single 'delete subaccounts event' page is returned",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.MovedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent3), 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{businessSubaccount3BusinessMapping}, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[2:3])).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when subaccounts create and subaccounts update events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)

				updatedEventFields := map[string]string{
					tenantFieldMapping.IDField:         "S1",
					tenantFieldMapping.NameField:       "updated-name",
					tenantFieldMapping.RegionField:     "test-region",
					tenantFieldMapping.SubdomainField:  "subdomain-1",
					tenantFieldMapping.EntityTypeField: "Subaccount",
				}

				updatedTenant := fixEvent(t, parentTenant1.Type, busTenant1GUID, updatedEventFields)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{parentTenant1BusinessMapping}, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				updatedTenantToCreate := []model.BusinessTenantMappingInput{fixBusinessTenantMappingInput("updated-name", "S1", "default", "subdomain-1", "test-region", busTenant1GUID, tenant.Subaccount)}
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(updatedTenantToCreate)).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when db is empty and both subaccount create and subaccount delete events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.MovedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: nil,
		},
		{
			Name:            "Success when multiple pages for subaccounts",
			TransactionerFn: txGen.ThatSucceeds,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[1:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when moving runtime and the runtime doesn't exist",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()
				return svc
			},
			LabelRepoFn: UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Should perform full resync when interval elapsed",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("11218367823", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[1:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Should NOT perform full resync when interval is not elapsed",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", tNowInMillis, nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[1:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error when can't get tenant fetcher configmap data",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				return &automock.EventAPIClient{}
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", testErr).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when expected page is empty",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageThreeQueryParams).Return(nil, nil).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: errors.New("next page was expected but response was empty"),
		},
		{
			Name:            "Error when couldn't fetch created subaccounts event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch updated subaccounts event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch deleted subaccounts event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch moved subaccounts event page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch next page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(nil, testErr).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when results count changed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 7, 2), nil).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: errors.New("total results number changed during fetching consecutive events pages"),
		},
		{
			Name:            "Error when couldn't start transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't commit transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccount1)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)).Return(nil)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when tenant creation fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccount1)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)).Return(testErr)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't delete",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.CreatedSubaccountType, tenantfetcher.MovedSubaccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessSubaccountsMappingPointers, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("DeleteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[:3]))).Return(testErr)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when moving subaccount and can't get it by external ID",
			TransactionerFn: txGen.ThatDoesntExpectCommit,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(nil, testErr).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Success when moving subaccount and can't find target parent by external ID should create it",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(nil, notFoundErr).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()
				return svc
			},
			LabelRepoFn: UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{targetInternalTenantInput}))).Return(nil)
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when moving subaccount and it is not found in our database it should be created",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(nil, notFoundErr).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				busSubaccounts[3].Parent = targetInternalTenant.ID

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[3:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error when moving subaccount and can't get target tenant by external ID",
			TransactionerFn: txGen.ThatDoesntExpectCommit,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(nil, testErr).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when moving subaccount and can't update it",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()
				return svc
			},
			LabelRepoFn: UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(testErr)
				return gqlClient
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when moving runtime and getting the runtimes fail with different error than 'NotFound'",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()
				return svc
			},
			LabelRepoFn: UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when moving runtime and converting the source GA fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(nil, testErr).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Error when moving runtime and getting scenarios label fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return(nil, testErr).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				return client
			},
			GqlClientFn:   UnusedGQLClient,
			ExpectedError: testErr,
		},
		{
			Name:            "Skip event when receiving event with wrong format",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				wrongFieldMapping := tenantfetcher.TenantFieldMapping{
					IDField:   "wrong",
					NameField: tenantFieldMapping.NameField,
				}
				wrongTenantEventFields := map[string]string{
					wrongFieldMapping.IDField:   "4",
					wrongFieldMapping.NameField: "qux",
				}

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "Subaccount", "id992", wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)
				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelRepoFn:         UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			GqlClientFn: UnusedGQLClient,
		},
		{
			Name:            "Error when can't update tenant fetcher configmap data",
			TransactionerFn: txGen.ThatSucceeds,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageTwoQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageThreeQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(subaccountEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(subaccountEvent4)+"]"), 3, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				// Moved tenants
				svc.On("GetTenantByExternalID", mock.Anything, busSubaccount4.ExternalTenant).Return(businessSubaccount4BusinessMapping, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "target").Return(targetInternalTenant, nil).Once()
				svc.On("GetTenantByExternalID", mock.Anything, "source").Return(sourceInternalTenant, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				runtimes := []*model.Runtime{
					{
						ID:   runtimeID,
						Name: "test-runtime",
					},
				}
				svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
				return svc
			},
			LabelRepoFn: func() *automock.LabelRepo {
				svc := &automock.LabelRepo{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(testErr).Once()
				return client
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				subaccountToExpect := model.BusinessTenantMappingInput{
					Name:           busSubaccount4.Name,
					ExternalTenant: busSubaccount4.ExternalTenant,
					Parent:         targetInternalTenant.ID,
					Type:           busSubaccount4.Type,
					Provider:       busSubaccount4.Provider,
				}

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("UpdateTenant", mock.Anything, busTenant4GUID, tenantConverter.ToGraphQLInput(subaccountToExpect)).Return(nil)
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[1:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			beforeEach()
			persist, transact := testCase.TransactionerFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			runtimeStorageSvc := testCase.RuntimeStorageSvcFn()
			labelSvc := testCase.LabelRepoFn()
			kubeClient := testCase.KubeClientFn()
			gqlClient := testCase.GqlClientFn()
			svc := tenantfetcher.NewSubaccountService(tenantfetcher.QueryConfig{
				PageNumField:   "pageNum",
				PageSizeField:  "pageSize",
				TimestampField: "timestamp",
				RegionField:    "region",
				PageSizeValue:  "1",
				PageStartValue: "1",
			}, transact, kubeClient, tenantfetcher.TenantFieldMapping{
				DetailsField:           "eventData",
				DiscriminatorField:     "",
				DiscriminatorValue:     "",
				EventsField:            "events",
				IDField:                "id",
				NameField:              "name",
				CustomerIDField:        "customerId",
				SubdomainField:         "subdomain",
				TotalPagesField:        "pages",
				TotalResultsField:      "total",
				EntityTypeField:        "type",
				GlobalAccountGUIDField: "globalAccountGUID",
				RegionField:            "region",
			}, tenantfetcher.MovedSubaccountsFieldMapping{
				LabelValue:   "id",
				SourceTenant: "source_tenant",
				TargetTenant: "target_tenant",
			}, provider, []string{testRegion}, apiClient, tenantStorageSvc, runtimeStorageSvc, labelSvc, time.Hour, gqlClient, tenantInsertChunkSize, tenantConverter)
			svc.SetRetryAttempts(1)

			// WHEN
			err := svc.SyncTenants()

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, runtimeStorageSvc, labelSvc, kubeClient, gqlClient)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		beforeEach()

		persist, transact := txGen.ThatSucceeds()
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()

		tenantStorageSvc := &automock.TenantStorageService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		gqlClient := &automock.DirectorGraphQLClient{}
		tenantsToCreate := append(parentTenants[:1], busSubaccounts[:1]...)
		gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)

		defer mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient)
		svc := tenantfetcher.NewSubaccountService(tenantfetcher.QueryConfig{
			PageNumField:   "pageNum",
			PageSizeField:  "pageSize",
			TimestampField: "timestamp",
			RegionField:    "region",
			PageSizeValue:  "1",
			PageStartValue: "1",
		}, transact, kubeClient, tenantfetcher.TenantFieldMapping{
			DetailsField:           "eventData",
			DiscriminatorField:     "",
			DiscriminatorValue:     "",
			EventsField:            "events",
			IDField:                "id",
			NameField:              "name",
			CustomerIDField:        "customerId",
			SubdomainField:         "subdomain",
			TotalPagesField:        "pages",
			TotalResultsField:      "total",
			EntityTypeField:        "type",
			GlobalAccountGUIDField: "globalAccountGUID",
			RegionField:            "region",
		}, tenantfetcher.MovedSubaccountsFieldMapping{
			LabelValue:   "id",
			SourceTenant: "source_tenant",
			TargetTenant: "target_tenant",
		}, provider, []string{testRegion}, apiClient, tenantStorageSvc, nil, nil, time.Hour, gqlClient, tenantInsertChunkSize, tenantConverter)

		// WHEN
		err := svc.SyncTenants()

		// THEN
		require.NoError(t, err)
	})
}

func attachNoResponseOnFirstPage(client *automock.EventAPIClient, queryParams tenantfetcher.QueryParams, eventTypes ...tenantfetcher.EventsType) {
	for _, eventType := range eventTypes {
		client.On("FetchTenantEventsPage", eventType, queryParams).Return(nil, nil).Once()
	}
}

func matchArrayWithoutOrderArgument(expected []graphql.BusinessTenantMappingInput) interface{} {
	return mock.MatchedBy(func(actual []graphql.BusinessTenantMappingInput) bool {
		if len(expected) != len(actual) {
			return false
		}
		matched := make([]bool, len(actual))
		for i := 0; i < len(matched); i++ {
			matched[i] = false
		}
		for i := 0; i < len(expected); i++ {
			for j := 0; j < len(actual); j++ {
				if assert.ObjectsAreEqual(expected[i], actual[j]) {
					matched[j] = true
				}
			}
		}
		for i := 0; i < len(matched); i++ {
			if matched[i] {
				continue
			}
			return false
		}
		return true
	})
}

func UnusedTenantStorageSvc() *automock.TenantStorageService {
	return &automock.TenantStorageService{}
}

func UnusedEventAPIClient() *automock.EventAPIClient {
	return &automock.EventAPIClient{}
}

func UnusedRuntimeStorageSvc() *automock.RuntimeService {
	return &automock.RuntimeService{}
}

func UnusedLabelSvc() *automock.LabelRepo {
	return &automock.LabelRepo{}
}

func UnusedGQLClient() *automock.DirectorGraphQLClient {
	return &automock.DirectorGraphQLClient{}
}
