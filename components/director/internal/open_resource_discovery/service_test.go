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
	testApplicationType = "testApplicationType"
)

func TestService_SyncORDDocuments(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	sanitizedDoc := fixSanitizedORDDocument()
	var nilSpecInput *model.SpecInput
	var nilBundleID *string

	testApplication := fixApplications()[0]
	testWebhookForApplication := fixWebhooksForApplication()[0]
	testWebhookForAppTemplate := fixOrdWebhooksForAppTemplate()[0]

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

	successfulAppGet := func() *automock.ApplicationService {
		appSvc := &automock.ApplicationService{}
		appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Once()
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
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn          func() *automock.ApplicationService
		webhookSvcFn      func() *automock.WebhookService
		bundleSvcFn       func() *automock.BundleService
		bundleRefSvcFn    func() *automock.BundleReferenceService
		apiSvcFn          func() *automock.APIService
		eventSvcFn        func() *automock.EventService
		specSvcFn         func() *automock.SpecService
		fetchReqFn        func() *automock.FetchRequestService
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
				return txGen.ThatSucceedsMultipleTimes(28)
			},
			appSvcFn:       successfulAppGet,
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
				return txGen.ThatSucceedsMultipleTimes(28)
			},
			appSvcFn:       successfulAppGet,
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
				return txGen.ThatSucceedsMultipleTimes(28)
			},
			appSvcFn:     successfulAppGet,
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
			Name: "Success when there is ORD webhook on app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(29)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Once()
				return appSvc
			},
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForAppTemplate).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
				return client
			},
		},
		{
			Name: "Error when synchronizing global resources from global registry should get them from DB and proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(28)
			},
			appSvcFn:     successfulAppGet,
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
				return txGen.ThatSucceedsMultipleTimes(28)
			},
			appSvcFn:     successfulAppGet,
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
			Name:            "Returns error when list by webhook type fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
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
			Name:              "Returns error when first transaction commit fails",
			TransactionerFn:   txGen.ThatFailsOnCommit,
			webhookSvcFn:      successfulWebhookList,
			ExpectedErr:       testErr,
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when second transaction begin fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(persistTx, testErr).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				return persistTx, transact
			},
			webhookSvcFn:      successfulWebhookList,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when second transaction commit fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				return persistTx, transact
			},
			appSvcFn:          successfulAppGet,
			tenantSvcFn:       successfulTenantSvc,
			webhookSvcFn:      successfulWebhookList,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when second transaction begin fails when there is app template ord webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(persistTx, testErr).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				return persistTx, transact
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when second transaction commit fails when there is app template ord webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				return appSvc
			},
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
			globalRegistrySvc: successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when get internal tenant id fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			webhookSvcFn:      successfulWebhookList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when get tenant fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return tenantSvc
			},
			webhookSvcFn:      successfulWebhookList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when application locking fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: successfulTenantSvc,
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn:      successfulWebhookList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources when event list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(3)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulWebhookList,
			clientFn:     successfulClientFetch,
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources when api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(3)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulWebhookList,
			clientFn:     successfulClientFetch,
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when list all applications by app template id fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when get internal tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				return appSvc
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when get tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				return appSvc
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Returns error when application locking fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Skips webhook when ORD documents fetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(nil, "", testErr)
				return client
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources for invalid ORD documents",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(3)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if vendor list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(4)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(4)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list vendors after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorID, *sanitizedDoc.Vendors[0]).Return(nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorID2, *sanitizedDoc.Vendors[1]).Return(nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if vendor update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(4)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(4)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if vendor create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(4)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(4)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if product list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorSvcFn:  successfulVendorUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list products after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(10)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(10)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorSvcFn:  successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), productID, *sanitizedDoc.Products[0]).Return(nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				return productSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			packageSvcFn:      successfulEmptyPackageList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if product update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorSvcFn:  successfulVendorUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if product create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorSvcFn:  successfulVendorUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if package list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(11)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(11)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(10)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list packages after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(13)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(13)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(12)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if package update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(11)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(12)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(11)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if package create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(11)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(12)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(11)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if bundle list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(13)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(14)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(13)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list bundles after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(15)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(16)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(15)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("Update", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0])).Return(nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if bundle update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(14)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(15)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(14)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if bundle create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(14)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(15)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(14)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(16)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
			bundleSvcFn:  successfulBundleUpdate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list apis after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(20)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			specSvcFn:      successfulAPISpecUpdate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if fetching bundle ids for api fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
			bundleSvcFn:  successfulBundleUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulBundleCreate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api spec get fetch request fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(17)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn:      successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list events after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(23)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(24)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs()[0], model.EventSpecReference, event2ID).Return("", nil).Once()

				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Twice()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if fetching bundle ids for event fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulVendorUpdate,
			packageSvcFn: successfulPackageUpdate,
			bundleSvcFn:  successfulBundleUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn:      successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductCreate,
			vendorSvcFn:    successfulVendorCreate,
			packageSvcFn:   successfulPackageCreate,
			bundleSvcFn:    successfulBundleCreate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:       successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}, mock.Anything, "").Return("", testErr).Once()
				return eventSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event spec get fetch request fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if tombstone list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(25)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(24)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Fails to list tombstones after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if tombstone update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(25)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			productSvcFn:   successfulProductUpdate,
			vendorSvcFn:    successfulVendorUpdate,
			packageSvcFn:   successfulPackageUpdate,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if tombstone create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(25)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if api resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
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
			ExpectedErr:       errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if package resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			ExpectedErr: errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if event resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			ExpectedErr: errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if vendor resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			ExpectedErr: errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if product resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			ExpectedErr: errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Does not resync resources if bundle resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			ExpectedErr: errors.New("failed to process 1 webhooks"),
		},
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(26)
			},
			appSvcFn:     successfulAppGet,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(26)
			},
			appSvcFn:       successfulAppGet,
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
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
			fetchReqSvc := &automock.FetchRequestService{}
			if test.fetchReqFn != nil {
				fetchReqSvc = test.fetchReqFn()
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

			ordCfg := ord.NewServiceConfig(4, 100)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
			err := svc.SyncORDDocuments(context.TODO())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
		})
	}
}
