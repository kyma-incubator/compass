package open_resource_discovery_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_SyncORDDocuments(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	sanitizedDoc := fixSanitizedORDDocument()
	var nilSpecInput *model.SpecInput
	var nilBundleID *string

	secondTransactionNotCommited := func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
		persistTx := &persistenceautomock.PersistenceTx{}
		persistTx.On("Commit").Return(nil).Once()

		transact := &persistenceautomock.Transactioner{}
		transact.On("Begin").Return(persistTx, nil).Twice()
		transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Twice()
		return persistTx, transact
	}

	successfulAppList := func() *automock.ApplicationService {
		appSvc := &automock.ApplicationService{}
		appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(fixApplicationPage(), nil).Once()
		return appSvc
	}

	successfulWebhookList := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooks(), nil).Once()
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
		return bundleRefSvc
	}

	successfulVendorUpdate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorORDID, *sanitizedDoc.Vendors[0]).Return(nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendor2ORDID, *sanitizedDoc.Vendors[1]).Return(nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulSAPVendorUpdate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		sapVendor := model.VendorInput{
			OrdID: open_resource_discovery.SapVendor,
			Title: open_resource_discovery.SapTitle,
		}

		modelVendor := []*model.Vendor{
			{
				OrdID:         open_resource_discovery.SapVendor,
				TenantID:      tenantID,
				ApplicationID: appID,
				Title:         open_resource_discovery.SapTitle,
			}}

		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(modelVendor, nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorORDID, sapVendor).Return(nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(modelVendor, nil).Once()
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

	successfulSAPVendorCreate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		sapVendor := model.VendorInput{
			OrdID: open_resource_discovery.SapVendor,
			Title: open_resource_discovery.SapTitle,
		}

		modelVendor := []*model.Vendor{
			{
				OrdID:         open_resource_discovery.SapVendor,
				TenantID:      tenantID,
				ApplicationID: appID,
				Title:         open_resource_discovery.SapTitle,
			}}

		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, sapVendor).Return("", nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(modelVendor, nil).Once()
		return vendorSvc
	}

	successfulProductUpdate := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		productSvc.On("Update", txtest.CtxWithDBMatcher(), productORDID, *sanitizedDoc.Products[0]).Return(nil).Once()
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
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		packagesSvc.On("Update", txtest.CtxWithDBMatcher(), packageID, *sanitizedDoc.Packages[0]).Return(nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageCreate := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0]).Return("", nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulSpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs()[0], model.EventSpecReference, event2ID).Return("", nil).Once()
		return specSvc
	}

	successfulAPISpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
		specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()

		return specSvc
	}

	successfulAPIUpdate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}).Return(nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulAPICreate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixApi1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}).Return("", nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixApi2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}).Return("", nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulEventUpdate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{}, []string{}).Return(nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{}, []string{}).Return(nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulEventCreate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}).Return("", nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], fixEvent2SpecInputs(), []string{bundleID}).Return("", nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{fixORDDocument()}, nil)
		return client
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn        func() *automock.ApplicationService
		webhookSvcFn    func() *automock.WebhookService
		bundleSvcFn     func() *automock.BundleService
		bundleRefSvcFn  func() *automock.BundleReferenceService
		apiSvcFn        func() *automock.APIService
		eventSvcFn      func() *automock.EventService
		specSvcFn       func() *automock.SpecService
		packageSvcFn    func() *automock.PackageService
		productSvcFn    func() *automock.ProductService
		vendorSvcFn     func() *automock.VendorService
		tombstoneSvcFn  func() *automock.TombstoneService
		clientFn        func() *automock.Client
		ExpectedErr     error
	}{
		{
			Name: "Success when resources are already in db should Update them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn:       successfulAppList,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}).Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
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
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), sanitizedDoc.Tombstones[0].OrdID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name: "Success when resources are not in db should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn:     successfulAppList,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixApi1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}).Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixApi2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}).Return("", nil).Once()
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Returns error when transaction opening fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when app list fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListGlobal", txtest.CtxWithDBMatcher(), 200, "").Return(nil, testErr).Once()
				return appSvc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when webhook list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Twice()
				return persistTx, transact
			},
			appSvcFn: successfulAppList,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return whSvc
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Skips app when ORD documents fetch fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(nil, testErr)
				return client
			},
		},
		{
			Name:            "Does not resync resources for invalid ORD documents",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if vendor list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if vendor update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), vendorORDID, *sanitizedDoc.Vendors[0]).Return(testErr).Once()
				return vendorSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if vendor create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[0]).Return("", testErr).Once()
				return vendorSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if product list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return productSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if product update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), productORDID, *sanitizedDoc.Products[0]).Return(testErr).Once()
				return productSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if product create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			vendorSvcFn:     successfulVendorUpdate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Products[0]).Return("", testErr).Once()
				return productSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if package list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if package update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), packageID, *sanitizedDoc.Packages[0]).Return(testErr).Once()
				return packagesSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if package create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0]).Return("", testErr).Once()
				return packagesSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if bundle list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if bundle update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if bundle create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if fetching bundle ids for api fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				return apiSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(testErr).Once()
				return apiSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixApi1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}).Return("", testErr).Once()
				return apiSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api spec delete fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(testErr).Once()
				return specSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api spec create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			productSvcFn:    successfulProductUpdate,
			vendorSvcFn:     successfulVendorUpdate,
			packageSvcFn:    successfulPackageUpdate,
			bundleSvcFn:     successfulBundleUpdate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api1ID).Return("", testErr).Once()
				return specSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if fetching bundle ids for event fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{}, []string{}).Return(testErr).Once()
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}).Return("", testErr).Once()
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec delete fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{}, []string{}).Return(nil).Once()
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if event spec create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixApi1SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectID", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{}, []string{}).Return(nil).Once()
				return eventSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone list fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone update fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), sanitizedDoc.Tombstones[0].OrdID, *sanitizedDoc.Tombstones[0]).Return(testErr).Once()
				return tombstoneSvc
			},
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if tombstone create fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if api resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixApi1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}).Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixApi2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}).Return("", nil).Once()
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
			clientFn: successfulClientFetch,
		},
		{
			Name:            "Does not resync resources if package resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Packages[0]).Return("", nil).Once()
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = packageORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if event resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
			webhookSvcFn:    successfulWebhookList,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], fixEvent1SpecInputs(), []string{bundleID}).Return("", nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], fixEvent2SpecInputs(), []string{bundleID}).Return("", nil).Once()
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = event1ORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if vendor resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), vendorORDID).Return(testErr).Once()
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = vendorORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if product resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
				productSvc.On("Delete", txtest.CtxWithDBMatcher(), productORDID).Return(testErr).Once()
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = productORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name:            "Does not resync resources if bundle resource deletion due to tombstone fails",
			TransactionerFn: secondTransactionNotCommited,
			appSvcFn:        successfulAppList,
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = bundleORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		// TODO: Delete the two tests below after the concept of central registry for Vendors fetching is productive
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn:     successfulAppList,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], fixApi1SpecInputs(), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}).Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], fixApi2SpecInputs(), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}).Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulSAPVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
				return client
			},
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn:       successfulAppList,
			webhookSvcFn:   successfulWebhookList,
			bundleSvcFn:    successfulBundleUpdate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}).Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}).Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecUpdate,
			packageSvcFn: successfulPackageUpdate,
			productSvcFn: successfulProductUpdate,
			vendorSvcFn:  successfulSAPVendorUpdate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), sanitizedDoc.Tombstones[0].OrdID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), baseURL).Return(open_resource_discovery.Documents{doc}, nil)
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
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}

			svc := open_resource_discovery.NewAggregatorService(tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, client)
			err := svc.SyncORDDocuments(context.TODO())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, client)
		})
	}
}
