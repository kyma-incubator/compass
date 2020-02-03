package tenantfetcher_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SyncTenants(t *testing.T) {
	// GIVEN
	provider := "default"
	fieldMapping := tenantfetcher.TenantFieldMapping{
		NameField: "name",
		IDField:   "id",
	}
	tenantEvents := []tenantfetcher.Event{
		fixEvent("1", "foo", fieldMapping),
		fixEvent("2", "bar", fieldMapping),
		fixEvent("3", "baz", fieldMapping),
	}
	businessTenants := []model.BusinessTenantMappingInput{
		fixBusinessTenantMappingInput("foo", "1", provider),
		fixBusinessTenantMappingInput("bar", "2", provider),
		fixBusinessTenantMappingInput("baz", "3", provider),
	}

	multiTenantEvents := append(tenantEvents, tenantEvents...)
	multiTenantEvents = append(multiTenantEvents, tenantEvents...)
	multiBusinessTenants := append(businessTenants, businessTenants...)
	multiBusinessTenants = append(multiBusinessTenants, businessTenants...)
	var nilSlice []model.BusinessTenantMappingInput

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ConverterFn        func() *automock.Converter
		APIClientFn        func() *automock.EventAPIClient
		TenantStorageSvcFn func() *automock.TenantStorageService
		ExpectedError      error
	}{
		{
			Name:            "Success when single page",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("EventsToTenants", tenantfetcher.CreatedEventsType, tenantEvents).Return(businessTenants).Once()
				conv.On("EventsToTenants", tenantfetcher.DeletedEventsType, tenantEvents).Return(businessTenants).Once()
				conv.On("EventsToTenants", tenantfetcher.UpdatedEventsType, tenantEvents).Return(businessTenants).Once()
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), append(businessTenants, businessTenants...)).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenants).Return(nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when empty page",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("EventsToTenants", tenantfetcher.DeletedEventsType, tenantEvents).Return(businessTenants).Once()
				conv.On("EventsToTenants", tenantfetcher.UpdatedEventsType, tenantEvents).Return(businessTenants).Once()
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), businessTenants).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenants).Return(nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:            "Success when multiple pages",
			TransactionerFn: txGen.ThatSucceeds,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("EventsToTenants", tenantfetcher.CreatedEventsType, multiTenantEvents).Return(multiBusinessTenants).Once()
				conv.On("EventsToTenants", tenantfetcher.DeletedEventsType, tenantEvents).Return(businessTenants).Once()
				conv.On("EventsToTenants", tenantfetcher.UpdatedEventsType, tenantEvents).Return(businessTenants).Once()
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 2).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 3).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 3, 1), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), append(multiBusinessTenants, businessTenants...)).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), businessTenants).Return(nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:            "Error when expected page is empty",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 2).Return(fixTenantEventsResponse(tenantEvents, 9, 3), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 3).Return(nil, nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: errors.New("next page was expected but response was empty"),
		},
		{
			Name:            "Error when couldn't fetch page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch updated events page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch deleted events page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't fetch next page",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 2).Return(nil, testErr).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when results count changed",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(fixTenantEventsResponse(tenantEvents, 6, 2), nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 2).Return(fixTenantEventsResponse(tenantEvents, 7, 2), nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: errors.New("total results number changed during fetching consecutive events pages"),
		},
		{
			Name:            "Error when couldn't start transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't commit transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), nilSlice).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), nilSlice).Return(nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't create",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), nilSlice).Return(testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:            "Error when couldn't delete",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			APIClientFn: func() *automock.EventAPIClient {
				client := &automock.EventAPIClient{}
				client.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
				client.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, nil).Once()
				return client
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), nilSlice).Return(nil).Once()
				svc.On("DeleteMany", txtest.CtxWithDBMatcher(), nilSlice).Return(testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			conv := testCase.ConverterFn()
			apiClient := testCase.APIClientFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			svc := tenantfetcher.NewService(transact, conv, apiClient, tenantStorageSvc)
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
			conv.AssertExpectations(t)
			apiClient.AssertExpectations(t)
			tenantStorageSvc.AssertExpectations(t)
		})
	}

	t.Run("Success after retry", func(t *testing.T) {
		// GIVEN
		persist, transact := txGen.ThatSucceeds()
		conv := &automock.Converter{}
		apiClient := &automock.EventAPIClient{}
		apiClient.On("FetchTenantEventsPage", tenantfetcher.CreatedEventsType, 1).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.UpdatedEventsType, 1).Return(nil, nil).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, testErr).Once()
		apiClient.On("FetchTenantEventsPage", tenantfetcher.DeletedEventsType, 1).Return(nil, nil).Once()
		tenantStorageSvc := &automock.TenantStorageService{}
		tenantStorageSvc.On("Create", txtest.CtxWithDBMatcher(), nilSlice).Return(nil).Once()
		tenantStorageSvc.On("DeleteMany", txtest.CtxWithDBMatcher(), nilSlice).Return(nil).Once()

		svc := tenantfetcher.NewService(transact, conv, apiClient, tenantStorageSvc)

		// WHEN
		err := svc.SyncTenants()

		// THEN
		require.NoError(t, err)

		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		conv.AssertExpectations(t)
		apiClient.AssertExpectations(t)
		tenantStorageSvc.AssertExpectations(t)
	})
}
