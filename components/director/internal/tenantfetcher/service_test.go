package tenantfetcher_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

	busTenantForDeletion1 := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", "", parent1, tenant.Account)
	busTenantForDeletion2 := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", "", parent2, tenant.Account)
	busTenantForDeletion3 := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", "", parent3, tenant.Account)
	businessTenantsForDeletion := []model.BusinessTenantMappingInput{busTenantForDeletion1, busTenantForDeletion2, busTenantForDeletion3}

	busTenant1WithParentGUID := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", region, parent1GUID, tenant.Account)
	busTenant2WithParentGUID := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", region, parent2GUID, tenant.Account)
	busTenant3WithParentGUID := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", region, parent3GUID, tenant.Account)
	existingParentsWithGUIDs := []model.BusinessTenantMappingInput{busTenant1WithParentGUID, busTenant2WithParentGUID, busTenant3WithParentGUID}

	busTenant1WithParentID := fixBusinessTenantMappingInput("foo", "1", provider, "subdomain-1", region, parent1, tenant.Account)
	busTenant2WithParentID := fixBusinessTenantMappingInput("bar", "2", provider, "subdomain-2", region, parent2, tenant.Account)
	busTenant3WithParentID := fixBusinessTenantMappingInput("baz", "3", provider, "subdomain-3", region, parent3, tenant.Account)
	newParentsWithIDs := []model.BusinessTenantMappingInput{busTenant1WithParentID, busTenant2WithParentID, busTenant3WithParentID}

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

	event1 := fixEvent(t, "GlobalAccount", event1Fields)
	event2 := fixEvent(t, "GlobalAccount", event2Fields)
	event3 := fixEvent(t, "GlobalAccount", event3Fields)

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

	emptySlice := []model.BusinessTenantMappingInput{}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	tNowInMillis := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn        func() *automock.EventAPIClient
		TenantStorageSvcFn func() *automock.TenantService
		KubeClientFn       func() *automock.KubeClient
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				// Created tenants
				tenantsToCreate := append(parentTenants[:2], newParentsWithIDs[:2]...)
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants, newParentsWithIDs...)
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				preExistingTenants := []*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
				}

				preExistingTenants = append(preExistingTenants, parentTenant2BusinessMapping, parentTenant3BusinessMapping)
				svc.On("List", txtest.CtxWithDBMatcher()).Return(preExistingTenants, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs)).Return(nil)

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants, newParentsWithIDs...)

				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantsBusinessMappingPointers, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(businessTenantsForDeletion)).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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

				updatedTenant := fixEvent(t, busTenant1.Type, updatedEventFields)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{parentTenant1BusinessMapping}, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[:1])).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.AssertNotCalled(t, "UpsertMany", mock.Anything, mock.Anything)
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), emptySlice).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenantsForDeletion[0:1]).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()

				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenantsForDeletion[0:1]).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("11218367823", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenantsForDeletion[0:1]).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", tNowInMillis, nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(parentTenantsBusinessMappingPointers[0:1], nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), existingParentsWithGUIDs[0:1]).Return(nil).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when tenant creation fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.DeletedAccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants[0:1], newParentsWithIDs[0])
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), tenantsToCreate).Return(testErr).Once()

				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't delete",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedAccountType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedAccountType, tenantfetcher.CreatedAccountType)

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantsBusinessMappingPointers, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(businessTenantsForDeletion)).Return(testErr).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
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

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "GlobalAccount", wrongTenantEventFields))
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
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			kubeClient := testCase.KubeClientFn()
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
				EntityTypeField:    "type",
			}, provider, region, apiClient, tenantStorageSvc, time.Hour)
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

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient)
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

		tenantStorageSvc := &automock.TenantService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		tenantStorageSvc.On("UpsertMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
		}, provider, region, apiClient, tenantStorageSvc, time.Hour)

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
	movedRuntimeLabelKey := "moved_runtime_key"

	tenantFieldMapping := tenantfetcher.TenantFieldMapping{
		NameField:       "name",
		IDField:         "id",
		CustomerIDField: "customerId",
		SubdomainField:  "subdomain",
		EntityTypeField: "type",
		RegionField:     "region",
		ParentIDField:   "parentId",
	}

	busTenant1GUID := "d1f08f02-2fda-4511-962a-17fd1f1aa477"
	busTenant2GUID := "49af7161-7dc7-472b-a969-d2f0430fc41d"
	busTenant3GUID := "72409a54-2b1a-4cbb-803b-515315c74d02"
	busTenant4GUID := "72409a54-2b1a-4cbb-803b-515315123212"

	parentTenant1 := fixBusinessTenantMappingInput(busTenant1GUID, busTenant1GUID, provider, "", "", "", tenant.Account)
	parentTenant2 := fixBusinessTenantMappingInput(busTenant2GUID, busTenant2GUID, provider, "", "", "", tenant.Account)
	parentTenant3 := fixBusinessTenantMappingInput(busTenant3GUID, busTenant3GUID, provider, "", "", "", tenant.Account)
	parentTenant4 := fixBusinessTenantMappingInput(busTenant4GUID, busTenant4GUID, provider, "", "", "", tenant.Account)
	parentTenants := []model.BusinessTenantMappingInput{parentTenant1, parentTenant2, parentTenant3, parentTenant4}

	parentTenant1BusinessMapping := parentTenant1.ToBusinessTenantMapping(busTenant1GUID)
	parentTenant2BusinessMapping := parentTenant2.ToBusinessTenantMapping(busTenant2GUID)
	parentTenant3BusinessMapping := parentTenant3.ToBusinessTenantMapping(busTenant3GUID)
	parentTenant4BusinessMapping := parentTenant4.ToBusinessTenantMapping(busTenant4GUID)

	busSubaccount1 := fixBusinessTenantMappingInput("foo", "S1", provider, "subdomain-1", "test-region", parentTenant1.ExternalTenant, tenant.Subaccount)
	busSubaccount2 := fixBusinessTenantMappingInput("bar", "S2", provider, "subdomain-2", "test-region", parentTenant2.ExternalTenant, tenant.Subaccount)
	busSubaccount3 := fixBusinessTenantMappingInput("baz", "S3", provider, "subdomain-3", "test-region", parentTenant3.ExternalTenant, tenant.Subaccount)
	busSubaccount4 := fixBusinessTenantMappingInput("bsk", "S4", provider, "subdomain-4", "test-region", parentTenant4.ExternalTenant, tenant.Subaccount)
	subaccountTenants := []model.BusinessTenantMappingInput{busSubaccount1, busSubaccount2, busSubaccount3, busSubaccount4}

	businessSubaccount3BusinessMapping := busSubaccount3.ToBusinessTenantMapping(busTenant3GUID)

	subaccountEvent1Fields := map[string]string{
		tenantFieldMapping.IDField:         "S1",
		tenantFieldMapping.NameField:       "foo",
		tenantFieldMapping.ParentIDField:   busTenant1GUID,
		tenantFieldMapping.RegionField:     "test-region",
		tenantFieldMapping.SubdomainField:  "subdomain-1",
		tenantFieldMapping.EntityTypeField: "Subaccount",
	}
	subaccountEvent2Fields := map[string]string{
		tenantFieldMapping.IDField:         "S2",
		tenantFieldMapping.NameField:       "bar",
		tenantFieldMapping.ParentIDField:   busTenant2GUID,
		tenantFieldMapping.RegionField:     "test-region",
		tenantFieldMapping.SubdomainField:  "subdomain-2",
		tenantFieldMapping.EntityTypeField: "Subaccount",
	}
	subaccountEvent3Fields := map[string]string{
		tenantFieldMapping.IDField:         "S3",
		tenantFieldMapping.NameField:       "baz",
		tenantFieldMapping.ParentIDField:   busTenant3GUID,
		tenantFieldMapping.RegionField:     "test-region",
		tenantFieldMapping.SubdomainField:  "subdomain-3",
		tenantFieldMapping.EntityTypeField: "Subaccount",
	}
	subaccountEvent4Fields := map[string]string{
		tenantFieldMapping.IDField:         "S4",
		tenantFieldMapping.NameField:       "bsk",
		tenantFieldMapping.ParentIDField:   busTenant4GUID,
		tenantFieldMapping.RegionField:     "test-region",
		tenantFieldMapping.SubdomainField:  "subdomain-4",
		tenantFieldMapping.EntityTypeField: "Subaccount",
		"source_tenant":                    "source",
		"target_tenant":                    "target",
	}

	subaccountEvent1 := fixEvent(t, "Subaccount", subaccountEvent1Fields)
	subaccountEvent2 := fixEvent(t, "Subaccount", subaccountEvent2Fields)
	subaccountEvent3 := fixEvent(t, "Subaccount", subaccountEvent3Fields)
	subaccountEvent4 := fixEvent(t, "Subaccount", subaccountEvent4Fields)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}
	subaccountEvents := eventsToJSONArray(subaccountEvent1, subaccountEvent2, subaccountEvent3, subaccountEvent4)

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

	emptySlice := []model.BusinessTenantMappingInput{}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	tNowInMillis := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn         func() *automock.EventAPIClient
		TenantStorageSvcFn  func() *automock.TenantService
		RuntimeStorageSvcFn func() *automock.RuntimeService
		LabelDefSvcFn       func() *automock.LabelDefinitionService
		LabelSvcFn          func() *automock.LabelService
		KubeClientFn        func() *automock.KubeClient
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				// Created tenants
				tenantsToCreate := []model.BusinessTenantMappingInput{busSubaccount1, busSubaccount2, parentTenant1, parentTenant2}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				// Moved tenants
				svc.On("GetInternalTenant", mock.Anything, "target").Return("targetInternalTenant", nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				filters := []*labelfilter.LabelFilter{
					{
						Key:   "moved_runtime_key",
						Query: str.Ptr(fmt.Sprintf("\"%s\"", "S4")),
					},
				}
				runtime := &model.Runtime{
					ID:     "runtime-id",
					Name:   "test-runtime",
					Tenant: "runtime-tenant",
				}
				svc.On("GetByFiltersGlobal", mock.Anything, filters).Return(runtime, nil).Once()
				svc.On("UpdateTenantID", mock.Anything, "runtime-id", "targetInternalTenant").Return(nil).Once()
				return svc
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", mock.Anything, model.LabelDefinition{
					Tenant: "targetInternalTenant",
					Key:    "moved_runtime_key",
				}).Return(nil).Once()
				return svc
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", mock.Anything, "runtime-tenant", model.RuntimeLabelableObject, "runtime-id", model.ScenariosKey).Return(&model.Label{Value: []interface{}{"DEFAULT"}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				// Created tenants
				tenantsToCreate := []model.BusinessTenantMappingInput{busSubaccount1, parentTenant1}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				// Created tenants
				tenantsToCreate := []model.BusinessTenantMappingInput{busSubaccount2, parentTenant2}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{businessSubaccount3BusinessMapping}, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument([]model.BusinessTenantMappingInput{busSubaccount3})).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
					tenantFieldMapping.ParentIDField:   busTenant1GUID,
					tenantFieldMapping.RegionField:     "test-region",
					tenantFieldMapping.SubdomainField:  "subdomain-1",
					tenantFieldMapping.EntityTypeField: "Subaccount",
				}

				updatedTenant := fixEvent(t, parentTenant1.Type, updatedEventFields)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{parentTenant1BusinessMapping}, nil).Once()
				updatedTenant := fixBusinessTenantMappingInput("updated-name", "S1", provider, "subdomain-1", "test-region", parentTenant1.ExternalTenant, tenant.Subaccount)
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument([]model.BusinessTenantMappingInput{updatedTenant})).Return(nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.AssertNotCalled(t, "UpsertMany", mock.Anything, mock.Anything)
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), emptySlice).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(subaccountTenants[1:])).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				// Moved tenants
				svc.On("GetInternalTenant", mock.Anything, "target").Return("targetInternalTenant", nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				filters := []*labelfilter.LabelFilter{
					{
						Key:   "moved_runtime_key",
						Query: str.Ptr(fmt.Sprintf("\"%s\"", "S4")),
					},
				}
				runtime := &model.Runtime{
					ID:     "runtime-id",
					Name:   "test-runtime",
					Tenant: "runtime-tenant",
				}
				svc.On("GetByFiltersGlobal", mock.Anything, filters).Return(runtime, nil).Once()
				svc.On("UpdateTenantID", mock.Anything, "runtime-id", "targetInternalTenant").Return(nil).Once()
				return svc
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", mock.Anything, model.LabelDefinition{
					Tenant: "targetInternalTenant",
					Key:    "moved_runtime_key",
				}).Return(nil).Once()
				return svc
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", mock.Anything, "runtime-tenant", model.RuntimeLabelableObject, "runtime-id", model.ScenariosKey).Return(&model.Label{Value: []interface{}{"DEFAULT"}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(subaccountTenants[1:])).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				// Moved tenants
				svc.On("GetInternalTenant", mock.Anything, "target").Return("targetInternalTenant", nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				filters := []*labelfilter.LabelFilter{
					{
						Key:   "moved_runtime_key",
						Query: str.Ptr(fmt.Sprintf("\"%s\"", "S4")),
					},
				}
				runtime := &model.Runtime{
					ID:     "runtime-id",
					Name:   "test-runtime",
					Tenant: "runtime-tenant",
				}
				svc.On("GetByFiltersGlobal", mock.Anything, filters).Return(runtime, nil).Once()
				svc.On("UpdateTenantID", mock.Anything, "runtime-id", "targetInternalTenant").Return(nil).Once()
				return svc
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", mock.Anything, model.LabelDefinition{
					Tenant: "targetInternalTenant",
					Key:    "moved_runtime_key",
				}).Return(nil).Once()
				return svc
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", mock.Anything, "runtime-tenant", model.RuntimeLabelableObject, "runtime-id", model.ScenariosKey).Return(&model.Label{Value: []interface{}{"DEFAULT"}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("11218367823", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
					parentTenant4BusinessMapping,
				}, nil).Once()

				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(subaccountTenants[1:])).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				// Moved tenants
				svc.On("GetInternalTenant", mock.Anything, "target").Return("targetInternalTenant", nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				filters := []*labelfilter.LabelFilter{
					{
						Key:   "moved_runtime_key",
						Query: str.Ptr(fmt.Sprintf("\"%s\"", "S4")),
					},
				}
				runtime := &model.Runtime{
					ID:     "runtime-id",
					Name:   "test-runtime",
					Tenant: "runtime-tenant",
				}
				svc.On("GetByFiltersGlobal", mock.Anything, filters).Return(runtime, nil).Once()
				svc.On("UpdateTenantID", mock.Anything, "runtime-id", "targetInternalTenant").Return(nil).Once()
				return svc
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", mock.Anything, model.LabelDefinition{
					Tenant: "targetInternalTenant",
					Key:    "moved_runtime_key",
				}).Return(nil).Once()
				return svc
			},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", mock.Anything, "runtime-tenant", model.RuntimeLabelableObject, "runtime-id", model.ScenariosKey).Return(&model.Label{Value: []interface{}{"DEFAULT"}}, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", tNowInMillis, nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), append(parentTenants[0:1], busSubaccount1)).Return(nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants[0:1], subaccountTenants[0])
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertMany", txtest.CtxWithDBMatcher(), tenantsToCreate).Return(testErr).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
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
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(testErr).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
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

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "Subaccount", wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.UpdatedSubaccountType, tenantfetcher.DeletedSubaccountType, tenantfetcher.MovedSubaccountType)
				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefSvc,
			LabelSvcFn:          UnusedLabelSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			runtimeStorageSvc := testCase.RuntimeStorageSvcFn()
			labelDefSvc := testCase.LabelDefSvcFn()
			labelSvc := testCase.LabelSvcFn()
			kubeClient := testCase.KubeClientFn()
			svc := tenantfetcher.NewSubaccountService(tenantfetcher.QueryConfig{
				PageNumField:   "pageNum",
				PageSizeField:  "pageSize",
				TimestampField: "timestamp",
				RegionField:    "region",
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
				EntityTypeField:    "type",
				ParentIDField:      "parentId",
				RegionField:        "region",
			}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
				LabelValue:   "id",
				SourceTenant: "source_tenant",
				TargetTenant: "target_tenant",
			}, provider, []string{testRegion}, apiClient, tenantStorageSvc, runtimeStorageSvc, labelDefSvc, labelSvc, movedRuntimeLabelKey, time.Hour)
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

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, runtimeStorageSvc, labelDefSvc, labelSvc, kubeClient)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.MovedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()

		tenantStorageSvc := &automock.TenantService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		tenantStorageSvc.On("UpsertMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		defer mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient)

		svc := tenantfetcher.NewSubaccountService(tenantfetcher.QueryConfig{
			PageNumField:   "pageNum",
			PageSizeField:  "pageSize",
			TimestampField: "timestamp",
			RegionField:    "region",
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
			EntityTypeField:    "type",
			ParentIDField:      "parentId",
			RegionField:        "region",
		}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
			LabelValue:   "id",
			SourceTenant: "source_tenant",
			TargetTenant: "target_tenant",
		}, provider, []string{testRegion}, apiClient, tenantStorageSvc, nil, nil, nil, movedRuntimeLabelKey, time.Hour)

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

func matchArrayWithoutOrderArgument(expected []model.BusinessTenantMappingInput) interface{} {
	return mock.MatchedBy(func(actual []model.BusinessTenantMappingInput) bool {
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

func UnusedTenantStorageSvc() *automock.TenantService {
	return &automock.TenantService{}
}

func UnusedRuntimeStorageSvc() *automock.RuntimeService {
	return &automock.RuntimeService{}
}

func UnusedLabelDefSvc() *automock.LabelDefinitionService {
	return &automock.LabelDefinitionService{}
}

func UnusedLabelSvc() *automock.LabelService {
	return &automock.LabelService{}
}
