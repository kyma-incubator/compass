package tenantfetcher_test

import (
	"bytes"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"testing"

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
	provider := "default"
	testRegion := "test-region"

	runtimeID := "runtime-id"

	var (
		//targetInternalTenantInput model.BusinessTenantMappingInput
		//targetInternalTenant      *model.BusinessTenantMapping
		//sourceInternalTenantInput model.BusinessTenantMappingInput
		//sourceInternalTenant      *model.BusinessTenantMapping
		tenantFieldMapping tenantfetcher.TenantFieldMapping
		busTenant1GUID     string
		//busTenant2GUID            string
		//busTenant3GUID            string
		//busTenant4GUID            string

		parentTenant1 model.BusinessTenantMappingInput
		//parentTenant2 model.BusinessTenantMappingInput
		//parentTenant3 model.BusinessTenantMappingInput
		//parentTenant4 model.BusinessTenantMappingInput
		parentTenants []model.BusinessTenantMappingInput

		parentTenant1BusinessMapping *model.BusinessTenantMapping
		//parentTenant2BusinessMapping *model.BusinessTenantMapping
		//parentTenant3BusinessMapping *model.BusinessTenantMapping
		//parentTenant4BusinessMapping *model.BusinessTenantMapping

		busSubaccount1 model.BusinessTenantMappingInput
		//busSubaccount2 model.BusinessTenantMappingInput
		//busSubaccount3 model.BusinessTenantMappingInput
		//busSubaccount4 model.BusinessTenantMappingInput
		busSubaccounts []model.BusinessTenantMappingInput

		businessSubaccount1BusinessMapping *model.BusinessTenantMapping
		//businessSubaccount2BusinessMapping *model.BusinessTenantMapping
		//businessSubaccount3BusinessMapping *model.BusinessTenantMapping
		//businessSubaccount4BusinessMapping *model.BusinessTenantMapping

		//businessSubaccountsMappingPointers []*model.BusinessTenantMapping

		subaccountEvent1Fields map[string]string
		//subaccountEvent2Fields map[string]string
		//subaccountEvent3Fields map[string]string
		//subaccountEvent4Fields map[string]string

		subaccountEvent1 []byte
		//subaccountEvent2 []byte
		//subaccountEvent3 []byte
		//subaccountEvent4 []byte

		subaccountEvents []byte
	)

	eventsToJSONArray := func(events ...[]byte) []byte {
		return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
	}

	pageOneQueryParams := tenantfetcher.QueryParams{
		"pageSize":   "1",
		"pageNum":    "1",
		"subaccount": "subaccount1",
		"region":     "test-region",
	}
	pageTwoQueryParams := tenantfetcher.QueryParams{
		"pageSize":  "1",
		"pageNum":   "2",
		"timestamp": "1",
		"region":    "test-region",
	}
	//pageThreeQueryParams := tenantfetcher.QueryParams{
	//	"pageSize":  "1",
	//	"pageNum":   "3",
	//	"timestamp": "1",
	//	"region":    "test-region",
	//}

	testErr := errors.New("test error")
	//notFoundErr := errors.New("Object not found")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	beforeEach := func() {
		//targetInternalTenantInput = fixBusinessTenantMappingInput("target", "target", provider, "", "", "", tenant.Account)
		//targetInternalTenant = targetInternalTenantInput.ToBusinessTenantMapping("internalID")

		//sourceInternalTenantInput = fixBusinessTenantMappingInput("source", "source", provider, "", "", "", tenant.Account)
		//sourceInternalTenant = sourceInternalTenantInput.ToBusinessTenantMapping("sourceInternalID")

		tenantFieldMapping = tenantfetcher.TenantFieldMapping{
			NameField:       "name",
			IDField:         "id",
			CustomerIDField: "customerId",
			SubdomainField:  "subdomain",
			EntityTypeField: "type",
			RegionField:     "region",
		}

		busTenant1GUID = "d1f08f02-2fda-4511-962a-17fd1f1aa477"
		//busTenant2GUID = "49af7161-7dc7-472b-a969-d2f0430fc41d"
		//busTenant3GUID = "72409a54-2b1a-4cbb-803b-515315c74d02"
		//busTenant4GUID = "72409a54-2b1a-4cbb-803b-515315123212"

		parentTenant1 = fixBusinessTenantMappingInput(busTenant1GUID, busTenant1GUID, provider, "", "", "", tenant.Account)
		//parentTenant2 = fixBusinessTenantMappingInput(busTenant2GUID, busTenant2GUID, provider, "", "", "", tenant.Account)
		//parentTenant3 = fixBusinessTenantMappingInput(busTenant3GUID, busTenant3GUID, provider, "", "", "", tenant.Account)
		//parentTenant4 = fixBusinessTenantMappingInput(busTenant4GUID, busTenant4GUID, provider, "", "", "", tenant.Account)
		parentTenants = []model.BusinessTenantMappingInput{parentTenant1}
		//, parentTenant2, parentTenant3, parentTenant4}

		parentTenant1BusinessMapping = parentTenant1.ToBusinessTenantMapping(busTenant1GUID)
		//parentTenant2BusinessMapping = parentTenant2.ToBusinessTenantMapping(busTenant2GUID)
		//parentTenant3BusinessMapping = parentTenant3.ToBusinessTenantMapping(busTenant3GUID)
		//parentTenant4BusinessMapping = parentTenant4.ToBusinessTenantMapping(busTenant4GUID)

		busSubaccount1 = fixBusinessTenantMappingInput("foo", "S1", provider, "subdomain-1", "test-region", parentTenant1.ExternalTenant, tenant.Subaccount)
		//busSubaccount2 = fixBusinessTenantMappingInput("bar", "S2", provider, "subdomain-2", "test-region", parentTenant2.ExternalTenant, tenant.Subaccount)
		//busSubaccount3 = fixBusinessTenantMappingInput("baz", "S3", provider, "subdomain-3", "test-region", parentTenant3.ExternalTenant, tenant.Subaccount)
		//busSubaccount4 = fixBusinessTenantMappingInput("bsk", "S4", provider, "subdomain-4", "test-region", parentTenant4.ExternalTenant, tenant.Subaccount)
		busSubaccounts = []model.BusinessTenantMappingInput{busSubaccount1}
		//, busSubaccount2, busSubaccount3, busSubaccount4}

		businessSubaccount1BusinessMapping = busSubaccount1.ToBusinessTenantMapping(busTenant1GUID)
		//businessSubaccount2BusinessMapping = busSubaccount2.ToBusinessTenantMapping(busTenant2GUID)
		//businessSubaccount3BusinessMapping = busSubaccount3.ToBusinessTenantMapping(busTenant3GUID)
		//businessSubaccount4BusinessMapping = busSubaccount4.ToBusinessTenantMapping(busTenant4GUID)

		//businessSubaccountsMappingPointers = []*model.BusinessTenantMapping{businessSubaccount1BusinessMapping}
		//, businessSubaccount2BusinessMapping, businessSubaccount3BusinessMapping}

		subaccountEvent1Fields = map[string]string{
			tenantFieldMapping.IDField:         "S1",
			tenantFieldMapping.NameField:       "foo",
			tenantFieldMapping.RegionField:     "test-region",
			tenantFieldMapping.SubdomainField:  "subdomain-1",
			tenantFieldMapping.EntityTypeField: "Subaccount",
		}
		//subaccountEvent2Fields = map[string]string{
		//	tenantFieldMapping.IDField:         "S2",
		//	tenantFieldMapping.NameField:       "bar",
		//	tenantFieldMapping.RegionField:     "test-region",
		//	tenantFieldMapping.SubdomainField:  "subdomain-2",
		//	tenantFieldMapping.EntityTypeField: "Subaccount",
		//}
		//subaccountEvent3Fields = map[string]string{
		//	tenantFieldMapping.IDField:         "S3",
		//	tenantFieldMapping.NameField:       "baz",
		//	tenantFieldMapping.RegionField:     "test-region",
		//	tenantFieldMapping.SubdomainField:  "subdomain-3",
		//	tenantFieldMapping.EntityTypeField: "Subaccount",
		//}
		//subaccountEvent4Fields = map[string]string{
		//	tenantFieldMapping.IDField:         "S4",
		//	tenantFieldMapping.NameField:       "bsk",
		//	tenantFieldMapping.RegionField:     "test-region",
		//	tenantFieldMapping.SubdomainField:  "subdomain-4",
		//	tenantFieldMapping.EntityTypeField: "Subaccount",
		//	"source_tenant":                    "source",
		//	"target_tenant":                    "target",
		//}

		subaccountEvent1 = fixEvent(t, "Subaccount", busTenant1GUID, subaccountEvent1Fields)
		//subaccountEvent2 = fixEvent(t, "Subaccount", busTenant2GUID, subaccountEvent2Fields)
		//subaccountEvent3 = fixEvent(t, "Subaccount", busTenant3GUID, subaccountEvent3Fields)
		//subaccountEvent4 = fixEvent(t, "Subaccount", busTenant4GUID, subaccountEvent4Fields)

		subaccountEvents = eventsToJSONArray(subaccountEvent1)
		//, subaccountEvent2, subaccountEvent3, subaccountEvent4)
	}

	//tNowInMillis := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	//ctxWithSubaccountMatcher := mock.MatchedBy(func(ctx context.Context) bool {
	//	tenantID, err := domainTenant.LoadFromContext(ctx)
	//	require.NoError(t, err)
	//	return tenantID == businessSubaccount4BusinessMapping.ID
	//})

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIClientFn         func() *automock.EventAPIClient
		TenantStorageSvcFn  func() *automock.TenantStorageService
		RuntimeStorageSvcFn func() *automock.RuntimeService
		LabelSvcFn          func() *automock.LabelService
		GqlClientFn         func() *automock.DirectorGraphQLClient
		ExpectedError       error
	}{
		{
			Name:            "Success processing create subaccount events",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(eventsToJSONArray(subaccountEvent1), 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			//RuntimeStorageSvcFn: func() *automock.RuntimeService {
			//	svc := &automock.RuntimeService{}
			//	runtimes := []*model.Runtime{
			//		{
			//			ID:   runtimeID,
			//			Name: "test-runtime",
			//		},
			//	}
			//	svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
			//	return svc
			//},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
				return svc
			},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:2], busSubaccounts[:2]...)
				gqlClient := &automock.DirectorGraphQLClient{}
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
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccounts[:1]...)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when tenant already exists",
			TransactionerFn: txGen.ThatSucceeds,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				attachNoResponseOnFirstPage(client, pageOneQueryParams, tenantfetcher.CreatedSubaccountType)
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{businessSubaccount1BusinessMapping}, nil).Once()
				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				return gqlClient
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

				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}

				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{
					parentTenant1BusinessMapping,
				}, nil).Once()

				return svc
			},
			//RuntimeStorageSvcFn: func() *automock.RuntimeService {
			//	svc := &automock.RuntimeService{}
			//	runtimes := []*model.Runtime{
			//		{
			//			ID:   runtimeID,
			//			Name: "test-runtime",
			//		},
			//	}
			//	svc.On("ListByFilters", ctxWithSubaccountMatcher, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()
			//	return svc
			//},
			//LabelSvcFn: func() *automock.LabelService {
			//	svc := &automock.LabelService{}
			//	svc.On("GetScenarioLabelsForRuntimes", mock.Anything, "sourceInternalID", []string{runtimeID}).Return([]model.Label{{Value: []interface{}{"DEFAULT"}}}, nil).Once()
			//	return svc
			//},
			//KubeClientFn: func() *automock.KubeClient {
			//	client := &automock.KubeClient{}
			//	client.On("GetTenantFetcherConfigMapData", mock.Anything).Return("1", "1", nil).Once()
			//	client.On("UpdateTenantFetcherConfigMapData", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			//	return client
			//},
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, matchArrayWithoutOrderArgument(tenantConverter.MultipleInputToGraphQLInput(busSubaccounts[1:]))).Return(nil)
				return gqlClient
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error when expected page is empty",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(nil, nil).Once()

				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn:         UnusedGQLClient,
			ExpectedError:       errors.New("next page was expected but response was empty"),
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
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn:         UnusedGQLClient,
			ExpectedError:       testErr,
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
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn:         UnusedGQLClient,
			ExpectedError:       testErr,
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
			LabelSvcFn:          UnusedLabelSvc,
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
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, nil).Once()

				return svc
			},
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn: func() *automock.DirectorGraphQLClient {
				tenantsToCreate := append(parentTenants[:1], busSubaccount1)

				gqlClient := &automock.DirectorGraphQLClient{}
				gqlClient.On("WriteTenants", mock.Anything, tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)).Return(testErr)
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

				wrongTenantEvents := eventsToJSONArray(fixEvent(t, "Subaccount", "id992", wrongTenantEventFields))
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedSubaccountType, pageOneQueryParams).Return(fixTenantEventsResponse(wrongTenantEvents, 1, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn:  UnusedTenantStorageSvc,
			RuntimeStorageSvcFn: UnusedRuntimeStorageSvc,
			LabelSvcFn:          UnusedLabelSvc,
			GqlClientFn:         UnusedGQLClient,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			beforeEach()
			persist, transact := testCase.TransactionerFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			runtimeStorageSvc := testCase.RuntimeStorageSvcFn()
			labelSvc := testCase.LabelSvcFn()
			gqlClient := testCase.GqlClientFn()
			svc := tenantfetcher.NewSubaccountOnDemandService([]string{testRegion}, tenantfetcher.QueryConfig{
				PageNumField:   "pageNum",
				PageSizeField:  "pageSize",
				TimestampField: "timestamp",
				RegionField:    "region",
				PageSizeValue:  "1",
				PageStartValue: "1",
			}, tenantfetcher.TenantFieldMapping{
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
			}, apiClient, transact, tenantStorageSvc, gqlClient, provider, tenantInsertChunkSize, tenantConverter)

			// WHEN
			err := svc.SyncTenant("subaccountID")

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, apiClient, tenantStorageSvc, runtimeStorageSvc, labelSvc, gqlClient)
		})
	}
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

func UnusedRuntimeStorageSvc() *automock.RuntimeService {
	return &automock.RuntimeService{}
}

func UnusedLabelSvc() *automock.LabelService {
	return &automock.LabelService{}
}

func UnusedLabelDefConverter() *automock.LabelDefConverter {
	return &automock.LabelDefConverter{}
}

func UnusedGQLClient() *automock.DirectorGraphQLClient {
	return &automock.DirectorGraphQLClient{}
}
