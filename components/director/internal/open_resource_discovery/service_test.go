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
	var testSpecData = "{}"
	var testSpec = model.Spec{}
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

	bundlePreSanitizedHash, err := ord.HashObject(fixORDDocument().ConsumptionBundles[0])
	require.NoError(t, err)

	successfulWebhookList := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
		return whSvc
	}

	successfulBundleUpdate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
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

	successfulEmptyBundleList := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()

		return bundlesSvc
	}

	successfulEmptyVendorList := func() *automock.VendorService {
		vendorService := &automock.VendorService{}
		vendorService.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()

		return vendorService
	}

	successfulSpecCreate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulSpecRecreate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs()[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulSpecCreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		return specSvc
	}

	successfulSpecWithOneEventCreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		event2Spec := fixEvent2SpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		return specSvc
	}

	successfulSpecRecreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs()[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		return specSvc
	}

	successfulSpecRefetch := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest(), fixSuccessfulFetchRequest(), fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(fixAPI2IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{api2spec1ID, api2spec2ID}, model.APISpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest(), fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(fixEvent2IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent2IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Times(3)

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Times(3)

		return specSvc
	}

	successfulAPISpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs()[0]
		api1SpecInput2 := fixAPI1SpecInputs()[1]
		api1SpecInput3 := fixAPI1SpecInputs()[2]

		api2SpecInput1 := fixAPI2SpecInputs()[0]
		api2SpecInput2 := fixAPI2SpecInputs()[1]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulFetchRequestFetch := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs()) + len(fixAPI2SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdate := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs())) // len(fixAPI2SpecInputs()) is excluded because it's API is part of tombstones

		return fetchReqSvc
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
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
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
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulOneEventCreate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
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
				return txGen.ThatSucceedsMultipleTimes(29)
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
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(29)
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
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(29)
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(30)
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
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(29)
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(29)
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
		},
		{
			Name: "Resync resources for invalid ORD documents when event resource name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(28)

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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulOneEventCreate,
			specSvcFn:    successfulSpecWithOneEventCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Resync resources for invalid ORD documents when bundle name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(28)

				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Times(3)
				return bundlesSvc
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{}, mock.Anything, "").Return(event1ID, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{}, mock.Anything, "").Return(event2ID, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Resync resources for invalid ORD documents when vendor ordID is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(28)

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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				return vendorSvc
			},
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
				doc.Vendors[0].OrdID = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Resync resources for invalid ORD documents when product title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(28)

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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				return productSvc
			},
			vendorSvcFn: successfulVendorCreate,
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
				doc.Products[0].Title = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
		},
		{
			Name: "Resync resources for invalid ORD documents when package title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(24)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(24)

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
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Times(3)
				return eventSvc
			},
			fetchReqFn: successfulFetchRequestFetchAndUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Times(3)
				return packagesSvc
			},
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Packages[0].Title = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
			bundleSvcFn:  successfulEmptyBundleList,
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Twice()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", testErr).Once()
				return bundlesSvc
			},
			clientFn:          successfulClientFetch,
			apiSvcFn:          successfulEmptyAPIList,
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", testErr).Once()
				return apiSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, testErr).Once()
				return specSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			eventSvcFn:        successfulEmptyEventList,
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
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
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
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:        successfulEmptyEventList,
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
		},
		{
			Name: "Resync resources returns error if api spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(29)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(29)
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
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Times(3)
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest(), fixFailedFetchRequest(), fixFailedFetchRequest()}, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(fixAPI2IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{api2spec1ID, api2spec2ID}, model.APISpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest(), fixFailedFetchRequest(), fixFailedFetchRequest()}, nil).Once()

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixAPI1SpecInputs()))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixAPI1SpecInputs()))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				return fetchReqSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()

				return eventSvc
			},
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs()[0], model.EventSpecReference, event2ID).Return("", nil, nil).Once()

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
		},
		{
			Name: "Does not resync specification resources if event create fails",
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
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", testErr).Once()
				return eventSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs()[0]
				api1SpecInput2 := fixAPI1SpecInputs()[1]
				api1SpecInput3 := fixAPI1SpecInputs()[2]

				api2SpecInput1 := fixAPI2SpecInputs()[0]
				api2SpecInput2 := fixAPI2SpecInputs()[1]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs()) + len(fixAPI2SpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs()) + len(fixAPI2SpecInputs()))

				return fetchReqSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil, nil).Once()
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], model.EventSpecReference, event1ID).Return("", nil, testErr).Once()
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(nil, testErr).Once()
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
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return(nil, testErr).Once()
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
			Name: "Resync resources returns error if event spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(29)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(29)
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
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs()[0]
				api1SpecInput2 := fixAPI1SpecInputs()[1]
				api1SpecInput3 := fixAPI1SpecInputs()[2]

				api2SpecInput1 := fixAPI2SpecInputs()[0]
				api2SpecInput2 := fixAPI2SpecInputs()[1]

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[0], model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[1], model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs()[2], model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[0], model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs()[1], model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, event2ID).Return(fixEvent2IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent2IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixAPI1SpecInputs()))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixAPI1SpecInputs()))

				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Times(3)
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs()))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs()))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				return fetchReqSvc
			},
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
			specSvcFn:      successfulSpecRecreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
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
			specSvcFn:      successfulSpecRecreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
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
			specSvcFn:      successfulSpecRecreate,
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
			specSvcFn:    successfulSpecCreate,
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(testErr).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
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
			specSvcFn:    successfulSpecCreate,
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
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("Delete", txtest.CtxWithDBMatcher(), event1ID).Return(testErr).Once()
				return eventSvc
			},
			specSvcFn:    successfulSpecCreate,
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
			specSvcFn:    successfulSpecCreate,
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
			specSvcFn:    successfulSpecCreate,
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("Delete", txtest.CtxWithDBMatcher(), bundleID).Return(testErr).Once()
				return bundlesSvc
			},
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
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
		},
		{
			Name: "Returns error when failing to open final transaction to commit fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(28)
				transact.On("Begin").Return(persistTx, testErr).Once()
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
			fetchReqFn:   successfulFetchRequestFetch,
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
			Name: "Returns error when failing to find spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(28)
				transact.On("Begin").Return(persistTx, nil).Once()
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs()[0]
				api1SpecInput2 := fixAPI1SpecInputs()[1]
				api1SpecInput3 := fixAPI1SpecInputs()[2]

				api2SpecInput1 := fixAPI2SpecInputs()[0]
				api2SpecInput2 := fixAPI2SpecInputs()[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(nil, testErr).Once()

				return specSvc
			},
			fetchReqFn:   successfulFetchRequestFetch,
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
			Name: "Returns error when failing to update spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(28)
				transact.On("Begin").Return(persistTx, nil).Once()
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs()[0]
				api1SpecInput2 := fixAPI1SpecInputs()[1]
				api1SpecInput3 := fixAPI1SpecInputs()[2]

				api2SpecInput1 := fixAPI2SpecInputs()[0]
				api2SpecInput2 := fixAPI2SpecInputs()[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(testErr).Once()

				return specSvc
			},
			fetchReqFn:   successfulFetchRequestFetch,
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
			Name: "Returns error when failing to update fetch request in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(28)
				transact.On("Begin").Return(persistTx, nil).Once()
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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs()[0]
				api1SpecInput2 := fixAPI1SpecInputs()[1]
				api1SpecInput3 := fixAPI1SpecInputs()[2]

				api2SpecInput1 := fixAPI2SpecInputs()[0]
				api2SpecInput2 := fixAPI2SpecInputs()[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Once()

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs()) + len(fixAPI2SpecInputs()) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(testErr).Once()

				return fetchReqSvc
			},
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
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(27)
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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
				return txGen.ThatSucceedsMultipleTimes(27)
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
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
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

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
			err := svc.SyncORDDocuments(context.TODO(), ord.MetricsConfig{})
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

