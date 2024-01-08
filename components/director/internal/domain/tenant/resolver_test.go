package tenant_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

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

	first := 2
	gqlAfter := graphql.PageCursor("test")
	searchTerm := ""
	testFirstParameterMissingError := errors.New("Invalid data [reason=missing required parameter 'first']")

	modelTenants := []*model.BusinessTenantMapping{
		newModelBusinessTenantMapping(testID, testName),
		newModelBusinessTenantMapping("test1", "name1"),
	}

	modelTenantsPage := &model.BusinessTenantMappingPage{
		Data: modelTenants,
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   string(gqlAfter),
			HasNextPage: true,
		},
		TotalCount: 3,
	}

	gqlTenants := []*graphql.Tenant{
		newGraphQLTenant(testID, "", testName),
		newGraphQLTenant("test1", "", "name1"),
	}

	gqlTenantsPage := &graphql.TenantPage{
		Data:       gqlTenants,
		TotalCount: modelTenantsPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(modelTenantsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(modelTenantsPage.PageInfo.EndCursor),
			HasNextPage: modelTenantsPage.PageInfo.HasNextPage,
		},
	}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		first          *int
		ExpectedOutput *graphql.TenantPage
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("ListPageBySearchTerm", txtest.CtxWithDBMatcher(), searchTerm, first, string(gqlAfter)).Return(modelTenantsPage, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				TenantConv.On("MultipleToGraphQL", modelTenants).Return(gqlTenants).Once()
				return TenantConv
			},
			first:          &first,
			ExpectedOutput: gqlTenantsPage,
		},
		{
			Name: "Returns error when getting tenants failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("ListPageBySearchTerm", txtest.CtxWithDBMatcher(), searchTerm, first, string(gqlAfter)).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn:  unusedTenantConverter,
			first:         &first,
			ExpectedError: testError,
		},
		{
			Name:          "Returns error when failing on begin",
			TxFn:          txGen.ThatFailsOnBegin,
			TenantSvcFn:   unusedTenantService,
			TenantConvFn:  unusedTenantConverter,
			first:         &first,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when failing on commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("ListPageBySearchTerm", txtest.CtxWithDBMatcher(), searchTerm, first, string(gqlAfter)).Return(modelTenantsPage, nil).Once()
				return TenantSvc
			},
			TenantConvFn:  unusedTenantConverter,
			first:         &first,
			ExpectedError: testError,
		},
		{
			Name: "Returns error when 'first' parameter is missing",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.AssertNotCalled(t, "ListPageBySearchTerm")
				return TenantSvc
			},
			TenantConvFn:  unusedTenantConverter,
			first:         nil,
			ExpectedError: testFirstParameterMissingError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.Tenants(ctx, testCase.first, &gqlAfter, &searchTerm)

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

