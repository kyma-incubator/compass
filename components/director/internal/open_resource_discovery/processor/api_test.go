package processor_test

import (
	"context"
	"testing"
	"time"

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

func TestAPIProcessor_Process(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	fixAPIDef := []*model.APIDefinition{
		fixAPI(apiID, str.Ptr(apiORDID)),
	}

	fixUpdatedAPIDef := []*model.APIDefinition{
		fixAPI(apiID, str.Ptr(apiORDID)),
	}

	fixUpdatedAPIDef[0].PackageID = str.Ptr(packageID2)

	fixAPIDef2 := []*model.APIDefinition{
		fixAPI(apiID, str.Ptr(apiORDID2)),
	}

	fixAPIInputs := []*model.APIDefinitionInput{
		fixAPIInput(),
	}

	successfulBundleReferenceGet := func() *automock.BundleReferenceService {
		bundleReferenceSvc := &automock.BundleReferenceService{}
		bundleReferenceSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &fixAPIDef[0].ID).Return([]string{}, nil).Once()
		return bundleReferenceSvc
	}

	successfulEntityTypeMapping := func() *automock.EntityTypeMappingService {
		entityTypeMappingSvc := &automock.EntityTypeMappingService{}
		entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{fixEntityTypeMappingModel(apiID)}, nil).Once()
		entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, fixAPIInput().EntityTypeMappings[0]).Return("", nil).Once()
		entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, apiID).Return(nil).Once()
		return entityTypeMappingSvc
	}

	resourceHashes := map[string]uint64{apiORDID: uint64ResourceHash}

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APISvcFn                   func() *automock.APIService
		EntityTypeSvcFn            func() *automock.EntityTypeService
		EntityTypeMappingSvcFn     func() *automock.EntityTypeMappingService
		BundleReferenceSvcFn       func() *automock.BundleReferenceService
		SpecSvcFn                  func() *automock.SpecService
		InputResource              resource.Type
		InputResourceID            string
		InputBundlesFromDB         []*model.Bundle
		InputPackagesFromDB        []*model.Package
		APIInput                   []*model.APIDefinitionInput
		InputResourceHashes        map[string]uint64
		ExpectedAPIDefOutput       []*model.APIDefinition
		ExpectedFetchRequestOutput []*processor.OrdFetchRequest
		ExpectedErr                error
	}{
		{
			Name: "Success empty API inputs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Twice()
				return apiSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixEmptyPackages(),
			APIInput:                   []*model.APIDefinitionInput{},
			InputResourceHashes:        resourceHashes,
			ExpectedAPIDefOutput:       fixAPIDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success - no change in resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				spec1 := fixAPIInputs[0].ResourceDefinitions[0].ToSpec()
				spec2 := fixAPIInputs[0].ResourceDefinitions[1].ToSpec()
				spec3 := fixAPIInputs[0].ResourceDefinitions[2].ToSpec()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, apiID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec2, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec3, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixPackages(),
			APIInput:                   fixAPIInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedAPIDefOutput:       fixAPIDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}},
		},
		{
			Name: "Success - refetch specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, apiID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{}, model.APISpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest()}, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixPackages(),
			APIInput:                   fixAPIInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedAPIDefOutput:       fixAPIDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{},
		},
		{
			Name: "Success - API not found",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Twice()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return(apiID, nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{}, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, fixAPIInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				apiInput := fixAPIInput()
				apiInput.OrdID = str.Ptr(apiORDID2)
				spec1 := apiInput.ResourceDefinitions[0].ToSpec()
				spec2 := apiInput.ResourceDefinitions[1].ToSpec()
				spec3 := apiInput.ResourceDefinitions[2].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec2, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec3, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixPackages(),
			APIInput:                   fixAPIInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedAPIDefOutput:       fixAPIDef2,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}},
		},
		{
			Name: "Fail while beginning transaction for listing APIs from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            []*model.APIDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing APIs by application id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            []*model.APIDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing APIs by application id from DB after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsTwice()
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            []*model.APIDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing APIs by Application Template Version id from DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return apiSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            []*model.APIDefinitionInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while updating in many bundles API",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(testErr).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{}, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, fixAPIInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			BundleReferenceSvcFn: successfulBundleReferenceGet,
			InputResource:        resource.Application,
			InputResourceID:      appID,
			InputBundlesFromDB:   fixEmptyBundles(),
			InputPackagesFromDB:  fixPackages(),
			APIInput:             fixAPIInputs,
			InputResourceHashes:  resourceHashes,
			ExpectedErr:          testErr,
		},
		{
			Name: "Fail while creating api",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return("", testErr).Once()
				return apiSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            fixAPIInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing by owner resource id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return(apiID, nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return(nil, testErr).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            fixAPIInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while creating entity type mapping",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return(apiID, nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{}, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, fixAPIInput().EntityTypeMappings[0]).Return("", testErr).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            fixAPIInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while deleting entity type mapping",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()

				return persistTx, transact
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return(apiID, nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{fixEntityTypeMappingModel(apiID)}, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, apiID).Return(testErr).Once()
				return entityTypeMappingSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            fixAPIInputs,
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
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef2, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				fixAPIInputs[0].LastUpdate = &currentTime

				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilString, nilString, *fixAPIInputs[0], nilSpecInputSlice, map[string]string{}, uint64ResourceHash, "").Return(apiID, nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), apiID, resource.API).Return([]*model.EntityTypeMapping{}, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, fixAPIInput().EntityTypeMappings[0]).Return("", nil).Once()
				return entityTypeMappingSvc
			},
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				apiInput := fixAPIInput()
				apiInput.OrdID = str.Ptr(apiORDID2)
				spec1 := apiInput.ResourceDefinitions[0].ToSpec()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.APISpecReference, apiID).Return("", nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixEmptyPackages(),
			APIInput:            fixAPIInputs,
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
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, apiID).Return(testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixPackages(),
			APIInput:            fixAPIInputs,
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
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Once()
				fixAPIInputs := []*model.APIDefinitionInput{
					fixAPIInput(),
				}
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, apiID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{}, model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixPackages(),
			APIInput: []*model.APIDefinitionInput{
				fixAPIInput(),
			},
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
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIsNoNewerLastUpdate(), nil).Once()
				fixAPIInputs := []*model.APIDefinitionInput{
					fixAPIInput(),
				}
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, fixAPIDef[0].ID, str.Ptr(packageID1), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.APISpecReference, apiID).Return([]string{}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDsGlobal", txtest.CtxWithDBMatcher(), []string{}, model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputBundlesFromDB:  fixEmptyBundles(),
			InputPackagesFromDB: fixPackages(),
			APIInput: []*model.APIDefinitionInput{
				fixAPIInput(),
			},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Success when updating package id for API",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			APISvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				fixAPIInputs[0].OrdPackageID = str.Ptr(packageORDID2)
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIDef, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixUpdatedAPIDef, nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, fixAPIDef[0].ID, str.Ptr(packageID2), *fixAPIInputs[0], nilSpecInput, map[string]string{}, map[string]string{}, []string{}, uint64ResourceHash, "").Return(nil).Once()
				return apiSvc
			},
			EntityTypeMappingSvcFn: successfulEntityTypeMapping,
			BundleReferenceSvcFn:   successfulBundleReferenceGet,
			SpecSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				spec1 := fixAPIInputs[0].ResourceDefinitions[0].ToSpec()
				spec2 := fixAPIInputs[0].ResourceDefinitions[1].ToSpec()
				spec3 := fixAPIInputs[0].ResourceDefinitions[2].ToSpec()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, apiID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec1, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec2, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *spec3, resource.Application, model.APISpecReference, apiID).Return("", nil, nil).Once()
				return specSvc
			},
			InputResource:              resource.Application,
			InputResourceID:            appID,
			InputBundlesFromDB:         fixEmptyBundles(),
			InputPackagesFromDB:        fixPackages(),
			APIInput:                   fixAPIInputs,
			InputResourceHashes:        resourceHashes,
			ExpectedAPIDefOutput:       fixUpdatedAPIDef,
			ExpectedFetchRequestOutput: []*processor.OrdFetchRequest{{FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}, {FetchRequest: nil, RefObjectOrdID: apiORDID}},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			apiSvc := &automock.APIService{}
			if test.APISvcFn != nil {
				apiSvc = test.APISvcFn()
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
			apiProcessor := processor.NewAPIProcessor(tx, apiSvc, entityTypeSvc, entityTypeMappingSvc, bundleReferenceSvc, specSvc)
			apis, fetchReq, err := apiProcessor.Process(ctx, test.InputResource, test.InputResourceID, test.InputBundlesFromDB, test.InputPackagesFromDB, test.APIInput, test.InputResourceHashes)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedAPIDefOutput, apis)
				require.Equal(t, test.ExpectedFetchRequestOutput, fetchReq)
			}

			mock.AssertExpectationsForObjects(t, tx, apiSvc, entityTypeSvc, entityTypeMappingSvc, bundleReferenceSvc, specSvc)
		})
	}
}
