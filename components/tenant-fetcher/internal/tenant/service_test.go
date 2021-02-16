package tenant_test

import (
	"context"
	"errors"

	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant/automock"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	//GIVEN
	testErr := errors.New("test error")
	ctx := context.Background()
	txGen := txtest.NewTransactionContextGenerator(testErr)
	tenantModel := model.TenantModel{
		ID:             testID,
		TenantId:       testID,
		Status:         tenantEntity.Active,
		TenantProvider: testProviderName,
	}

	testCases := []struct {
		Name           string
		TenantRepoFn   func() *automock.TenantRepository
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UidFn          func() *automock.UIDService
		ExpectedOutput error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error when creating tenant in database",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(testErr)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when committing transaction in database",
			TxFn: txGen.ThatFailsOnCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("Create", txtest.CtxWithDBMatcher(), tenantModel).Return(nil)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID)
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when beginning transaction in database",
			TxFn: txGen.ThatFailsOnBegin,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "Create")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.AssertNotCalled(t, "Generate")
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			tenantRepo := testCase.TenantRepoFn()
			uidSvc := testCase.UidFn()

			svc := tenant.NewService(tenantRepo, transact, uidSvc)

			//WHEN
			err := svc.Create(ctx, tenantModel)

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	//GIVEN
	testErr := errors.New("test error")
	ctx := context.Background()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name           string
		TenantRepoFn   func() *automock.TenantRepository
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UidFn          func() *automock.UIDService
		ExpectedOutput error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(nil).Once()
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error when deleting tenant from database",
			TxFn: txGen.ThatSucceeds,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(testErr)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when committing transaction in database",
			TxFn: txGen.ThatFailsOnCommit,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.On("DeleteByExternalID", txtest.CtxWithDBMatcher(), testID).Return(nil)
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when beginning transaction in database",
			TxFn: txGen.ThatFailsOnBegin,
			TenantRepoFn: func() *automock.TenantRepository {
				tenantMappingRepo := &automock.TenantRepository{}
				tenantMappingRepo.AssertNotCalled(t, "DeleteByExternalID")
				return tenantMappingRepo
			},
			UidFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, transact := testCase.TxFn()
			tenantRepo := testCase.TenantRepoFn()
			uidSvc := testCase.UidFn()

			svc := tenant.NewService(tenantRepo, transact, uidSvc)

			//WHEN
			err := svc.DeleteByExternalID(ctx, testID)

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

}
