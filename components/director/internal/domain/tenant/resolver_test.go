package tenant_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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

func TestResolver_Labels(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantID := "2af44425-d02d-4aed-9086-b0fc3122b508"
	testTenant := &graphql.Tenant{ID: "externalID", InternalID: tenantID}

	testLabelKey := "my-key"
	testLabels := map[string]*model.Label{
		testLabelKey: {
			ID:         "5d0ec128-47da-418a-99f5-8409105ce82d",
			Tenant:     str.Ptr(tenantID),
			Key:        testLabelKey,
			Value:      "value",
			ObjectID:   tenantID,
			ObjectType: model.TenantLabelableObject,
		},
	}

	t.Run("Succeeds", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(testLabels, nil)
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatSucceeds()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		result, err := resolver.Labels(ctx, testTenant, nil)
		assert.NoError(t, err)

		assert.NotNil(t, result)
		assert.Len(t, result, len(testLabels))
		assert.Equal(t, testLabels[testLabelKey].Value, result[testLabelKey])
	})
	t.Run("Succeeds when labels do not exist", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, testTenant.InternalID))
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatSucceeds()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		labels, err := resolver.Labels(ctx, testTenant, nil)
		assert.NoError(t, err)
		assert.Nil(t, labels)
	})
	t.Run("Returns error when the provided tenant is nil", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatDoesntStartTransaction()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		_, err := resolver.Labels(ctx, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Tenant cannot be empty")
	})
	t.Run("Returns error when starting transaction fails", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatFailsOnBegin()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		result, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
	t.Run("Returns error when it fails to list labels", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(nil, testError)
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatDoesntExpectCommit()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		_, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
	t.Run("Returns error when commit fails", func(t *testing.T) {
		tenantSvc := &automock.BusinessTenantMappingService{}
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(testLabels, nil)
		tenantConv := &automock.BusinessTenantMappingConverter{}
		persist, transact := txGen.ThatFailsOnCommit()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv)

		_, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}
