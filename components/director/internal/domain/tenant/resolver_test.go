package tenant_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_Tenants(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)
	modelTenants := []*model.BusinessTenantMapping{
		newModelBusinessTenantMapping(testID, testName),
		newModelBusinessTenantMapping("test1", "name1"),
	}

	graphqlTenants := []*graphql.Tenant{
		newGraphQLTenant(testID, "", testName),
		newGraphQLTenant("test1", "", "name1"),
	}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		ExpectedOutput []*graphql.Tenant
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(modelTenants, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				TenantConv.On("MultipleToGraphQL", modelTenants).Return(graphqlTenants).Once()
				return TenantConv
			},
			ExpectedOutput: graphqlTenants,
		},
		{
			Name: "Returns error when getting tenants failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				return TenantConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when failing on begin",
			TxFn: txGen.ThatFailsOnBegin,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				return TenantConv
			},
			ExpectedError: testError,
		},
		{
			Name: "Returns error when failing on commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("List", txtest.CtxWithDBMatcher()).Return(modelTenants, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				return TenantConv
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

			// WHEN
			result, err := resolver.Tenants(ctx)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)

		})
	}

}
