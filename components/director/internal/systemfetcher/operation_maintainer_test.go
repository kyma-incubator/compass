package systemfetcher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOperationMaintainer_Maintain(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	tenants := []*model.BusinessTenantMapping{
		{
			ID:             "id-1",
			Type:           tenantpkg.Account,
			ExternalTenant: "external-id-1",
		},
		{
			ID:             "id-2",
			Type:           tenantpkg.Customer,
			ExternalTenant: "external-id-2",
		},
		{
			ID:             "id-3",
			Type:           tenantpkg.Unknown,
			ExternalTenant: "external-id-3",
		},
	}
	operation := &model.Operation{
		ID:            "op-id",
		OpType:        "",
		Status:        "",
		ErrorSeverity: model.OperationErrorSeverityNone,
		Data:          json.RawMessage("{}"),
	}
	operations := []*model.Operation{
		operation,
	}
	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationSvcFn             func() *automock.OperationService
		BusinessTenantMappingSvcFn func() *automock.BusinessTenantMappingService
		ExpectedErr                error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			OperationSvcFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("ListAllByType", txtest.CtxWithDBMatcher(), model.OperationTypeSystemFetching).Return([]*model.Operation{operation}, nil).Once()
				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.AnythingOfType("[]*model.OperationInput")).Run(func(args mock.Arguments) {
					arg := args.Get(1)
					res, ok := arg.([]*model.OperationInput)
					if !ok {
						return
					}
					assert.Equal(t, 2, len(res))
				}).Return(nil).Once()
				svc.On("DeleteMultiple", txtest.CtxWithDBMatcher(), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(1)
					res, ok := arg.([]string)
					if !ok {
						return
					}
					assert.Equal(t, 1, len(res))
				}).Return(nil).Once()
				return svc
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				svc := &automock.BusinessTenantMappingService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return svc
			},
		},
		{
			Name: "Error while beginning transaction",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while listing tenants",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				svc := &automock.BusinessTenantMappingService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while list all operations",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationSvcFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("ListAllByType", txtest.CtxWithDBMatcher(), model.OperationTypeSystemFetching).Return(nil, testErr).Once()
				return svc
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				svc := &automock.BusinessTenantMappingService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while create multiple operations",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationSvcFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("ListAllByType", txtest.CtxWithDBMatcher(), model.OperationTypeSystemFetching).Return(operations, nil).Once()
				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.Anything).Return(testErr).Once()
				return svc
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				svc := &automock.BusinessTenantMappingService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while delete multiple operations",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationSvcFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("ListAllByType", txtest.CtxWithDBMatcher(), model.OperationTypeSystemFetching).Return(operations, nil).Once()
				svc.On("CreateMultiple", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil).Once()
				svc.On("DeleteMultiple", txtest.CtxWithDBMatcher(), mock.Anything).Return(testErr).Once()
				return svc
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				svc := &automock.BusinessTenantMappingService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(tenants, nil).Once()
				return svc
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			_, tx := testCase.TransactionerFn()
			opSvc := testCase.OperationSvcFn()
			businessTenantMappingSvc := testCase.BusinessTenantMappingSvcFn()

			opMaintainer := systemfetcher.NewOperationMaintainer(model.OperationTypeSystemFetching, tx, opSvc, businessTenantMappingSvc)

			// WHEN
			err := opMaintainer.Maintain(ctx)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, opSvc, businessTenantMappingSvc)
		})
	}
}