func TestResolver_Tenant(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantNotFoundError := apperrors.NewNotFoundError(resource.Tenant, testExternal)

	testCases := []struct {
		Name           string
		TxFn           func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		TenantFetcher  func() *automock.Fetcher
		TenantInput    graphql.BusinessTenantMappingInput
		IDInput        string
		ExpectedError  error
		ExpectedResult *graphql.Tenant
	}{
		{
			Name: "Success",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistenceTx, transact := txGen.ThatSucceeds()
				return []*persistenceautomock.PersistenceTx{persistenceTx, {}}, transact
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("ToGraphQL", expectedTenantModel).Return(expectedTenantGQL)
				return conv
			},
			TenantFetcher:  unusedFetcherService,
			IDInput:        testExternal,
			ExpectedError:  nil,
			ExpectedResult: expectedTenantGQL,
		},
		{
			Name: "Success when tenant has to be fetched",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				secondPersistTx := &persistenceautomock.PersistenceTx{}
				secondPersistTx.On("Commit").Return(nil).Once()
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(secondPersistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return []*persistenceautomock.PersistenceTx{persistTx, secondPersistTx}, transact
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(nil, tenantNotFoundError).Once()
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("ToGraphQL", expectedTenantModel).Return(expectedTenantGQL)
				return conv
			},
			TenantFetcher: func() *automock.Fetcher {
				fetcher := &automock.Fetcher{}
				fetcher.On("FetchOnDemand", txtest.CtxWithDBMatcher(), testExternal, "").Return(nil)
				return fetcher
			},
			IDInput:        testExternal,
			ExpectedError:  nil,
			ExpectedResult: expectedTenantGQL,
		},
		{
			Name: "That returns error when can not start transaction",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistenceTx, transact := txGen.ThatFailsOnBegin()
				return []*persistenceautomock.PersistenceTx{persistenceTx, {}}, transact
			},
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			TenantFetcher:  unusedFetcherService,
			IDInput:        testExternal,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "That returns error when can not get tenant by external ID",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return []*persistenceautomock.PersistenceTx{persistTx, {}}, transact
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(nil, tenantNotFoundError).Once()
				return TenantSvc
			},
			TenantConvFn: unusedTenantConverter,
			TenantFetcher: func() *automock.Fetcher {
				fetcher := &automock.Fetcher{}
				fetcher.On("FetchOnDemand", txtest.CtxWithDBMatcher(), testExternal, "").Return(tenantNotFoundError)
				return fetcher
			},
			IDInput:        testExternal,
			ExpectedError:  tenantNotFoundError,
			ExpectedResult: expectedTenantGQL,
		},
		{
			Name: "That returns error when can not fetch tenant",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistenceTx, transact := txGen.ThatDoesntExpectCommit()
				return []*persistenceautomock.PersistenceTx{persistenceTx, {}}, transact
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			TenantFetcher:  unusedFetcherService,
			IDInput:        testExternal,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "That returns error when cannot commit",
			TxFn: func() ([]*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistenceTx, transact := txGen.ThatFailsOnCommit()
				return []*persistenceautomock.PersistenceTx{persistenceTx, {}}, transact
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), testExternal).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			TenantFetcher:  unusedFetcherService,
			IDInput:        testExternal,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			fetcherSvc := testCase.TenantFetcher()
			persistencesTx, transact := testCase.TxFn()

			defer mock.AssertExpectationsForObjects(t, transact, tenantSvc, tenantConv, fetcherSvc, persistencesTx[0], persistencesTx[1])

			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, fetcherSvc)

			// WHEN
			result, err := resolver.Tenant(ctx, testCase.IDInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}
		})
	}
}

