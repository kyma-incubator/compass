package tenantfetcher_test

import (
	"bytes"
	"fmt"
	"testing"

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

	tenantFieldMapping := tenantfetcher.TenantFieldMapping{
		NameField: "name",
		IDField:   "id",
	}

	movedRuntimeFieldMapping := tenantfetcher.MovedRuntimeByLabelFieldMapping{
		LabelValue:   "id",
		SourceTenant: "source_tenant",
		TargetTenant: "target_tenant",
	}

	busTenant1 := fixBusinessTenantMappingInput("foo", "1", provider)
	busTenant2 := fixBusinessTenantMappingInput("bar", "2", provider)
	busTenant3 := fixBusinessTenantMappingInput("baz", "3", provider)

	businessTenants := []model.BusinessTenantMappingInput{busTenant1, busTenant2, busTenant3}

	event1 := fixEvent(busTenant1.ExternalTenant, "foo", tenantFieldMapping)
	event2 := fixEvent(busTenant2.ExternalTenant, "bar", tenantFieldMapping)
	event3 := fixEvent(busTenant3.ExternalTenant, "baz", tenantFieldMapping)

	eventsToJsonArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}
	tenantEvents := eventsToJsonArray(event1, event2, event3)

	emptyEvents := eventsToJsonArray()

	sourceExtAccId := "sourceExternalId"
	targetExtAccId := "targetExternalId"
	targetIntTenant := "targetIntTenant"
	movedRuntimeLabelKey := "moved_runtime_key"
	movedRuntimeLabelValue := "sample_runtime_label_value"

	successfullyGettingInternalTenant := func() *automock.TenantService {
		svc := &automock.TenantService{}
		svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccId).Return(targetIntTenant, nil).Once()
		return svc
	}

	movedRuntimeByLabelEvent := eventsToJsonArray(fixMovedRuntimeByLabelEvent(movedRuntimeLabelValue, sourceExtAccId, targetExtAccId, movedRuntimeFieldMapping))

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

	multiTenantEvents := append(tenantEvents, tenantEvents...)
	multiTenantEvents = append(multiTenantEvents, tenantEvents...)
	multiBusinessTenants := append(businessTenants, businessTenants...)
	multiBusinessTenants = append(multiBusinessTenants, businessTenants...)
	emptySlice := []model.BusinessTenantMappingInput{}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn         func() *automock.EventAPIClient
		TenantStorageSvcFn  func() *automock.TenantService
		RuntimeStorageSvcFn func() *automock.RuntimeService
		LabelDefSvcFn       func() *automock.LabelDefinitionService
		KubeClientFn        func() *automock.KubeClient
		ExpectedError       error
	}{
		{
			Name:            "Success processing all kind of events",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event2), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event3), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(movedRuntimeByLabelEvent, 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				// create and update event
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants[:2])).Return(nil).Once()

				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccId).Return(targetIntTenant, nil).Once()

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
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants)).Return(nil).Once()
				return svc
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when populated db and single `create event` page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.DeletedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenants[0].ToBusinessTenantMapping(fixID()),
				}, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants[1:])).Return(nil).Once()
				return svc
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants)).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when populated db and single 'delete event' page",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.CreatedEventsType, tenantfetcher.UpdatedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				businessTenantMappings := make([]*model.BusinessTenantMapping, 0, len(businessTenants))
				for _, tenant := range businessTenants {
					businessTenantMappings = append(businessTenantMappings, tenant.ToBusinessTenantMapping(fixID()))
				}

				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantMappings, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants)).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccId).Return(targetIntTenant, nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
				return client
			},
		},
		{
			Name:            "Success when create and update events refer to the same tenants",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				updatedTenant := fixEvent(busTenant1.ExternalTenant, "updated-name", tenantFieldMapping)
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(updatedTenant), 1, 1), nil).Once()

				return client
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}

				receivedTenants := []model.BusinessTenantMappingInput{{
					Name:           "updated-name",
					ExternalTenant: busTenant1.ExternalTenant,
					Provider:       busTenant1.Provider,
				}}

				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, receivedTenants)).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.AssertNotCalled(t, "CreateManyIfNotExists", mock.Anything, mock.Anything)
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), emptySlice).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					businessTenants[0].ToBusinessTenantMapping(fixID()),
				}, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants[1:])).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenants[0:1]).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.AssertNotCalled(t, "List", txtest.CtxWithDBMatcher())
				svc.AssertNotCalled(t, "CreateManyIfNotExists", txtest.CtxWithDBMatcher(), mock.Anything)
				svc.AssertNotCalled(t, "DeleteMany", txtest.CtxWithDBMatcher(), emptySlice)
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			TenantStorageSvcFn:  func() *automock.TenantService { return &automock.TenantService{} },
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			TenantStorageSvcFn:  func() *automock.TenantService { return &automock.TenantService{} },
			LabelDefSvcFn:       func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				return client
			},
			LabelDefSvcFn:       func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			TenantStorageSvcFn:  func() *automock.TenantService { return &automock.TenantService{} },
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)
				return client
			},
			LabelDefSvcFn:       func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't create",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
				attachNoResponseOnFirstPage(client, tenantfetcher.UpdatedEventsType, tenantfetcher.DeletedEventsType, tenantfetcher.MovedRuntimeByLabelEventsType)

				return client
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				svc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), mock.Anything).Return(testErr).Once()
				svc.AssertNotCalled(t, "DeleteMany")
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn:       func() *automock.LabelDefinitionService { return &automock.LabelDefinitionService{} },
			RuntimeStorageSvcFn: func() *automock.RuntimeService { return &automock.RuntimeService{} },
			TenantStorageSvcFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				businessTenantMappings := make([]*model.BusinessTenantMapping, 0, len(businessTenants))
				for _, tenant := range businessTenants {
					businessTenantMappings = append(businessTenantMappings, tenant.ToBusinessTenantMapping(fixID()))
				}

				svc.On("List", txtest.CtxWithDBMatcher()).Return(businessTenantMappings, nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), matchArrayWithoutOrderArgument(t, businessTenants)).Return(testErr).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
			LabelDefSvcFn: UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				err := apperrors.NewNotFoundError(resource.Label, "foo")
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(nil, err).Once()

				return svc
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()
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
			LabelDefSvcFn: UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByFiltersGlobal", txtest.CtxWithDBMatcher(), matchMovedRuntimeLabelFilter(movedRuntimeLabelKey, movedRuntimeLabelValue)).Return(nil, testErr).Once()

				return svc
			},
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
				svc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), targetExtAccId).Return("", testErr).Once()
				return svc
			},
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
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
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when receiving event with wrong format",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				wrongFieldMApping := tenantfetcher.TenantFieldMapping{
					IDField:   "wrong",
					NameField: tenantFieldMapping.NameField,
				}
				wrongTenantEvents := eventsToJsonArray(fixEvent("4", "qux", wrongFieldMApping))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				return client
			},
			LabelDefSvcFn:       UnusedLabelLabelSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
				client.AssertNotCalled(t, "UpdateTenantFetcherConfigMapData")
				return client
			},
			ExpectedError: errors.New("invalid format"),
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
				TotalPagesField:    "pages",
				TotalResultsField:  "total",
			}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
				LabelValue:   "id",
				SourceTenant: "source_tenant",
				TargetTenant: "target_tenant",
			}, provider, apiClient, tenantStorageSvc, runtimeStorageSvc, labelDefSvc, movedRuntimeLabelKey)
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

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			apiClient.AssertExpectations(t)
			tenantStorageSvc.AssertExpectations(t)
			runtimeStorageSvc.AssertExpectations(t)
			labelDefSvc.AssertExpectations(t)
			kubeClient.AssertExpectations(t)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJsonArray(event1), 1, 1), nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, pageOneQueryParams).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.MovedRuntimeByLabelEventsType, pageOneQueryParams).Return(fixTenantEventsResponse(emptyEvents, 0, 0), nil).Once()

		tenantStorageSvc := &automock.TenantService{}
		tenantStorageSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
		tenantStorageSvc.On("CreateManyIfNotExists", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
		kubeClient := &automock.KubeClient{}
		kubeClient.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", nil).Once()
		kubeClient.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything).Return(nil).Once()

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
			TotalPagesField:    "pages",
			TotalResultsField:  "total",
		}, tenantfetcher.MovedRuntimeByLabelFieldMapping{
			LabelValue:   "id",
			SourceTenant: "source_tenant",
			TargetTenant: "target_tenant",
		}, provider, apiClient, tenantStorageSvc, nil, nil, movedRuntimeLabelKey)

		// WHEN
		err := svc.SyncTenants()

		// THEN
		require.NoError(t, err)

		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		apiClient.AssertExpectations(t)
		tenantStorageSvc.AssertExpectations(t)
		kubeClient.AssertExpectations(t)
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

func matchArrayWithoutOrderArgument(t *testing.T, expected []model.BusinessTenantMappingInput) interface{} {
	return mock.MatchedBy(func(actual []model.BusinessTenantMappingInput) bool {
		if len(expected) != len(actual) {
			return false
		}
		return assert.ElementsMatch(t, expected, actual, "parameters do not match")
	})
}

func matchMovedRuntimeLabelFilter(labelKey, labelValue string) interface{} {
	return mock.MatchedBy(func(filters []*labelfilter.LabelFilter) bool {
		return len(filters) == 1 && filters[0].Key == labelKey && *filters[0].Query == fmt.Sprintf("\"%s\"", labelValue)
	})
}

func matchLabelDefinition(tenant, labelKey string) interface{} {
	return mock.MatchedBy(func(labelDef model.LabelDefinition) bool {
		return labelDef.Tenant == tenant && labelDef.Key == labelKey
	})
}

func UnusedLabelLabelSvc() *automock.LabelDefinitionService {
	return &automock.LabelDefinitionService{}
}

func UnusedRuntimeStorageSvc() *automock.RuntimeService {
	return &automock.RuntimeService{}
}

func UnusedTenantStorageSvc() *automock.TenantService {
	return &automock.TenantService{}
}
