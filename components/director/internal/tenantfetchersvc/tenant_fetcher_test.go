package tenantfetchersvc

import (
	"context"
	"testing"

	domainTenant "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/require"

	tfautomock "github.com/kyma-incubator/compass/components/director/internal/tenantfetcher/automock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	tenantConverter = domainTenant.NewConverter()
)

func TestFetcher_FetchTenantOnDemand(t *testing.T) {
	// GIVEN
	provider := "default"
	tenantID := "tenantID"
	businessSubaccount1BusinessMapping := model.BusinessTenantMapping{ExternalTenant: tenantID}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	// Subscribe flow
	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantStorageSvcFn func() *tfautomock.TenantStorageService
		APIClientFn        func() *tfautomock.EventAPIClient
		GqlClientFn        func() *tfautomock.DirectorGraphQLClient
		ExpectedErrorMsg   error
	}{
		{
			Name:            "Success when tenant exists",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *tfautomock.TenantStorageService {
				svc := &tfautomock.TenantStorageService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return([]*model.BusinessTenantMapping{&businessSubaccount1BusinessMapping}, nil).Once()
				return svc
			},
			APIClientFn:      UnusedEventAPIClient,
			GqlClientFn:      UnusedGQLClient,
			ExpectedErrorMsg: nil,
		},
		{
			Name:               "Error when cannot create tenant",
			TransactionerFn:    txGen.ThatFailsOnBegin,
			TenantStorageSvcFn: UnusedTenantStorageSvc,
			APIClientFn:        UnusedEventAPIClient,
			GqlClientFn:        UnusedGQLClient,
			ExpectedErrorMsg:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			apiClient := testCase.APIClientFn()
			gqlClient := testCase.GqlClientFn()

			defer mock.AssertExpectationsForObjects(t, persist, tenantStorageSvc, apiClient, gqlClient)

			onDemandSvc := tenantfetcher.NewSubaccountOnDemandService(tenantfetcher.QueryConfig{}, tenantfetcher.TenantFieldMapping{}, apiClient, transact, tenantStorageSvc, gqlClient, provider, tenantConverter)

			subscriber := NewTenantFetcher(*onDemandSvc)

			// WHEN
			err := subscriber.FetchTenantOnDemand(context.TODO(), tenantID)

			// THEN
			if testCase.ExpectedErrorMsg != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func UnusedTenantStorageSvc() *tfautomock.TenantStorageService {
	return &tfautomock.TenantStorageService{}
}

func UnusedEventAPIClient() *tfautomock.EventAPIClient {
	return &tfautomock.EventAPIClient{}
}

func UnusedGQLClient() *tfautomock.DirectorGraphQLClient {
	return &tfautomock.DirectorGraphQLClient{}
}
