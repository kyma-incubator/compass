package processor_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCapabilityProcessor_Process(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	fixCapabilities := []*model.Capability{
		fixCapability(capabilityID, str.Ptr(capabilityORDID)),
	}

	fixCapabilities2 := []*model.Capability{
		fixCapability(capabilityID, str.Ptr(capabilityORDID2)),
	}

	fixCapabilityInputs := []*model.CapabilityInput{
		fixCapabilityInput(),
	}

	resourceHashes := map[string]uint64{ordID: uint64ResourceHash}

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		CapabilitySvcFn            func() *automock.CapabilityService
		SpecSvcFn                  func() *automock.SpecService
		InputResource              resource.Type
		InputResourceID            string
		InputPackagesFromDB        []*model.Package
		CapabilityInput            []*model.CapabilityInput
		InputResourceHashes        map[string]uint64
		ExpectedCapabilityOutput   []*model.Capability
		ExpectedFetchRequestOutput []*processor.OrdFetchRequest
		ExpectedErr                error
	}{
		{
			Name: "Success empty API inputs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Twice()
				return capabilitySvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputPackagesFromDB:        fixEmptyPackages(),
			CapabilityInput:            []*model.CapabilityInput{},
			InputResourceHashes:        resourceHashes,
			ExpectedCapabilityOutput:   fixCapabilities,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Twice()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				spec := fixCapabilityInputs[0].CapabilityDefinitions[0].ToSpec()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capabilityID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec, resource.Application, model.CapabilitySpecReference, capabilityID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputPackagesFromDB:        fixEmptyPackages(),
			CapabilityInput:            fixCapabilityInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedCapabilityOutput:   fixCapabilities,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: capabilityORDID}},
		},
		{
			Name: "Success - refetch specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capabilityID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{}, model.CapabilitySpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest()}, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputPackagesFromDB:        fixEmptyPackages(),
			CapabilityInput:            fixCapabilityInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedCapabilityOutput:   fixCapabilities,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success - API not found",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				apiSvc := &automock.CapabilityService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities2, nil).Twice()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *fixCapabilityInputs[0], nilSpecInputSlice, emptyHash).Return(capabilityID, nil).Once()
				return apiSvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				capabilityInput := fixCapabilityInput()
				capabilityInput.OrdID = str.Ptr(capabilityORDID2)
				spec := capabilityInput.CapabilityDefinitions[0].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec, resource.Application, model.CapabilitySpecReference, capabilityID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputPackagesFromDB:        fixPackages(),
			CapabilityInput:            fixCapabilityInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedCapabilityOutput:   fixCapabilities2,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: capabilityORDID}},
		},
		{
			Name: "Fail while beginning transaction for listing capabilities from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     []*model.CapabilityInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing capabilities by application id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return capabilitySvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     []*model.CapabilityInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing capabilities by application id from DB after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return capabilitySvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     []*model.CapabilityInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing capabilities by Application Template Version id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return capabilitySvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     []*model.CapabilityInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while updating capabilities",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(testErr).Once()
				return capabilitySvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while creating api",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities2, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, *fixCapabilityInputs[0], nilSpecInputSlice, emptyHash).Return("", testErr).Once()
				return capabilitySvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while creating spec by reference object id with delayed fetch request",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities2, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, *fixCapabilityInputs[0], nilSpecInputSlice, emptyHash).Return(capabilityID, nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				capabilityInput := fixCapabilityInput()
				capabilityInput.OrdID = str.Ptr(capabilityORDID2)
				spec1 := capabilityInput.CapabilityDefinitions[0].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.CapabilitySpecReference, capabilityID).Return("", nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while deleting spec by reference object id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities, nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capabilityID).Return(testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing fetch requests by reference object ids",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capabilityID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{}, model.CapabilitySpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing fetch requests by reference object ids for application template version",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			CapabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, fixCapabilities[0].ID, *fixCapabilityInputs[0], emptyHash).Return(nil).Once()
				return capabilitySvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.CapabilitySpecReference, capabilityID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDsGlobal", txtest.CtxWithDBMatcher(), []string{}, model.CapabilitySpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputPackagesFromDB: fixEmptyPackages(),
			CapabilityInput:     fixCapabilityInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			capabilitySvc := &automock.CapabilityService{}
			if test.CapabilitySvcFn != nil {
				capabilitySvc = test.CapabilitySvcFn()
			}

			specSvc := &automock.SpecService{}
			if test.SpecSvcFn != nil {
				specSvc = test.SpecSvcFn()
			}

			ctx := context.TODO()
			ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)
			apiProcessor := processor.NewCapabilityProcessor(tx, capabilitySvc, specSvc)
			capabilities, fetchReq, err := apiProcessor.Process(ctx, test.InputResource, test.InputResourceID, test.InputPackagesFromDB, test.CapabilityInput, test.InputResourceHashes)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedCapabilityOutput, capabilities)
				require.Equal(t, test.ExpectedFetchRequestOutput, fetchReq)
			}

			mock.AssertExpectationsForObjects(t, tx, capabilitySvc, specSvc)
		})
	}
}