func TestResolver_TenantByID(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantParent := ""
	tenantInternalID := "internal"

	expectedTenantModel := &model.BusinessTenantMapping{
		ID:             testInternal,
		Name:           testName,
		ExternalTenant: testInternal,
		Parent:         tenantParent,
		Type:           tnt.Account,
		Provider:       testProvider,
		Status:         tnt.Active,
		Initialized:    nil,
	}

	expectedTenantGQL := &graphql.Tenant{
		ID:          testInternal,
		InternalID:  tenantInternalID,
		Name:        str.Ptr(testName),
		Type:        string(tnt.Account),
		ParentID:    tenantParent,
		Initialized: nil,
		Labels:      nil,
	}

	testCases := []struct {
		Name            string
		TxFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn     func() *automock.BusinessTenantMappingService
		TenantConvFn    func() *automock.BusinessTenantMappingConverter
		InternalIDInput string
		ExpectedError   error
		ExpectedResult  *graphql.Tenant
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), testInternal).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("ToGraphQL", expectedTenantModel).Return(expectedTenantGQL)
				return conv
			},
			InternalIDInput: testInternal,
			ExpectedError:   nil,
			ExpectedResult:  expectedTenantGQL,
		},
		{
			Name:            "That returns error when can not start transaction",
			TxFn:            txGen.ThatFailsOnBegin,
			TenantSvcFn:     unusedTenantService,
			TenantConvFn:    unusedTenantConverter,
			InternalIDInput: testInternal,
			ExpectedError:   testError,
			ExpectedResult:  nil,
		},
		{
			Name: "That returns error when can not get tenant by internal ID",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), testInternal).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn:    unusedTenantConverter,
			InternalIDInput: testInternal,
			ExpectedError:   testError,
			ExpectedResult:  nil,
		},
		{
			Name: "That returns error when cannot commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), testInternal).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			TenantConvFn:    unusedTenantConverter,
			InternalIDInput: testInternal,
			ExpectedError:   testError,
			ExpectedResult:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.TenantByID(ctx, testCase.InternalIDInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_TenantByLowestOwnerForResource(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	resourceTypeStr := "application"
	resourceType := resource.Type(resourceTypeStr)
	objectID := "objectID"

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		ExpectedError  error
		ExpectedResult string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resourceType, objectID).Return(testInternal, nil).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			ExpectedError:  nil,
			ExpectedResult: testInternal,
		},
		{
			Name:           "That returns error when can not start transaction",
			TxFn:           txGen.ThatFailsOnBegin,
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name: "That returns error when can not get lowest owner for resource",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resourceType, objectID).Return("", testError).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name: "That returns error when cannot commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resourceType, objectID).Return(testInternal, nil).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			ExpectedError:  testError,
			ExpectedResult: testInternal,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			tenantID, err := resolver.TenantByLowestOwnerForResource(ctx, resourceTypeStr, objectID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, tenantID)
			}

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
		tenantSvc := unusedTenantService()
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(testLabels, nil)
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatSucceeds()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		result, err := resolver.Labels(ctx, testTenant, nil)
		assert.NoError(t, err)

		assert.NotNil(t, result)
		assert.Len(t, result, len(testLabels))
		assert.Equal(t, testLabels[testLabelKey].Value, result[testLabelKey])
	})
	t.Run("Succeeds when labels do not exist", func(t *testing.T) {
		tenantSvc := unusedTenantService()
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, testTenant.InternalID))
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatSucceeds()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		labels, err := resolver.Labels(ctx, testTenant, nil)
		assert.NoError(t, err)
		assert.Nil(t, labels)
	})
	t.Run("Returns error when the provided tenant is nil", func(t *testing.T) {
		tenantSvc := unusedTenantService()
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatDoesntStartTransaction()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		_, err := resolver.Labels(ctx, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Tenant cannot be empty")
	})
	t.Run("Returns error when starting transaction fails", func(t *testing.T) {
		tenantSvc := unusedTenantService()
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatFailsOnBegin()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		result, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
	t.Run("Returns error when it fails to list labels", func(t *testing.T) {
		tenantSvc := unusedTenantService()
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(nil, testError)
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatDoesntExpectCommit()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		_, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
	t.Run("Returns error when commit fails", func(t *testing.T) {
		tenantSvc := unusedTenantService()
		tenantSvc.On("ListLabels", txtest.CtxWithDBMatcher(), testTenant.InternalID).Return(testLabels, nil)
		tenantConv := unusedTenantConverter()
		persist, transact := txGen.ThatFailsOnCommit()

		defer mock.AssertExpectationsForObjects(t, tenantSvc, tenantConv, persist, transact)

		resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

		_, err := resolver.Labels(ctx, testTenant, nil)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}

func TestResolver_Write(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantNames := []string{"name1", "name2"}
	tenantExternalTenants := []string{"external1", "external2"}
	tenantParent := ""
	tenantSubdomain := "subdomain"
	tenantRegion := "region"
	tenantProvider := "test"

	tenantsToUpsertGQL := []*graphql.BusinessTenantMappingInput{
		{
			Name:           tenantNames[0],
			ExternalTenant: tenantExternalTenants[0],
			Parent:         str.Ptr(tenantParent),
			Subdomain:      str.Ptr(tenantSubdomain),
			Region:         str.Ptr(tenantRegion),
			Type:           string(tnt.Account),
			Provider:       tenantProvider,
		},
		{
			Name:           tenantNames[1],
			ExternalTenant: tenantExternalTenants[1],
			Parent:         str.Ptr(tenantParent),
			Subdomain:      str.Ptr(tenantSubdomain),
			Region:         str.Ptr(tenantRegion),
			Type:           string(tnt.Account),
			Provider:       tenantProvider,
		},
	}
	tenantsToUpsertModel := []model.BusinessTenantMappingInput{
		{
			Name:           tenantNames[0],
			ExternalTenant: tenantExternalTenants[0],
			Parent:         tenantParent,
			Subdomain:      tenantSubdomain,
			Region:         tenantRegion,
			Type:           string(tnt.Account),
			Provider:       tenantProvider,
		},
		{
			Name:           tenantNames[1],
			ExternalTenant: tenantExternalTenants[1],
			Parent:         tenantParent,
			Subdomain:      tenantSubdomain,
			Region:         tenantRegion,
			Type:           string(tnt.Account),
			Provider:       tenantProvider,
		},
	}

	upsertedTenantsIDs := []string{"6f4a589c-ac2e-4870-acb5-5a58abc85c6a", "eace9a3a-383b-44d1-8864-7e951ac5ec06"}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		TenantsInput   []*graphql.BusinessTenantMappingInput
		ExpectedError  error
		ExpectedResult []string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := unusedTenantService()
				TenantSvc.On("UpsertMany", txtest.CtxWithDBMatcher(), tenantsToUpsertModel[0], tenantsToUpsertModel[1]).Return(upsertedTenantsIDs, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				TenantConv.On("MultipleInputFromGraphQL", tenantsToUpsertGQL).Return(tenantsToUpsertModel).Once()
				return TenantConv
			},
			TenantsInput:   tenantsToUpsertGQL,
			ExpectedError:  nil,
			ExpectedResult: upsertedTenantsIDs,
		},
		{
			Name:           "Returns error when can not start transaction",
			TxFn:           txGen.ThatFailsOnBegin,
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantsToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "Returns error when can not create the tenants",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("UpsertMany", txtest.CtxWithDBMatcher(), tenantsToUpsertModel[0], tenantsToUpsertModel[1]).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				TenantConv.On("MultipleInputFromGraphQL", tenantsToUpsertGQL).Return(tenantsToUpsertModel).Once()
				return TenantConv
			},
			TenantsInput:   tenantsToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "Returns error when can not commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("UpsertMany", txtest.CtxWithDBMatcher(), tenantsToUpsertModel[0], tenantsToUpsertModel[1]).Return(upsertedTenantsIDs, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				TenantConv := &automock.BusinessTenantMappingConverter{}
				TenantConv.On("MultipleInputFromGraphQL", tenantsToUpsertGQL).Return(tenantsToUpsertModel).Once()
				return TenantConv
			},
			TenantsInput:   tenantsToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.Write(ctx, testCase.TenantsInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_WriteSingle(t *testing.T) {
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantName := "name1"
	tenantExternalTenant := "external1"
	tenantParent := ""
	tenantSubdomain := "subdomain"
	tenantRegion := "region"
	tenantProvider := "test"
	tenantID := "2af44425-d02d-4aed-9086-b0fc3122b508"

	tenantToUpsertGQL := graphql.BusinessTenantMappingInput{
		Name:           tenantName,
		ExternalTenant: tenantExternalTenant,
		Parent:         str.Ptr(tenantParent),
		Subdomain:      str.Ptr(tenantSubdomain),
		Region:         str.Ptr(tenantRegion),
		Type:           string(tnt.Account),
		Provider:       tenantProvider,
	}
	tenantToUpsertModel := model.BusinessTenantMappingInput{
		Name:           tenantName,
		ExternalTenant: tenantExternalTenant,
		Parent:         tenantParent,
		Subdomain:      tenantSubdomain,
		Region:         tenantRegion,
		Type:           string(tnt.Account),
		Provider:       tenantProvider,
	}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		TenantsInput   graphql.BusinessTenantMappingInput
		ExpectedError  error
		ExpectedResult string
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantSvc := unusedTenantService()
				tenantSvc.On("UpsertSingle", txtest.CtxWithDBMatcher(), tenantToUpsertModel).Return(tenantID, nil).Once()
				return tenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				tenantConv := &automock.BusinessTenantMappingConverter{}
				tenantConv.On("InputFromGraphQL", tenantToUpsertGQL).Return(tenantToUpsertModel).Once()
				return tenantConv
			},
			TenantsInput:   tenantToUpsertGQL,
			ExpectedError:  nil,
			ExpectedResult: tenantID,
		},
		{
			Name:           "Returns error when can not start transaction",
			TxFn:           txGen.ThatFailsOnBegin,
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name: "Error when upserting",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantSvc := unusedTenantService()
				tenantSvc.On("UpsertSingle", txtest.CtxWithDBMatcher(), tenantToUpsertModel).Return("", testError).Once()
				return tenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				tenantConv := &automock.BusinessTenantMappingConverter{}
				tenantConv.On("InputFromGraphQL", tenantToUpsertGQL).Return(tenantToUpsertModel).Once()
				return tenantConv
			},
			TenantsInput:   tenantToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name: "Returns error when fails to commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				tenantSvc := &automock.BusinessTenantMappingService{}
				tenantSvc.On("UpsertSingle", txtest.CtxWithDBMatcher(), tenantToUpsertModel).Return(tenantID, nil).Once()
				return tenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				tenantConv := &automock.BusinessTenantMappingConverter{}
				tenantConv.On("InputFromGraphQL", tenantToUpsertGQL).Return(tenantToUpsertModel).Once()
				return tenantConv
			},
			TenantsInput:   tenantToUpsertGQL,
			ExpectedError:  testError,
			ExpectedResult: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.WriteSingle(ctx, testCase.TenantsInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_Delete(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantExternalTenants := []string{"external1", "external2"}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		TenantsInput   []string
		ExpectedError  error
		ExpectedResult int
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("DeleteMany", txtest.CtxWithDBMatcher(), tenantExternalTenants).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantExternalTenants,
			ExpectedError:  nil,
			ExpectedResult: 2,
		},
		{
			Name:           "Returns error when can not start transaction",
			TxFn:           txGen.ThatFailsOnBegin,
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantExternalTenants,
			ExpectedError:  testError,
			ExpectedResult: -1,
		},
		{
			Name: "Returns error when can not create the tenants",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("DeleteMany", txtest.CtxWithDBMatcher(), tenantExternalTenants).Return(testError).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantExternalTenants,
			ExpectedError:  testError,
			ExpectedResult: -1,
		},
		{
			Name: "Returns error when can not commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("DeleteMany", txtest.CtxWithDBMatcher(), tenantExternalTenants).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn:   unusedTenantConverter,
			TenantsInput:   tenantExternalTenants,
			ExpectedError:  testError,
			ExpectedResult: -1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.Delete(ctx, testCase.TenantsInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_Update(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	tenantParent := ""
	tenantInternalID := "internal"

	tenantsToUpdateGQL := []*graphql.BusinessTenantMappingInput{
		{
			Name:           testName,
			ExternalTenant: testExternal,
			Parent:         str.Ptr(tenantParent),
			Subdomain:      str.Ptr(testSubdomain),
			Region:         str.Ptr(testRegion),
			Type:           string(tnt.Account),
			Provider:       testProvider,
		},
	}

	tenantsToUpdateModel := []model.BusinessTenantMappingInput{
		{
			Name:           testName,
			ExternalTenant: testExternal,
			Parent:         tenantParent,
			Subdomain:      testSubdomain,
			Region:         testRegion,
			Type:           string(tnt.Account),
			Provider:       testProvider,
		},
	}

	expectedTenantModel := &model.BusinessTenantMapping{
		ID:             testExternal,
		Name:           testName,
		ExternalTenant: testExternal,
		Parent:         tenantParent,
		Type:           tnt.Account,
		Provider:       testProvider,
		Status:         tnt.Active,
		Initialized:    nil,
	}

	expectedTenantGQL := &graphql.Tenant{
		ID:          testExternal,
		InternalID:  tenantInternalID,
		Name:        str.Ptr(testName),
		Type:        string(tnt.Account),
		ParentID:    tenantParent,
		Initialized: nil,
		Labels:      nil,
	}

	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn    func() *automock.BusinessTenantMappingService
		TenantConvFn   func() *automock.BusinessTenantMappingConverter
		TenantInput    graphql.BusinessTenantMappingInput
		IDInput        string
		ExpectedError  error
		ExpectedResult *graphql.Tenant
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), tenantsToUpdateGQL[0].ExternalTenant).Return(expectedTenantModel, nil).Once()
				TenantSvc.On("Update", txtest.CtxWithDBMatcher(), tenantInternalID, tenantsToUpdateModel[0]).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("MultipleInputFromGraphQL", tenantsToUpdateGQL).Return(tenantsToUpdateModel)
				conv.On("ToGraphQL", expectedTenantModel).Return(expectedTenantGQL)
				return conv
			},
			TenantInput:    *tenantsToUpdateGQL[0],
			IDInput:        tenantInternalID,
			ExpectedError:  nil,
			ExpectedResult: expectedTenantGQL,
		},
		{
			Name:           "Returns error when can not start transaction",
			TxFn:           txGen.ThatFailsOnBegin,
			TenantSvcFn:    unusedTenantService,
			TenantConvFn:   unusedTenantConverter,
			TenantInput:    *tenantsToUpdateGQL[0],
			IDInput:        tenantInternalID,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "Returns error when updating tenant fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("Update", txtest.CtxWithDBMatcher(), tenantInternalID, tenantsToUpdateModel[0]).Return(testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("MultipleInputFromGraphQL", tenantsToUpdateGQL).Return(tenantsToUpdateModel)
				return conv
			},
			TenantInput:    *tenantsToUpdateGQL[0],
			IDInput:        tenantInternalID,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "Returns error when can not get tenant by external ID",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), tenantsToUpdateGQL[0].ExternalTenant).Return(nil, testError).Once()
				TenantSvc.On("Update", txtest.CtxWithDBMatcher(), tenantInternalID, tenantsToUpdateModel[0]).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("MultipleInputFromGraphQL", tenantsToUpdateGQL).Return(tenantsToUpdateModel)
				return conv
			},
			TenantInput:    *tenantsToUpdateGQL[0],
			IDInput:        tenantInternalID,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
		{
			Name: "Returns error when can not commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), tenantsToUpdateGQL[0].ExternalTenant).Return(expectedTenantModel, nil).Once()
				TenantSvc.On("Update", txtest.CtxWithDBMatcher(), tenantInternalID, tenantsToUpdateModel[0]).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("MultipleInputFromGraphQL", tenantsToUpdateGQL).Return(tenantsToUpdateModel)
				return conv
			},
			TenantInput:    *tenantsToUpdateGQL[0],
			IDInput:        tenantInternalID,
			ExpectedError:  testError,
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvcFn()
			tenantConv := testCase.TenantConvFn()
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.Update(ctx, testCase.IDInput, testCase.TenantInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_AddTenantAccess(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name             string
		TxFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn      func() *automock.BusinessTenantMappingService
		TenantConvFn     func() *automock.BusinessTenantMappingConverter
		Input            graphql.TenantAccessInput
		ExpectedErrorMsg string
		ExpectedResult   *graphql.TenantAccess
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("CreateTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(tenantAccessGQL, nil).Once()
				return conv
			},
			Input:          tenantAccessInput,
			ExpectedResult: tenantAccessGQL,
		},
		{
			Name: "Error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("CreateTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(tenantAccessGQL, nil).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name: "Error when converting to graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("CreateTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(nil, testError).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while converting to graphql tenant access",
		},
		{
			Name: "Error when getting tenant access",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("CreateTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(nil, testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while fetching stored tenant access for tenant",
		},
		{
			Name: "Error when creating tenant access record",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("CreateTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while creating tenant access record",
		},
		{
			Name: "Error when getting internal tenant",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return("", testError).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(tenantAccessWithoutInternalTenantModel, nil).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while getting internal tenant",
		},
		{
			Name: "Error when converting input",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessInputFromGraphQL", tenantAccessInput).Return(nil, testError).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while converting tenant access inpu",
		},
		{
			Name:             "Error when converting input",
			TxFn:             txGen.ThatFailsOnBegin,
			Input:            tenantAccessInput,
			ExpectedErrorMsg: testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := unusedTenantService()
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			tenantConv := unusedTenantConverter()
			if testCase.TenantConvFn != nil {
				tenantConv = testCase.TenantConvFn()
			}
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.AddTenantAccess(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_RemoveTenantAccess(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	testCases := []struct {
		Name             string
		TxFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantSvcFn      func() *automock.BusinessTenantMappingService
		TenantConvFn     func() *automock.BusinessTenantMappingConverter
		Input            graphql.TenantAccessInput
		ExpectedErrorMsg string
		ExpectedResult   *graphql.TenantAccess
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				TenantSvc.On("DeleteTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(tenantAccessGQL, nil).Once()
				return conv
			},
			Input:          tenantAccessInput,
			ExpectedResult: tenantAccessGQL,
		},
		{
			Name: "Error when committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				TenantSvc.On("DeleteTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(tenantAccessGQL, nil).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name: "Error when converting to graphql",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				TenantSvc.On("DeleteTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(nil).Once()
				return TenantSvc
			},
			TenantConvFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToGraphQL", tenantAccessModel).Return(nil, testError).Once()
				return conv
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while converting to graphql tenant access",
		},
		{
			Name: "Error when deleting tenant access record",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(tenantAccessModelWithoutExternalTenant, nil).Once()
				TenantSvc.On("DeleteTenantAccessForResourceRecursively", txtest.CtxWithDBMatcher(), tenantAccessModel).Return(testError).Once()
				return TenantSvc
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while deleting tenant access record",
		},
		{
			Name: "Error when getting tenant access record",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(nil, testError).Once()
				return TenantSvc
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while fetching stored tenant access for tenant",
		},
		{
			Name: "Error not found when getting tenant access record",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				TenantSvc.On("GetTenantAccessForResource", txtest.CtxWithDBMatcher(), testInternal, testID, resource.Application).Return(nil, apperrors.NewNotFoundErrorWithType(resource.TenantAccess)).Once()
				return TenantSvc
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "Object not found [object=tenantAccess]",
		},
		{
			Name: "Error when resource type is invalid",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return(testInternal, nil).Once()
				return TenantSvc
			},
			Input:            tenantAccessInputWithInvalidResourceType,
			ExpectedErrorMsg: fmt.Sprintf("Unknown tenant access resource type %q", invalidResourceType),
		},
		{
			Name: "Error when getting internal tenant",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), testExternal).Return("", testError).Once()
				return TenantSvc
			},
			Input:            tenantAccessInput,
			ExpectedErrorMsg: "while getting internal tenant",
		},
		{
			Name:             "Error when converting input",
			TxFn:             txGen.ThatFailsOnBegin,
			Input:            tenantAccessInput,
			ExpectedErrorMsg: testError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := unusedTenantService()
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			tenantConv := unusedTenantConverter()
			if testCase.TenantConvFn != nil {
				tenantConv = testCase.TenantConvFn()
			}
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConv, nil)

			// WHEN
			result, err := resolver.RemoveTenantAccess(ctx, testCase.Input.TenantID, testCase.Input.ResourceID, testCase.Input.ResourceType)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConv)
		})
	}
}

