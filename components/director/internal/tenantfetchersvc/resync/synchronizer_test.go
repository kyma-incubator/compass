package resync_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantsSynchronizer_Synchronize(t *testing.T) {
	ctx := context.TODO()
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	const (
		region          = "central"
		provider        = "test-provider"
		newTenantID     = "da363eb6-9444-4452-9bf6-40ee7e8da4d8"
		movedTenantID   = "5ca5aa4c-6498-45d0-87c6-d150b2cef1d2"
		deletedTenantID = "f2f99619-ab1c-4875-b4a6-15e5871fec39"

		parentTenantID         = "20e67c37-d1a1-418d-a61a-37b485a2f163"
		internalParentTenantID = "52f74825-83b5-46f4-884e-3ce2d589061f"

		failedToFetchNewTenantsErrMsg     = "failed to fetch new tenants"
		failedToFetchMovedTenantsErrMsg   = "failed to fetch moved tenants"
		failedToFetchDeletedTenantsErrMsg = "failed to fetch deleted tenants"
		failedToCreateTenantsErrMsg       = "failed to create tenants"
		failedToMoveTenantsErrMsg         = "failed to move tenants"
		failedToDeleteTenantsErrMsg       = "failed to delete tenants"
		failedToGetExistingTenantsErrMsg  = "failed to get existing tenants"
	)

	lastConsumedTenantTimestamp := strconv.FormatInt(time.Now().Add(time.Duration(-10)*time.Minute).UnixNano()/int64(time.Millisecond), 10) // 10 minutes ago
	lastResyncTimestamp := strconv.FormatInt(time.Now().Add(time.Duration(-5)*time.Minute).UnixNano()/int64(time.Millisecond), 10)          // 5 minutes ago

	jobCfg := resync.JobConfig{
		TenantProvider: provider,
		EventsConfig: resync.EventsConfig{
			RegionalAPIConfigs: map[string]*resync.EventsAPIConfig{region: {RegionName: region}},
		},
		ResyncConfig: resync.ResyncConfig{FullResyncInterval: time.Hour},
	}

	newAccountTenant := model.BusinessTenantMappingInput{ExternalTenant: newTenantID, Region: region}
	movedAccountTenant := model.MovedSubaccountMappingInput{SubaccountID: movedTenantID}
	deletedAccountTenant := model.BusinessTenantMappingInput{ExternalTenant: deletedTenantID}
	emptyTenantsResult := make([]model.BusinessTenantMappingInput, 0)

	kubeClientFn := func() *automock.KubeClient {
		client := &automock.KubeClient{}
		client.On("GetTenantFetcherConfigMapData", ctx).Return(lastConsumedTenantTimestamp, lastResyncTimestamp, nil)
		client.On("UpdateTenantFetcherConfigMapData", ctx, mock.Anything, mock.Anything).Return(nil)
		return client
	}

	kubeClientWithResyncTimestampFn := func() *automock.KubeClient {
		client := &automock.KubeClient{}
		client.On("GetTenantFetcherConfigMapData", ctx).Return(lastConsumedTenantTimestamp, lastResyncTimestamp, nil)
		return client
	}

	noOpMoverFn := func() *automock.TenantMover {
		svc := &automock.TenantMover{}
		svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
		return svc
	}

	testCases := []struct {
		Name               string
		JobCfg             resync.JobConfig
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantStorageSvcFn func() *automock.TenantStorageService
		TenantCreatorFn    func() *automock.TenantCreator
		TenantMoverFn      func() *automock.TenantMover
		TenantDeleterFn    func() *automock.TenantDeleter
		KubeClientFn       func() *automock.KubeClient
		ExpectedErrMsg     string
	}{
		{
			Name:            "Success when create, move and delete events are present for different tenants",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				deleted := deletedAccountTenant.ToBusinessTenantMapping("123")
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{newTenantID, deletedTenantID}).Return([]*model.BusinessTenantMapping{deleted}, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{newAccountTenant}, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{newAccountTenant}).Return(nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{movedAccountTenant}, nil)
				svc.On("MoveTenants", ctx, []model.MovedSubaccountMappingInput{movedAccountTenant}).Return(nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{deletedAccountTenant}, nil)
				svc.On("DeleteTenants", ctx, []model.BusinessTenantMappingInput{deletedAccountTenant}).Return(nil)
				return svc
			},
			KubeClientFn: kubeClientFn,
		},
		{
			Name:               "Success when no new events are present",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn: kubeClientFn,
		},
		{
			Name:            "Tenant is not created when both create and delete events are present for the same unknown tenant",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{deletedTenantID}).Return([]*model.BusinessTenantMapping{}, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{deletedAccountTenant}, nil)
				return svc
			},
			TenantMoverFn: noOpMoverFn,
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{deletedAccountTenant}, nil)
				return svc
			},
			KubeClientFn: kubeClientFn,
		},
		{
			Name:            "Parent tenant is also created when tenant from create event has unknown parent",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{parentTenantID, newTenantID}).Return([]*model.BusinessTenantMapping{}, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}

				tenantWithParent := model.BusinessTenantMappingInput{
					ExternalTenant: newTenantID,
					Parent:         parentTenantID,
					Type:           string(tenant.Subaccount),
					Region:         region,
				}
				expectedParentTenant := model.BusinessTenantMappingInput{
					Name:           parentTenantID,
					ExternalTenant: parentTenantID,
					Region:         region,
					Type:           string(tenant.Account),
					Provider:       provider,
				}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{tenantWithParent}, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{expectedParentTenant, tenantWithParent}).Return(nil)
				return svc
			},
			TenantMoverFn: noOpMoverFn,
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn: kubeClientFn,
		},
		{
			Name:            "Child tenant is correctly associated with internal ID of parent when parent is pre-existing",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				parentTenant := &model.BusinessTenantMapping{
					ID:             internalParentTenantID,
					ExternalTenant: parentTenantID,
				}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{parentTenantID, newTenantID}).Return([]*model.BusinessTenantMapping{parentTenant}, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}

				tenantWithParent := model.BusinessTenantMappingInput{
					ExternalTenant: newTenantID,
					Parent:         parentTenantID,
					Type:           string(tenant.Subaccount),
					Region:         region,
				}

				expectedTenantWithParent := tenantWithParent
				expectedTenantWithParent.Parent = internalParentTenantID

				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{tenantWithParent}, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{expectedTenantWithParent}).Return(nil)
				return svc
			},
			TenantMoverFn: noOpMoverFn,
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn: kubeClientFn,
		},
		{
			Name:               "Fails when fetching created tenants returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(nil, errors.New(failedToFetchNewTenantsErrMsg))
				return svc
			},
			TenantMoverFn:   func() *automock.TenantMover { return &automock.TenantMover{} },
			TenantDeleterFn: func() *automock.TenantDeleter { return &automock.TenantDeleter{} },
			KubeClientFn:    kubeClientWithResyncTimestampFn,
			ExpectedErrMsg:  failedToFetchNewTenantsErrMsg,
		},
		{
			Name:               "Fails when fetching moved tenants returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return(nil, errors.New(failedToFetchMovedTenantsErrMsg))
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter { return &automock.TenantDeleter{} },
			KubeClientFn:    kubeClientWithResyncTimestampFn,
			ExpectedErrMsg:  failedToFetchMovedTenantsErrMsg,
		},
		{
			Name:               "Fails when fetching deleted tenants returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			TenantMoverFn: noOpMoverFn,
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(nil, errors.New(failedToFetchDeletedTenantsErrMsg))
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: failedToFetchDeletedTenantsErrMsg,
		},
		{
			Name:            "Fails when creating new tenants returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{newTenantID}).Return(nil, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{newAccountTenant}, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{newAccountTenant}).Return(errors.New(failedToCreateTenantsErrMsg))
				return svc
			},
			TenantMoverFn: noOpMoverFn,
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: failedToCreateTenantsErrMsg,
		},
		{
			Name:               "Fails when moving tenants returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{movedAccountTenant}, nil)
				svc.On("MoveTenants", ctx, []model.MovedSubaccountMappingInput{movedAccountTenant}).Return(errors.New(failedToMoveTenantsErrMsg))
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: failedToMoveTenantsErrMsg,
		},
		{
			Name:            "Fails when deleting tenants returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{deletedTenantID}).Return([]*model.BusinessTenantMapping{deletedAccountTenant.ToBusinessTenantMapping(deletedTenantID)}, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{deletedAccountTenant}, nil)
				svc.On("DeleteTenants", ctx, []model.BusinessTenantMappingInput{deletedAccountTenant}).Return(errors.New(failedToDeleteTenantsErrMsg))
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: failedToDeleteTenantsErrMsg,
		},
		{
			Name:            "Fails when fetching existing tenants returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{newTenantID}).Return(nil, errors.New(failedToGetExistingTenantsErrMsg))
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{newAccountTenant}, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: failedToGetExistingTenantsErrMsg,
		},
		{
			Name:               "Fails when fetching existing tenants returns an error caused by failed transaction start",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatFailsOnBegin,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{newAccountTenant}, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: testErr.Error(),
		},
		{
			Name:            "Fails when fetching existing tenants returns an error caused by failed transaction commit",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatFailsOnCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("ListsByExternalIDs", txtest.CtxWithDBMatcher(), []string{newTenantID}).Return(nil, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				svc.On("TenantsToCreate", ctx, region, lastConsumedTenantTimestamp).Return([]model.BusinessTenantMappingInput{newAccountTenant}, nil)
				return svc
			},
			TenantMoverFn: func() *automock.TenantMover {
				svc := &automock.TenantMover{}
				svc.On("TenantsToMove", ctx, region, lastConsumedTenantTimestamp).Return([]model.MovedSubaccountMappingInput{}, nil)
				return svc
			},
			TenantDeleterFn: func() *automock.TenantDeleter {
				svc := &automock.TenantDeleter{}
				svc.On("TenantsToDelete", ctx, region, lastConsumedTenantTimestamp).Return(emptyTenantsResult, nil)
				return svc
			},
			KubeClientFn:   kubeClientWithResyncTimestampFn,
			ExpectedErrMsg: testErr.Error(),
		},
		{
			Name:               "Fails when getting resync info returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatDoesntStartTransaction,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn:    func() *automock.TenantCreator { return &automock.TenantCreator{} },
			TenantMoverFn:      func() *automock.TenantMover { return &automock.TenantMover{} },
			TenantDeleterFn:    func() *automock.TenantDeleter { return &automock.TenantDeleter{} },
			KubeClientFn: func() *automock.KubeClient {
				client := &automock.KubeClient{}
				client.On("GetTenantFetcherConfigMapData", ctx).Return("", "", testErr)
				return client
			},
			ExpectedErrMsg: testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			tenantCreator := testCase.TenantCreatorFn()
			tenantMover := testCase.TenantMoverFn()
			tenantDeleter := testCase.TenantDeleterFn()
			kubeClient := testCase.KubeClientFn()

			metricsPusher := &automock.AggregationFailurePusher{}
			if len(testCase.ExpectedErrMsg) > 0 {
				metricsPusher.On("ReportAggregationFailure", ctx, mock.MatchedBy(func(actual error) bool {
					return strings.Contains(actual.Error(), testCase.ExpectedErrMsg)
				}))
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, tenantStorageSvc, tenantCreator, tenantMover,
				tenantDeleter, kubeClient, metricsPusher)

			synchronizer := resync.NewTenantSynchronizer(testCase.JobCfg, transact, tenantStorageSvc, tenantCreator, tenantMover, tenantDeleter, kubeClient, metricsPusher)
			err := synchronizer.Synchronize(context.TODO())
			if len(testCase.ExpectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err, "unexpected error while running tenant resync")
			}
		})
	}
}

