package tenantfetcher_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_SyncTenants(t *testing.T) {
	// GIVEN
	provider := "default"
	region := "eu-1"

	tenantFieldMapping := tenantfetcher.TenantFieldMapping{
		NameField:       "name",
		IDField:         "id",
		CustomerIDField: "customerId",
		SubdomainField:  "subdomain",
	}

	movedRuntimeFieldMapping := tenantfetcher.MovedRuntimeByLabelFieldMapping{
		LabelValue:   "id",
		SourceTenant: "source_tenant",
		TargetTenant: "target_tenant",
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
	}
	event2Fields := map[string]string{
		tenantFieldMapping.IDField:         busTenant2.ExternalTenant,
		tenantFieldMapping.NameField:       "bar",
		tenantFieldMapping.CustomerIDField: parent2,
		tenantFieldMapping.SubdomainField:  "subdomain-2",
	}
	event3Fields := map[string]string{
		tenantFieldMapping.IDField:         busTenant3.ExternalTenant,
		tenantFieldMapping.NameField:       "baz",
		tenantFieldMapping.CustomerIDField: parent3,
		tenantFieldMapping.SubdomainField:  "subdomain-3",
	}

	event1 := fixEvent(t, event1Fields)
	event2 := fixEvent(t, event2Fields)
	event3 := fixEvent(t, event3Fields)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}
	tenantEvents := eventsToJSONArray(event1, event2, event3)

	emptyEvents := eventsToJSONArray()

	sourceExtAccID := "sourceExternalId"
	targetExtAccID := "targetExternalId"
	targetIntTenant := "targetIntTenant"
	movedRuntimeLabelKey := "moved_runtime_key"
	movedRuntimeLabelValue := "sample_runtime_label_value"

	successfullyGettingInternalTenant := func() *automock.TenantService {
		svc := &automock.TenantService{}
		svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccID).Return(targetIntTenant, nil).Once()
		return svc
	}

	movedRuntimeByLabelEvent := eventsToJSONArray(fixMovedRuntimeByLabelEvent(movedRuntimeLabelValue, sourceExtAccID, targetExtAccID, movedRuntimeFieldMapping))

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
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn         func() *automock.EventAPIClient
		TenantStorageSvcFn  func() *automock.TenantService
		RuntimeStorageSvcFn func() *automock.RuntimeService
		LabelDefSvcFn       func() *automock.LabelDefinitionService
		LabelRepoFn         func() *automock.LabelRepository
		KubeClientFn        func() *automock.KubeClient
		ExpectedError       error
	}{
		{
			Name:            "Success processing all kind of events",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event2), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event3), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				// Created tenants
				tenantsToCreate := append(parentTenants[:2], newParentsWithIDs[:2]...)
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				// Deleted tenants
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()

				// moved tenants
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccID).Return(targetIntTenant, nil).Once()

				return svc
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", txtest.CtxWithDBMatcher(), matchLabelDefinition(targetIntTenant, movedRuntimeLabelKey)).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(&model.Runtime{}, nil).Once()
				svc.On("UpdateTenantID", txtest.CtxWithDBMatcher(), mock.AnythingOfType("string"), targetIntTenant).Return(nil).Once()
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
			Name:            "Success when empty db and single 'create event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants, newParentsWithIDs...)
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

				return svc
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when populated db with parents and single `create event` page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.DeletedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
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
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs)).Return(nil)

				return svc
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty db and single 'update event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants, newParentsWithIDs...)

				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(tenantsToCreate)).Return(nil).Once()

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
			Name:            "Success when all tenants already exist and single 'delete event' page is returned",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
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
			Name:            "Successfully process 'moved runtime by label' event",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", txtest.CtxWithDBMatcher(), matchLabelDefinition(targetIntTenant, movedRuntimeLabelKey)).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(&model.Runtime{}, nil).Once()
				svc.On("UpdateTenantID", txtest.CtxWithDBMatcher(), mock.AnythingOfType("string"), targetIntTenant).Return(nil).Once()
				return svc
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccID).Return(targetIntTenant, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
		},
		{
			Name:            "Success when create and update events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				updatedEventFields := map[string]string{
					tenantFieldMapping.IDField:         busTenant1.ExternalTenant,
					tenantFieldMapping.NameField:       "updated-name",
					tenantFieldMapping.CustomerIDField: busTenant1.Parent,
				}

				updatedTenant := fixEvent(t, updatedEventFields)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{parentTenant1BusinessMapping}, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[:1])).Return(nil).Once()

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
			Name:            "Success when db is empty and both create and delete events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.AssertNotCalled(t, "UpsertManyIfNotExists", mock.Anything, mock.Anything)
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
			Name:            "Success when multiple pages",
			TransactionerFn: txGen.ThatSucceeds,

			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

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
			Name:            "Do not update configmap if no new events are available",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(nil, 0, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()

				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageThreeQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse([]byte("["+string(event1)+"]"), 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenant1BusinessMapping,
					parentTenant1BusinessMapping,
					parentTenant2BusinessMapping,
					parentTenant3BusinessMapping,
				}, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(existingParentsWithGUIDs[1:])).Return(nil).Once()

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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageThreeQueryParams).Return(nil, nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()
				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch updated events page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch deleted events page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch moved runtime by label events page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(nil, testErr).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(nil, testErr).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageTwoQueryParams).Return(fixTenantEventsResponse(tenantEvents, 7, 2), nil).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(parentTenantsBusinessMappingPointers[0:1], nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), existingParentsWithGUIDs[0:1]).Return(nil).Once()

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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				tenantsToCreate := append(parentTenants[0:1], newParentsWithIDs[0])
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), tenantsToCreate).Return(testErr).Once()

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
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.CreatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
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
			Name:            "Do nothing when 'moved runtime by label' event type is received but no runtime is found",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				err := apperrors.NewNotFoundError(resource.Label, "foo")
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(nil, err).Once()

				return svc
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error getting runtime while processing 'moved runtime by label' event",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(nil, testErr).Once()

				return svc
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
			Name:            "Error getting old internal tenant while processing 'moved runtime by label' event",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(&model.Runtime{}, nil).Once()
				return svc
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccID).Return("", testErr).Once()
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
			Name:            "Error while upserting label definition while processing 'moved runtime by label' event",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", txtest.CtxWithDBMatcher(), matchLabelDefinition(targetIntTenant, movedRuntimeLabelKey)).Return(testErr).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(&model.Runtime{}, nil).Once()
				return svc
			},
			TenantStorageSvcFn: successfullyGettingInternalTenant,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error while changing runtime tenant while processing 'moved runtime by label' event",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn: func() *automock.LabelDefinitionService {
				svc := &automock.LabelDefinitionService{}
				svc.On("Upsert", txtest.CtxWithDBMatcher(), matchLabelDefinition(targetIntTenant, movedRuntimeLabelKey)).Return(nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(&model.Runtime{}, nil).Once()
				svc.On("UpdateTenantID", txtest.CtxWithDBMatcher(), mock.AnythingOfType("string"), targetIntTenant).Return(testErr).Once()
				return svc
			},
			TenantStorageSvcFn: successfullyGettingInternalTenant,
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

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				return client
			},
			LabelDefSvcFn:       UnusedLabelDefinitionSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
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
			runtimeStorageSvc := testCase.RuntimeStorageSvcFn()
			labelDefSvc := testCase.LabelDefSvcFn()
			kubeClient := testCase.KubeClientFn()
			svc := tenantfetcher.NewService(tenantfetcher.QueryConfig{
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
			}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
				LabelValue:   "id",
				SourceTenant: "source_tenant",
				TargetTenant: "target_tenant",
			}, provider, region, apiClient, tenantStorageSvc, runtimeStorageSvc, labelDefSvc, movedRuntimeLabelKey, time.Hour)
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

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, runtimeStorageSvc, labelDefSvc, kubeClient)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(event1), 1, 1), nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

		tenantStorageSvc := &automock.TenantService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		tenantStorageSvc.On("UpsertManyIfNotExists", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		defer mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, kubeClient)

		svc := tenantfetcher.NewService(tenantfetcher.QueryConfig{
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
		}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
			LabelValue:   "id",
			SourceTenant: "source_tenant",
			TargetTenant: "target_tenant",
		}, provider, region, apiClient, tenantStorageSvc, nil, nil, movedRuntimeLabelKey, time.Hour)

		// WHEN
		err := svc.SyncTenants()

		// THEN
		require.NoError(t, err)
	})
}