func TestResolver_RootTenant(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testError)

	externalTenant := "external-tenant"

	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantConverterFn func() *automock.BusinessTenantMappingConverter
		TenantSvcFn       func() *automock.BusinessTenantMappingService
		ExpectedError     error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantConverterFn: func() *automock.BusinessTenantMappingConverter {
				converter := &automock.BusinessTenantMappingConverter{}
				converter.On("ToGraphQL", expectedTenantModel).Return(expectedTenantGQL).Once()
				return converter
			},
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetParentRecursivelyByExternalTenant", txtest.CtxWithDBMatcher(), externalTenant).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
		},
		{
			Name:          "That returns error when can not start transaction",
			TxFn:          txGen.ThatFailsOnBegin,
			ExpectedError: testError,
		},
		{
			Name: "That returns error when can not get parent by external tenant",
			TxFn: txGen.ThatDoesntExpectCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetParentRecursivelyByExternalTenant", txtest.CtxWithDBMatcher(), externalTenant).Return(nil, testError).Once()
				return TenantSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "That returns error when cannot commit",
			TxFn: txGen.ThatFailsOnCommit,
			TenantSvcFn: func() *automock.BusinessTenantMappingService {
				TenantSvc := &automock.BusinessTenantMappingService{}
				TenantSvc.On("GetParentRecursivelyByExternalTenant", txtest.CtxWithDBMatcher(), externalTenant).Return(expectedTenantModel, nil).Once()
				return TenantSvc
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := &automock.BusinessTenantMappingService{}
			if testCase.TenantSvcFn != nil {
				tenantSvc = testCase.TenantSvcFn()
			}
			tenantConverter := &automock.BusinessTenantMappingConverter{}
			if testCase.TenantConverterFn != nil {
				tenantConverter = testCase.TenantConverterFn()
			}
			persist, transact := testCase.TxFn()
			resolver := tenant.NewResolver(transact, tenantSvc, tenantConverter, nil)

			// WHEN
			result, err := resolver.RootTenant(ctx, externalTenant)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedTenantGQL, result)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, tenantSvc, tenantConverter)
		})
	}
}