func TestTenantsSynchronizer_SynchronizeTenant(t *testing.T) {
	ctx := context.TODO()
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	const (
		region      = "central"
		newTenantID = "da363eb6-9444-4452-9bf6-40ee7e8da4d8"

		parentTenantID         = "20e67c37-d1a1-418d-a61a-37b485a2f163"
		internalParentTenantID = "52f74825-83b5-46f4-884e-3ce2d589061f"

		failedToFetchNewTenantsErrMsg    = "failed to fetch new tenants"
		failedToGetExistingTenantsErrMsg = "failed to get existing tenants"
	)

	jobCfg := resync.JobConfig{
		TenantProvider: resync.TenantOnDemandProvider,
	}

	newSubaccountTenant := model.BusinessTenantMappingInput{ExternalTenant: newTenantID, Parent: parentTenantID, Region: region, Type: string(tenant.Subaccount)}

	testCases := []struct {
		Name               string
		JobCfg             resync.JobConfig
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantStorageSvcFn func() *automock.TenantStorageService
		TenantCreatorFn    func() *automock.TenantCreator
		ExpectedErrMsg     string
	}{
		{
			Name:   "Success when create event is present for tenant",
			JobCfg: jobCfg,
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				parentTnt := &model.BusinessTenantMapping{ID: internalParentTenantID, ExternalTenant: parentTenantID}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newSubaccountTenant.Parent).Return(parentTnt, nil)

				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				tenantWithExistingParent := newSubaccountTenant
				tenantWithExistingParent.Parent = internalParentTenantID
				svc.On("FetchTenant", ctx, newTenantID).Return(&newSubaccountTenant, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{tenantWithExistingParent}).Return(nil)
				return svc
			},
		},
		{
			Name:            "[temporary] Success when create event is missing for tenant",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				lazilyStoredTnt := model.BusinessTenantMappingInput{
					Name:           newTenantID,
					ExternalTenant: newTenantID,
					Parent:         parentTenantID, // we expect the parent tenant ID to be internal tenant
					Type:           string(tenant.Subaccount),
					Provider:       "lazily-tenant-fetcher",
				}
				svc.On("FetchTenant", ctx, newTenantID).Return(nil, nil)
				svc.On("CreateTenants", ctx, []model.BusinessTenantMappingInput{lazilyStoredTnt}).Return(nil)
				return svc
			},
		},
		{
			Name:            "Success when tenant already exists",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(newSubaccountTenant.ToBusinessTenantMapping(newTenantID), nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator { return &automock.TenantCreator{} },
		},
		{
			Name:            "Fails when tenant from create event has no parent tenant",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				tenantWithoutParent := newSubaccountTenant
				tenantWithoutParent.Parent = ""
				svc.On("FetchTenant", ctx, newTenantID).Return(&tenantWithoutParent, nil)
				return svc
			},
			ExpectedErrMsg: fmt.Sprintf("parent tenant not found of tenant with ID %s", newTenantID),
		},
		{
			Name:            "Fails when checking for already existing tenant returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, errors.New(failedToGetExistingTenantsErrMsg))
				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator { return &automock.TenantCreator{} },
			ExpectedErrMsg:  failedToGetExistingTenantsErrMsg,
		},
		{
			Name:            "Fails when fetching tenant returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatSucceeds,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)

				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				tenantWithExistingParent := newSubaccountTenant
				tenantWithExistingParent.Parent = internalParentTenantID
				svc.On("FetchTenant", ctx, newTenantID).Return(nil, errors.New(failedToFetchNewTenantsErrMsg))
				return svc
			},
			ExpectedErrMsg: failedToFetchNewTenantsErrMsg,
		},
		{
			Name:   "Fails when parent retrieval returns an error",
			JobCfg: jobCfg,
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(1)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(2)

				return persistTx, transact
			},
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newSubaccountTenant.Parent).Return(nil, testErr)

				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator {
				svc := &automock.TenantCreator{}
				tenantWithExistingParent := newSubaccountTenant
				tenantWithExistingParent.Parent = internalParentTenantID
				svc.On("FetchTenant", ctx, newTenantID).Return(&newSubaccountTenant, nil)
				return svc
			},
			ExpectedErrMsg: testErr.Error(),
		},
		{
			Name:               "Fails when transaction start returns an error",
			JobCfg:             jobCfg,
			TransactionerFn:    txGen.ThatFailsOnBegin,
			TenantStorageSvcFn: func() *automock.TenantStorageService { return &automock.TenantStorageService{} },
			TenantCreatorFn:    func() *automock.TenantCreator { return &automock.TenantCreator{} },
			ExpectedErrMsg:     testErr.Error(),
		},
		{
			Name:            "Fails when first transaction commit returns an error",
			JobCfg:          jobCfg,
			TransactionerFn: txGen.ThatFailsOnCommit,
			TenantStorageSvcFn: func() *automock.TenantStorageService {
				svc := &automock.TenantStorageService{}
				svc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), newTenantID).Return(nil, nil)

				return svc
			},
			TenantCreatorFn: func() *automock.TenantCreator { return &automock.TenantCreator{} },
			ExpectedErrMsg:  testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			tenantStorageSvc := testCase.TenantStorageSvcFn()
			tenantCreator := testCase.TenantCreatorFn()

			defer mock.AssertExpectationsForObjects(t, persist, transact, tenantStorageSvc, tenantCreator)

			synchronizer := resync.NewTenantSynchronizer(testCase.JobCfg, transact, tenantStorageSvc, tenantCreator, nil, nil, nil, nil)
			err := synchronizer.SynchronizeTenant(ctx, parentTenantID, newTenantID)
			if len(testCase.ExpectedErrMsg) > 0 {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
