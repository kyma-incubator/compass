package tenantfetchersvc_test

import (
	"context"
	"testing"

	domainTenant "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/require"

	tfautomock "github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/automock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	tenantCnv = domainTenant.NewConverter()
)

func TestFetcher_FetchTenantOnDemand(t *testing.T) {
	// GIVEN
	var (
		provider       = "default"
		tenantID       = "tenantID"
		parentTenantID = "parentID"
	)
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
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *tfautomock.TenantStorageService {
				svc := &tfautomock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), businessSubaccount1BusinessMapping.ExternalTenant).Return(&businessSubaccount1BusinessMapping, nil).Once()
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

			onDemandSvc := tenantfetchersvc.NewSubaccountOnDemandService(tenantfetchersvc.QueryConfig{}, tenantfetchersvc.TenantFieldMapping{}, apiClient, transact, tenantStorageSvc, gqlClient, provider, tenantCnv)

			tf := tenantfetchersvc.NewTenantFetcher(*onDemandSvc)

			// WHEN
			err := tf.FetchTenantOnDemand(context.TODO(), tenantID, parentTenantID)

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