func TestService_ProcessApplications(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testApplication := fixApplications()[0]
	testWebhookForApplication := fixWebhooksForApplication()[0]

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForApplication).Return(ord.Documents{}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn          func() *automock.ApplicationService
		webhookSvcFn      func() *automock.WebhookService
		tenantSvcFn       func() *automock.TenantService
		globalRegistrySvc func() *automock.GlobalRegistryService
		clientFn          func() *automock.Client
		appIDs            func() []string
		ExpectedErr       error
	}{
		{
			Name: "Success when empty app IDs array",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntStartTransaction()
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			globalRegistrySvc: func() *automock.GlobalRegistryService {
				return &automock.GlobalRegistryService{}
			},
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appIDs: func() []string {
				return []string{}
			},
		},
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetch,
			appIDs: func() []string {
				return []string{appID}
			},
		},
		{
			Name: "Error while listing webhooks for application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(1)
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appIDs: func() []string {
				return []string{appID}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while retrieving application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appIDs: func() []string {
				return []string{appID}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while getting lowest owner of resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appIDs: func() []string {
				return []string{appID}
			},
			ExpectedErr: testErr,
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
			bndlRefSvc := &automock.BundleReferenceService{}
			apiSvc := &automock.APIService{}
			eventSvc := &automock.EventService{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneSvc := &automock.TombstoneService{}
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

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
			err := svc.ProcessApplications(context.TODO(), ord.MetricsConfig{}, test.appIDs())
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

func TestService_ProcessApplicationTemplates(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testApplication := fixApplications()[0]
	testWebhookForAppTemplate := fixOrdWebhooksForAppTemplate()[0]

	successfulClientFetchForAppTemplate := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testApplication, testWebhookForAppTemplate).Return(ord.Documents{}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn          func() *automock.ApplicationService
		webhookSvcFn      func() *automock.WebhookService
		tenantSvcFn       func() *automock.TenantService
		globalRegistrySvc func() *automock.GlobalRegistryService
		clientFn          func() *automock.Client
		appTemplateIDs    func() []string
		ExpectedErr       error
	}{
		{
			Name: "Success when empty application template IDs array",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntStartTransaction()
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			globalRegistrySvc: func() *automock.GlobalRegistryService {
				return &automock.GlobalRegistryService{}
			},
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateIDs: func() []string {
				return []string{}
			},
		},
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			appSvcFn:    successfulAppTemplateAppSvc,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn:          successfulClientFetchForAppTemplate,
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
		},
		{
			Name: "Error while listing webhooks for application templates",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(1)
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while listing applications by application template id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return appSvc
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while getting application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
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
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while getting lowest owner of resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
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
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvc: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			ExpectedErr: testErr,
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
			bndlRefSvc := &automock.BundleReferenceService{}
			apiSvc := &automock.APIService{}
			eventSvc := &automock.EventService{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneSvc := &automock.TombstoneService{}
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

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, client)
			err := svc.ProcessApplicationTemplates(context.TODO(), ord.MetricsConfig{}, test.appTemplateIDs())
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

func successfulGlobalRegistrySvc() *automock.GlobalRegistryService {
	globalRegistrySvc := &automock.GlobalRegistryService{}
	globalRegistrySvc.On("SyncGlobalResources", context.TODO()).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
	return globalRegistrySvc
}

func successfulTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
	tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Once()
	return tenantSvc
}

func successfulAppGet() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Once()
	return appSvc
}

func successfulAppTemplateAppSvc() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Once()

	return appSvc
}