func attachNoResponseOnFirstPage(client *automock.EventAPIClient, eventTypes ...tenantfetcher.EventsType) {
	queryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
	}

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

func matchBusinessTenantMappingArrayWithoutOrder(expected []model.BusinessTenantMappingInput) interface{} {
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

func match(expected []interface{}) interface{} {
	return mock.MatchedBy(func(actual []interface{}) bool {
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

func matchMovedRuntimeLabelFilter(labelKey, labelValue string) interface{} {
	return mock.MatchedBy(func(filters []*labelfilter.LabelFilter) bool {
		return len(filters) == 1 && filters[0].Key == labelKey && *filters[0].Query == fmt.Sprintf("\"%s\"", labelValue)
	})
}

func matchSubdomainLabelInput(tenantID, subdomain string) interface{} {
	return mock.MatchedBy(func(label *model.LabelInput) bool {
		return label.ObjectType == model.TenantLabelableObject && label.Key == "subdomain" &&
			label.ObjectID == tenantID && label.Value == subdomain
	})
}

func matchLabelDefinition(tenant, labelKey string) interface{} {
	return mock.MatchedBy(func(labelDef model.LabelDefinition) bool {
		return labelDef.Tenant == tenant && labelDef.Key == labelKey
	})
}

func UnusedLabelDefinitionSvc() *automock.LabelDefinitionService {
	return &automock.LabelDefinitionService{}
}

func UnusedRuntimeStorageSvc() *automock.RuntimeService {
	return &automock.RuntimeService{}
}

func UnusedTenantStorageSvc() *automock.TenantService {
	return &automock.TenantService{}
}
