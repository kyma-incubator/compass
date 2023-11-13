package processor_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventProcessor_Process(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(errTest)

	fixEventDef := []*model.EventDefinition{
		fixEvent(event1ID, str.Ptr(eventORDID)),
	}

	fixEventDef2 := []*model.EventDefinition{
		fixEvent(event1ID, str.Ptr(eventORDID2)),
	}

	fixEventInputs := []*model.EventDefinitionInput{
		fixEventInput(),
	}

	successfulBundleReferenceGet := func() *automock.BundleReferenceService {
		bundleReferenceSvc := &automock.BundleReferenceService{}
		bundleReferenceSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &fixEventDef[0].ID).Return([]string{}, nil).Once()
		return bundleReferenceSvc
	}

	successfulEntityTypeMapping := func() *automock.EntityTypeMappingService {
		entityTypeMappingSvc := &automock.EntityTypeMappingService{}
		entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event1ID, resource.EventDefinition).Return([]*model.EntityTypeMapping{fixEntityTypeMappingModel(ID1)}, nil).Once()
		entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, event1ID, fixEventInput().EntityTypeMappings[0]).Return("", nil).Once()
		entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(nil).Once()
		return entityTypeMappingSvc
	}

	resourceHashes := map[string]uint64{ordID: uint64ResourceHash}

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventSvcFn                 func() *automock.EventService
		EntityTypeSvcFn            func() *automock.EntityTypeService
		EntityTypeMappingSvcFn     func() *automock.EntityTypeMappingService
		BundleReferenceSvcFn       func() *automock.BundleReferenceService
		SpecSvcFn                  func() *automock.SpecService
		InputResource              resource.Type
		InputResourceID            string
		InputBundlesFromDB         []*model.Bundle
		InputPackagesFromDB        []*model.Package
		EventInput                 []*model.EventDefinitionInput
		InputResourceHashes        map[string]uint64
		ExpectedEventDefOutput     []*model.EventDefinition
		ExpectedFetchRequestOutput []*processor.OrdFetchRequest
		ExpectedErr                error
	}{
		{
			Name: "Success empty Event inputs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Twice()
				return eventSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixEmptyPackages(),
			EventInput:                 []*model.EventDefinitionInput{},
			InputResourceHashes:        resourceHashes,
			ExpectedEventDefOutput:     fixEventDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixEventDef[0].ID, *fixEventInputs[0], nilSpecInput, []string{}, []string{}, []string{}, emptyHash, "").Return(nil).Once()

				return eventSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				spec := fixEventInputs[0].ResourceDefinitions[0].ToSpec()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec, resource.Application, model.EventSpecReference, event1ID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixEmptyPackages(),
			EventInput:                 fixEventInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedEventDefOutput:     fixEventDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: eventORDID}},
		},
		{
			Name: "Success - refetch specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				event := fixEvent(event1ID, str.Ptr(eventORDID))
				event.LastUpdate = str.Ptr("2024-01-25T15:47:04+00:00")
				eventModelsWithChangedLastUpdate := []*model.EventDefinition{
					event,
				}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(eventModelsWithChangedLastUpdate, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixEventDef[0].ID, *fixEventInputs[0], nilSpecInput, []string{}, []string{}, []string{}, emptyHash, "").Return(nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{}, model.EventSpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest()}, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixEmptyPackages(),
			EventInput:                 fixEventInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedEventDefOutput:     fixEventDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success - Event not found",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Twice()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, str.Ptr(packageID), *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return(ID1, nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), ID1, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1, fixEventInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				eventInput := fixEventInput()
				eventInput.OrdID = str.Ptr(apiORDID2)
				spec1 := eventInput.ResourceDefinitions[0].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.EventSpecReference, ID1).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixPackages(),
			EventInput:                 fixEventInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedEventDefOutput:     fixEventDef2,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: eventORDID}},
		},
		{
			Name: "Fail while beginning transaction for listing Events from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          []*model.EventDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while listing Events by application id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, errTest).Once()
				return eventSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          []*model.EventDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while listing Events by application id from DB after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, errTest).Once()
				return eventSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          []*model.EventDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while listing Events by Application Template Version id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, errTest).Once()
				return eventSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          []*model.EventDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while updating in many bundles API",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixEventDef[0].ID, *fixEventInputs[0], nilSpecInput, []string{}, []string{}, []string{}, emptyHash, "").Return(errTest).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event1ID, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, event1ID, fixEventInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			BundleReferenceSvcFn: successfulBundleReferenceGet,
			InputResource:        resource.Application,
			InputResourceID:      appID,
			InputBundlesFromDB:   fixEmptyBundles(),
			InputPackagesFromDB:  fixEmptyPackages(),
			EventInput:           fixEventInputs,
			InputResourceHashes:  resourceHashes,
			ExpectedErr:          errTest,
		},
		{
			Name: "Fail while creating event",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return("", errTest).Once()
				return eventSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while listing by owner resource id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return(ID1, nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), ID1, resource.EventDefinition).Return(nil, errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while creating entity type mapping",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return(ID1, nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), ID1, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1, fixEventInput().EntityTypeMappings[0]).Return("", errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while deleting entity type mapping",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return(ID1, nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), ID1, resource.EventDefinition).Return([]*model.EntityTypeMapping{fixEntityTypeMappingModel(ID1)}, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while creating spec by reference object id with delayed fetch request",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef2, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixEventInputs[0], nilSpecInputSlice, []string{}, emptyHash, "").Return(ID1, nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), ID1, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1, fixEventInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				eventInput := fixEventInput()
				eventInput.OrdID = str.Ptr(eventORDID2)
				spec1 := eventInput.ResourceDefinitions[0].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.EventSpecReference, ID1).Return("", nil, errTest).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fail while deleting spec by reference object id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			EventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventDef, nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixEventDef[0].ID, *fixEventInputs[0], nilSpecInput, []string{}, []string{}, []string{}, emptyHash, "").Return(nil).Once()
				return eventSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(errTest).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			EventInput:          fixEventInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			eventSvc := &automock.EventService{}
			if test.EventSvcFn != nil {
				eventSvc = test.EventSvcFn()
			}

			entityTypeSvc := &automock.EntityTypeService{}
			if test.EntityTypeSvcFn != nil {
				entityTypeSvc = test.EntityTypeSvcFn()
			}

			entityTypeMappingSvc := &automock.EntityTypeMappingService{}
			if test.EntityTypeMappingSvcFn != nil {
				entityTypeMappingSvc = test.EntityTypeMappingSvcFn()
			}

			bundleReferenceSvc := &automock.BundleReferenceService{}
			if test.BundleReferenceSvcFn != nil {
				bundleReferenceSvc = test.BundleReferenceSvcFn()
			}

			specSvc := &automock.SpecService{}
			if test.SpecSvcFn != nil {
				specSvc = test.SpecSvcFn()
			}

			ctx := context.TODO()
			ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)
			apiProcessor := processor.NewEventProcessor(tx, eventSvc, entityTypeSvc, entityTypeMappingSvc, bundleReferenceSvc, specSvc)
			events, fetchReq, err := apiProcessor.Process(ctx, test.InputResource, test.InputResourceID, test.InputBundlesFromDB, test.InputPackagesFromDB, test.EventInput, test.InputResourceHashes)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedEventDefOutput, events)
				require.Equal(t, test.ExpectedFetchRequestOutput, fetchReq)
			}

			mock.AssertExpectationsForObjects(t, tx, eventSvc, entityTypeSvc, entityTypeMappingSvc, bundleReferenceSvc, specSvc)
		})
	}
}
