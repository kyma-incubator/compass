package ord_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	applicationTypeLabel = "applicationType"
	testApplicationType  = "testApplicationType"
)

func TestService_SyncORDDocuments(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	sanitizedDoc := fixSanitizedORDDocument()
	var nilSpecInput *model.SpecInput
	var nilBundleID *string

	testApplication := fixApplicationPage().Data[0]
	testWebhook := fixWebhooks()[0]

	api1PreSanitizedHash, err := ord.HashObject(fixORDDocument().APIResources[0])
	require.NoError(t, err)

	api2PreSanitizedHash, err := ord.HashObject(fixORDDocument().APIResources[1])
	require.NoError(t, err)

	event1PreSanitizedHash, err := ord.HashObject(fixORDDocument().EventResources[0])
	require.NoError(t, err)

	event2PreSanitizedHash, err := ord.HashObject(fixORDDocument().EventResources[1])
	require.NoError(t, err)

	packagePreSanitizedHash, err := ord.HashObject(fixORDDocument().Packages[0])
	require.NoError(t, err)

	thirdTransactionNotCommited := func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
		persistTx := &persistenceautomock.PersistenceTx{}
		persistTx.On("Commit").Return(nil).Twice()

		transact := &persistenceautomock.Transactioner{}
		transact.On("Begin").Return(persistTx, nil).Times(3)
		transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
		transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
		return persistTx, transact
	}

	successfulLabelRepo := func() *automock.LabelRepository {
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListGlobalByKeyAndObjects", txtest.CtxWithDBMatcher(), model.ApplicationLabelableObject, mock.MatchedBy(func(objectIDs []string) bool {
			return len(objectIDs) == 1 && objectIDs[0] == testApplication.ID
		}), applicationTypeLabel).Return([]*model.Label{
			{
				Value:    testApplicationType,
				ObjectID: testApplication.ID,
			},
		}, nil).Once()
		return labelRepo
	}

	successfulAppListAndGet := func() *automock.ApplicationService {
		appSvc := &automock.ApplicationService{}
		appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
		appSvc.On("GetForUpdate", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil, nil).Once()
		return appSvc
	}

	successfulTenantSvc := func() *automock.TenantService {
		tenantSvc := &automock.TenantService{}
		tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
		tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Once()
		return tenantSvc
	}

	successfulWebhookList := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListForApplicationWithSelectForUpdate", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooks(), nil).Once()
		whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
		return whSvc
	}

	successfulBundleUpdate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("Update", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0])).Return(nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0]).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleReferenceFetchingOfAPIBundleIDs := func() *automock.BundleReferenceService {
		bundleRefSvc := &automock.BundleReferenceService{}
		firstAPIID := api1ID
		secondAPIID := api2ID
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &firstAPIID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &secondAPIID).Return([]string{bundleID}, nil).Once()
		return bundleRefSvc
	}

	successfulBundleReferenceFetchingOfBundleIDs := func() *automock.BundleReferenceService {
		bundleRefSvc := &automock.BundleReferenceService{}
		firstAPIID := api1ID
		secondAPIID := api2ID
		firstEventID := event1ID
		secondEventID := event2ID
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &firstAPIID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &secondAPIID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &firstEventID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &secondEventID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &firstAPIID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &secondAPIID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &firstEventID).Return([]string{bundleID}, nil).Once()
		bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &secondEventID).Return([]string{bundleID}, nil).Once()
		return bundleRefSvc
	}

	successfulVendorUpdate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorID, *sanitizedDoc.Vendors[0]).Return(nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorID2, *sanitizedDoc.Vendors[1]).Return(nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulVendorCreate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[0]).Return("", nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulProductUpdate := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		productSvc.On("Update", txtest.CtxWithDBMatcher(), productID, *sanitizedDoc.Products[0]).Return(nil).Once()
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulProductCreate := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		productSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Products[0]).Return("", nil).Once()
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulPackageUpdate := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackagesWithHash(), nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		packagesSvc.On("Update", txtest.CtxWithDBMatcher(), packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageCreate := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulEmptyAPIList := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()

		return apiSvc
	}

	successfulEmptyEventList := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()

		return eventSvc
	}

	successfulEmptyPackageList := func() *automock.PackageService {
		pkgService := &automock.PackageService{}
		pkgService.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()

		return pkgService
	}

	successfulEmptyVendorList := func() *automock.VendorService {
		vendorService := &automock.VendorService{}
		vendorService.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()

		return vendorService
	}

	successfulSpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs()[0], model.EventSpecReference, event2ID).Return("", nil).Once()

		return specSvc
	}

	successfulSpecRefetch := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1Specs(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api1spec1ID, model.APISpecReference).Return(fixSuccessfulFetchRequest(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api1spec2ID, model.APISpecReference).Return(fixSuccessfulFetchRequest(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api1spec3ID, model.APISpecReference).Return(fixFailedFetchRequest(), nil).Once()

		specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), api1spec3ID, model.APISpecReference).Return(nil, nil).Once()

		specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(fixAPI2Specs(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api2spec1ID, model.APISpecReference).Return(fixSuccessfulFetchRequest(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api2spec2ID, model.APISpecReference).Return(fixFailedFetchRequest(), nil).Once()

		specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), api2spec2ID, model.APISpecReference).Return(nil, nil).Once()

		specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1Specs(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), event1specID, model.EventSpecReference).Return(fixFailedFetchRequest(), nil).Once()

		specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), event1specID, model.EventSpecReference).Return(nil, nil).Once()

		specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(fixEvent2Specs(), nil).Once()
		specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), event2specID, model.EventSpecReference).Return(fixFailedFetchRequest(), nil).Once()

		specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), event2specID, model.EventSpecReference).Return(nil, nil).Once()

		return specSvc
	}

	successfulAPISpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

		return specSvc
	}

	successfulAPIUpdate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulAPICreate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulEventUpdate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Twice()
		return eventSvc
	}

	successfulEventCreate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], fixEvent2SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulGlobalRegistrySvc := func() *automock.GlobalRegistryService {
		globalRegistrySvc := &automock.GlobalRegistryService{}
		globalRegistrySvc.On("SyncGlobalResources", context.TODO()).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
		return globalRegistrySvc
	}

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		labelRepoFn       func() *automock.LabelRepository
		appSvcFn          func() *automock.ApplicationService
		webhookSvcFn      func() *automock.WebhookService
		bundleSvcFn       func() *automock.BundleService
		bundleRefSvcFn    func() *automock.BundleReferenceService
		apiSvcFn          func() *automock.APIService
		eventSvcFn        func() *automock.EventService
		specSvcFn         func() *automock.SpecService
		packageSvcFn      func() *automock.PackageService
		productSvcFn      func() *automock.ProductService
		vendorSvcFn       func() *automock.VendorService
		tombstoneSvcFn    func() *automock.TombstoneService
		tenantSvcFn       func() *automock.TenantService
		globalRegistrySvc func() *automock.GlobalRegistryService
		clientFn          func() *automock.Client
		ExpectedErr       error
	}{
		{
			Name: "Success when resources are already in db and APIs/Events versions are incremented should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:    successfulLabelRepo,
			appSvcFn:       successfulAppListAndGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecUpdate,
			packageSvcFn: successfulPackageUpdate,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name: "Success when resources are already in db and APIs/Events versions are NOT incremented should Update them and refetch only failed API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:    successfulLabelRepo,
			appSvcFn:       successfulAppListAndGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				return eventSvc
			},
			specSvcFn:    successfulSpecRefetch,
			packageSvcFn: successfulPackageUpdate,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name: "Success when resources are not in db should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:  successfulLabelRepo,
			appSvcFn:     successfulAppListAndGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name: "Error when synchronizing global resources from global registry should get them from DB and proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:  successfulLabelRepo,
			appSvcFn:     successfulAppListAndGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: func() *automock.GlobalRegistryService {
				globalRegistrySvc := &automock.GlobalRegistryService{}
				globalRegistrySvc.On("SyncGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				globalRegistrySvc.On("ListGlobalResources", context.TODO()).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
				return globalRegistrySvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name: "Error when synchronizing global resources from global registry and get them from DB should proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:  successfulLabelRepo,
			appSvcFn:     successfulAppListAndGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: func() *automock.GlobalRegistryService {
				globalRegistrySvc := &automock.GlobalRegistryService{}
				globalRegistrySvc.On("SyncGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				globalRegistrySvc.On("ListGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				return globalRegistrySvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Returns error when app templates webhooks list fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(nil, testErr).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       testErr,
		},
		{
			Name: "Returns error when app list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       testErr,
		},
		{
			Name: "Returns error when labels list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListGlobalByKeyAndObjects", txtest.CtxWithDBMatcher(), model.ApplicationLabelableObject, mock.Anything, applicationTypeLabel).Return(nil, testErr).Once()
				return labelRepo
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       testErr,
		},

		{
			Name:              "Returns error when transaction opening fails",
			TransactionerFn:   txGen.ThatFailsOnBegin,
			ExpectedErr:       testErr,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when app list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when get internal tenant id fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 app"),
		},
		{
			Name:            "Returns error when get tenant fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return tenantSvc
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 app"),
		},
		{
			Name:            "Returns error when application looking fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			tenantSvcFn:     successfulTenantSvc,
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
				appSvc.On("GetForUpdate", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 app"),
		},
		{
			Name:            "Does not resync resources when event list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			webhookSvcFn:    successfulWebhookList,
			clientFn:        successfulClientFetch,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources when api list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			webhookSvcFn:    successfulWebhookList,
			clientFn:        successfulClientFetch,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when webhook list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			labelRepoFn: successfulLabelRepo,
			appSvcFn:    successfulAppListAndGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationWithSelectForUpdate", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				whSvc.On("ListForApplicationTemplates", txtest.CtxWithDBMatcher()).Return(fixWebhooks(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 app"),
		},
		{
			Name: "Returns error when processing application webhooks",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			labelRepoFn:    successfulLabelRepo,
			appSvcFn:       successfulAppListAndGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecUpdate,
			packageSvcFn: successfulPackageUpdate,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 applications"),
		},
		{
			Name:            "Skips app when ORD documents fetch fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(nil, "", testErr)
				return client
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources for invalid ORD documents",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if vendor list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if vendor update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorID, *sanitizedDoc.Vendors[0]).Return(testErr).Once()
				return vendorSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if vendor create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[0]).Return("", testErr).Once()
				return vendorSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if product list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return productSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if product update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), productID, *sanitizedDoc.Products[0]).Return(testErr).Once()
				return productSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if product create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Products[0]).Return("", testErr).Once()
				return productSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if package list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if package update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(testErr).Once()
				return packagesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if package create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", testErr).Once()
				return packagesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if bundle list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if bundle update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("Update", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0])).Return(testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if bundle create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0]).Return("", testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if api list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name:            "Does not resync resources if fetching bundle ids for api fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn: func() *automock.BundleReferenceService {
				bundleRefSvc := &automock.BundleReferenceService{}
				firstAPIID := api1ID
				bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &firstAPIID).Return(nil, testErr).Once()
				return bundleRefSvc
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()

				return apiSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(testErr).Once()
				return apiSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", testErr).Once()
				return apiSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api spec delete fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(testErr).Once()
				return specSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
		},
		{
			Name:            "Does not resync resources if api spec create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", testErr).Once()
				return specSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
		},
		{
			Name:            "Does not resync resources if api spec list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api spec get fetch request fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1Specs(), nil).Once()
				specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api1spec1ID, model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api spec refetch fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1Specs(), nil).Once()
				specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), api1spec1ID, model.APISpecReference).Return(fixFailedFetchRequest(), nil).Once()
				specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), api1spec1ID, model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn:       successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if fetching bundle ids for event fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn: func() *automock.BundleReferenceService {
				bundleRefSvc := &automock.BundleReferenceService{}
				firstAPIID := api1ID
				secondAPIID := api2ID
				firstEventID := event1ID
				bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &firstAPIID).Return([]string{bundleID}, nil).Once()
				bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleAPIReference, &secondAPIID).Return([]string{bundleID}, nil).Once()
				bundleRefSvc.On("GetBundleIDsForObject", txtest.CtxWithDBMatcher(), model.BundleEventReference, &firstEventID).Return(nil, testErr).Once()
				return bundleRefSvc
			},
			apiSvcFn:  successfulAPIUpdate,
			specSvcFn: successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn:       successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn:     successfulBundleCreate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec delete fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec get fetch request fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1Specs(), nil).Once()
				specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), event1specID, model.EventSpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec refetch fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil).Once()

				specSvc.On("ListByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1Specs(), nil).Once()
				specSvc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), event1specID, model.EventSpecReference).Return(fixFailedFetchRequest(), nil).Once()

				specSvc.On("RefetchSpec", txtest.CtxWithDBMatcher(), event1specID, model.EventSpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone list fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			eventSvcFn:      successfulEventUpdate,
			specSvcFn:       successfulSpecUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone update fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			eventSvcFn:      successfulEventUpdate,
			specSvcFn:       successfulSpecUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone create fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(testErr).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if package resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("Delete", txtest.CtxWithDBMatcher(), packageID).Return(testErr).Once()
				return packagesSvc
			},
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = packageORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = packageORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = packageORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if event resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], fixEvent2SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("Delete", txtest.CtxWithDBMatcher(), event1ID).Return(testErr).Once()
				return eventSvc
			},
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = event1ORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = event1ORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = event1ORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if vendor resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			packageSvcFn:    successfulPackageCreate,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[0]).Return("", nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), vendorID).Return(testErr).Once()
				return vendorSvc
			},
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = vendorORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = vendorORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = vendorORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if product resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			packageSvcFn:    successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Products[0]).Return("", nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Delete", txtest.CtxWithDBMatcher(), productID).Return(testErr).Once()
				return productSvc
			},
			vendorSvcFn: successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = productORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = productORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = productORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if bundle resource deletion due to tombstone fails",
			TransactionerFn: thirdTransactionNotCommited,
			labelRepoFn:     successfulLabelRepo,
			appSvcFn:        successfulAppListAndGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0]).Return("", nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("Delete", txtest.CtxWithDBMatcher(), bundleID).Return(testErr).Once()
				return bundlesSvc
			},
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = bundleORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = bundleORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = bundleORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:  successfulLabelRepo,
			appSvcFn:     successfulAppListAndGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixAPI1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixAPI2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulEmptyVendorList,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			labelRepoFn:    successfulLabelRepo,
			appSvcFn:       successfulAppListAndGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecUpdate,
			packageSvcFn: successfulPackageUpdate,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulEmptyVendorList,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhook).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			labelRepo := &automock.LabelRepository{}
			if test.labelRepoFn != nil {
				labelRepo = test.labelRepoFn()
			}
			appSvc := &automock.ApplicationService{}
			if test.appSvcFn != nil {
				appSvc = test.appSvcFn()
			}
			whSvc := &automock.WebhookService{}
			if test.webhookSvcFn != nil {
				whSvc = test.webhookSvcFn()
			}
			bndlSvc := &automock.BundleService{}
			if test.bundleSvcFn != nil {
				bndlSvc = test.bundleSvcFn()
			}
			bndlRefSvc := &automock.BundleReferenceService{}
			if test.bundleRefSvcFn != nil {
				bndlRefSvc = test.bundleRefSvcFn()
			}
			apiSvc := &automock.APIService{}
			if test.apiSvcFn != nil {
				apiSvc = test.apiSvcFn()
			}
			eventSvc := &automock.EventService{}
			if test.eventSvcFn != nil {
				eventSvc = test.eventSvcFn()
			}
			specSvc := &automock.SpecService{}
			if test.specSvcFn != nil {
				specSvc = test.specSvcFn()
			}
			packageSvc := &automock.PackageService{}
			if test.packageSvcFn != nil {
				packageSvc = test.packageSvcFn()
			}
			productSvc := &automock.ProductService{}
			if test.productSvcFn != nil {
				productSvc = test.productSvcFn()
			}
			vendorSvc := &automock.VendorService{}
			if test.vendorSvcFn != nil {
				vendorSvc = test.vendorSvcFn()
			}
			tombstoneSvc := &automock.TombstoneService{}
			if test.tombstoneSvcFn != nil {
				tombstoneSvc = test.tombstoneSvcFn()
			}
			tenantSvc := &automock.TenantService{}
			if test.tenantSvcFn != nil {
				tenantSvc = test.tenantSvcFn()
			}
			globalRegistrySvc := &automock.GlobalRegistryService{}
			if test.globalRegistrySvc != nil {
				globalRegistrySvc = test.globalRegistrySvc()
			}
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}

			ordCfg := ord.NewServiceConfig(4)
			svc := ord.NewAggregatorService(ordCfg, tx, labelRepo, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
			err := svc.SyncORDDocuments(context.TODO())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, labelRepo, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
		})
	}
}
