package ord_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
	testApplicationType                  = "testApplicationType"
	processApplicationFnName             = "ProcessApplication"
	processAppInAppTemplateContextFnName = "ProcessAppInAppTemplateContext"
	processApplicationTemplateFnName     = "ProcessApplicationTemplate"
)

func TestService_Processing(t *testing.T) {
	testErr := errors.New("Test error")
	processingORDDocsErr := errors.New("processing ORD documents")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyORDMapping := application.ORDWebhookMapping{}
	ordMappingWithProxy := application.ORDWebhookMapping{ProxyURL: proxyURL, Type: applicationTypeLabelValue}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: &sync.Map{}}

	sanitizedDoc := fixSanitizedORDDocument()
	sanitizedDocForProxy := fixSanitizedORDDocumentForProxyURL()
	sanitizedStaticDoc := fixSanitizedStaticORDDocument()
	var testSpecData = "{}"
	var testSpec = model.Spec{}
	var nilSpecInput *model.SpecInput
	var nilBundleID *string

	testApplication := fixApplications()[0]
	testResource := ord.Resource{
		Type:          resource.Application,
		ID:            testApplication.ID,
		Name:          testApplication.Name,
		LocalTenantID: testApplication.LocalTenantID,
		ParentID:      &appTemplateID,
	}
	testResourceForAppTemplate := ord.Resource{
		Type: resource.ApplicationTemplate,
		ID:   appTemplateID,
		Name: appTemplateName,
	}
	testWebhookForApplication := fixWebhooksForApplication()[0]
	testWebhookForAppTemplate := fixOrdWebhooksForAppTemplate()[0]
	testStaticWebhookForAppTemplate := fixStaticOrdWebhooksForAppTemplate()[0]

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

	capability1PreSanitizedHash, err := ord.HashObject(fixORDDocument().Capabilities[0])
	require.NoError(t, err)

	capability2PreSanitizedHash, err := ord.HashObject(fixORDDocument().Capabilities[1])
	require.NoError(t, err)

	bundlePreSanitizedHash, err := ord.HashObject(fixORDDocument().ConsumptionBundles[0])
	require.NoError(t, err)

	c := fixORDStaticDocument().ConsumptionBundles[0]
	bundlePreSanitizedHashStaticDoc, err := ord.HashObject(c)
	require.NoError(t, err)

	successfulWebhookConversion := func() *automock.WebhookConverter {
		whConv := &automock.WebhookConverter{}
		whConv.On("InputFromGraphQL", fixTenantMappingWebhookGraphQLInput()).Return(fixTenantMappingWebhookModelInput(), nil).Once()
		return whConv
	}

	successfulWebhookList := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
		return whSvc
	}

	successfulStaticWebhookListAppTemplate := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixStaticOrdWebhooksForAppTemplate(), nil).Once()
		return whSvc
	}

	successfulTenantMappingOnlyCreation := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
		whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
		whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
		whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return([]*model.Webhook{}, nil).Once()
		whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("id", nil).Once()
		return whSvc
	}

	successfulTenantMappingOnlyCreationWithProxyURL := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return([]*model.Webhook{fixWebhookForApplicationWithProxyURL()}, nil).Once()
		return whSvc
	}

	successfulAppTemplateTenantMappingOnlyCreation := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
		whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
		whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
		whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return([]*model.Webhook{}, nil).Once()
		whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("id", nil).Once()
		return whSvc
	}

	successfulTombstoneProcessing := func() *automock.TombstoneProcessor {
		tombstoneProcessor := &automock.TombstoneProcessor{}
		tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixORDDocument().Tombstones).Return(fixTombstones(), nil).Once()
		return tombstoneProcessor
	}

	successfulBundleCreateForApplicationForProxy := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDocForProxy.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleUpdateForApplication := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleUpdateForStaticDoc := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, bundleID, bundleUpdateInputFromCreateInput(*sanitizedStaticDoc.ConsumptionBundles[0]), bundlePreSanitizedHashStaticDoc).Return(nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreateForStaticDoc := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreateWithGenericParam := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulListTwiceAndCreateBundle := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
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

	successfulVendorUpdateForApplication := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID, *sanitizedDoc.Vendors[0]).Return(nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID2, *sanitizedDoc.Vendors[1]).Return(nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulVendorUpdateForStaticDoc := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixVendors(), nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, vendorID, *sanitizedStaticDoc.Vendors[0]).Return(nil).Once()
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, vendorID2, *sanitizedStaticDoc.Vendors[1]).Return(nil).Once()
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulVendorCreate := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[0]).Return("", nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
		vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulVendorCreateForStaticDoc := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Vendors[0]).Return("", nil).Once()
		vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Vendors[1]).Return("", nil).Once()
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixVendors(), nil).Once()
		return vendorSvc
	}

	successfulProductUpdateForApplication := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productID, *sanitizedDoc.Products[0]).Return(nil).Once()
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulProductUpdateForStaticDoc := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixProducts(), nil).Once()
		productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, productID, *sanitizedStaticDoc.Products[0]).Return(nil).Once()
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulProductCreate := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Products[0]).Return("", nil).Once()
		productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulProductCreateForStaticDoc := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Products[0]).Return("", nil).Once()
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixProducts(), nil).Once()
		return productSvc
	}

	successfulPackageCreateForProxy := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDocForProxy.Packages[0], mock.Anything).Return("", nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageUpdateForApplication := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackagesWithHash(), nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageUpdateForStaticDoc := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackagesWithHash(), nil).Once()
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Once()
		packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, packageID, *sanitizedStaticDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageCreate := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		packagesSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", nil).Once()
		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
		return packagesSvc
	}

	successfulPackageCreateForStaticDoc := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		packagesSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Packages[0], mock.Anything).Return("", nil).Once()
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Once()
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

	successfulEmptyCapabilityList := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()

		return capabilitySvc
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

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		return specSvc
	}

	successfulSpecRecreate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulSpecCreateAndUpdateForProxy := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(proxyURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(proxyURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(proxyURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(proxyURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(proxyURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(proxyURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(proxyURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(proxyURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecCreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecCreateAndUpdateForApisAndEvents := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulSpecCreateAndUpdateForStaticDoc := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByIDGlobal", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnlyGlobal", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecWithOneEventCreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecRecreateAndUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecRecreateAndUpdateForApisAndEvents := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulSpecRecreateAndUpdateForStaticDoc := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		capabilitySpec := fixCapabilitySpecInputs()[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.EventSpecReference, event1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.EventSpecReference, event2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.CapabilitySpecReference, capability1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.CapabilitySpecReference, capability2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByIDGlobal", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnlyGlobal", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecRefetch := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest(), fixSuccessfulFetchRequest(), fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(fixAPI2IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{api2spec1ID, api2spec2ID}, model.APISpecReference).Return([]*model.FetchRequest{fixSuccessfulFetchRequest(), fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(fixEvent2IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent2IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(fixCapability1IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixCapability1IDs(), model.CapabilitySpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability2ID).Return(fixCapability2IDs(), nil).Once()
		specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixCapability2IDs(), model.CapabilitySpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Times(5)

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Times(5)

		return specSvc
	}

	successfulAPISpecUpdate := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		return specSvc
	}

	successfulFetchRequestFetchAndUpdateForProxy := func() *automock.FetchRequestService {
		headerMatcher := func() interface{} {
			return mock.MatchedBy(func(headers *sync.Map) bool {
				value, ok := headers.Load("target_host")
				return ok && value == baseURL
			})
		}
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, headerMatcher()).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return fetchReqSvc
	}

	successfulFetchRequestFetch := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdate := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdateForStaticDoc := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		fetchReqSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return fetchReqSvc
	}

	successfulAPICreateAndDeleteForProxy := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDocForProxy.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDocForProxy.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
		return apiSvc
	}

	successfulAPIUpdate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
		apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulAPICreate := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		return apiSvc
	}

	successfulAPICreateAndDelete := func() *automock.APIService {
		apiSvc := &automock.APIService{}
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
		apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
		apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
		return apiSvc
	}

	successfulCapabilityUpdate := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
		capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
		capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability2ID, *sanitizedStaticDoc.Capabilities[1], capability2PreSanitizedHash).Return(nil).Once()
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Twice()
		return capabilitySvc
	}

	successfulCapabilityUpdateForStaticDoc := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilitiesWithHash(), nil).Once()
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilities(), nil).Once()
		capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
		capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, capability2ID, *sanitizedStaticDoc.Capabilities[1], capability2PreSanitizedHash).Return(nil).Once()
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilities(), nil).Once()
		return capabilitySvc
	}

	successfulCapabilityCreate := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return(capability1ID, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[1], ([]*model.SpecInput)(nil), mock.Anything).Return(capability2ID, nil).Once()
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
		return capabilitySvc
	}

	successfulCapabilityCreateForProxy := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDocForProxy.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return(capability1ID, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDocForProxy.Capabilities[1], ([]*model.SpecInput)(nil), mock.Anything).Return(capability2ID, nil).Once()
		capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
		return capabilitySvc
	}

	successfulCapabilityCreateForStaticDoc := func() *automock.CapabilityService {
		capabilitySvc := &automock.CapabilityService{}
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, str.Ptr(packageID), *sanitizedStaticDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return(capability1ID, nil).Once()
		capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, str.Ptr(packageID), *sanitizedStaticDoc.Capabilities[1], ([]*model.SpecInput)(nil), mock.Anything).Return(capability2ID, nil).Once()
		capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilities(), nil).Once()
		return capabilitySvc
	}

	successfulEventCreateForProxy := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDocForProxy.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDocForProxy.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulEventUpdate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Twice()
		return eventSvc
	}

	successfulEventUpdateForStaticDoc := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, event1ID, *sanitizedStaticDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, event2ID, *sanitizedStaticDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Twice()
		return eventSvc
	}

	successfulEventCreate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulEventCreateForStaticDoc := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, nilBundleID, str.Ptr(packageID), *sanitizedStaticDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, nilBundleID, str.Ptr(packageID), *sanitizedStaticDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulOneEventCreate := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
		return eventSvc
	}

	successfulEntityTypeFetchForAppTemplateVersion := func() *automock.EntityTypeService {
		entityTypeSvc := &automock.EntityTypeService{}
		entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEntityTypes(), nil).Once()
		return entityTypeSvc
	}

	successfulEntityTypeFetchForApplication := func() *automock.EntityTypeService {
		entityTypeSvc := &automock.EntityTypeService{}
		entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEntityTypes(), nil).Once()
		return entityTypeSvc
	}

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
		return client
	}

	successfulClientFetchForStaticDoc := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testStaticWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDStaticDocument()}, baseURL, nil)
		return client
	}

	successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy := func() *automock.Client {
		headerMatcher := func() interface{} {
			return mock.MatchedBy(func(ordRequestObject webhook.OpenResourceDiscoveryWebhookRequestObject) bool {
				value, ok := ordRequestObject.Headers.Load("target_host")
				return ok && value == baseURL && ordRequestObject.Application.BaseURL == baseURL
			})
		}

		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, fixWebhookForApplicationWithProxyURL(), ordMappingWithProxy, headerMatcher()).Return(ord.Documents{fixORDDocumentWithoutCredentialExchanges()}, baseURL, nil)
		return client
	}

	successfulAppTemplateVersionList := func() *automock.ApplicationTemplateVersionService {
		svc := &automock.ApplicationTemplateVersionService{}
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Twice()
		return svc
	}

	successfulAppTemplateVersionListAndUpdate := func() *automock.ApplicationTemplateVersionService {
		svc := &automock.ApplicationTemplateVersionService{}
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Twice()
		svc.On("Update", txtest.CtxWithDBMatcher(), appTemplateVersionID, appTemplateID, *fixAppTemplateVersionInput()).Return(nil).Once()
		svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Twice()
		return svc
	}

	successfulAppTemplateVersionForCreation := func() *automock.ApplicationTemplateVersionService {
		svc := &automock.ApplicationTemplateVersionService{}
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
		svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
		svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Twice()
		return svc
	}

	successfulAppTemplateVersionListForAppTemplateFlow := func() *automock.ApplicationTemplateVersionService {
		svc := &automock.ApplicationTemplateVersionService{}
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Times(4)
		return svc
	}

	successfulLabelGetByKey := func() *automock.LabelService {
		svc := &automock.LabelService{}
		svc.On("GetByKey", txtest.CtxWithDBMatcher(), tenantID, model.ApplicationLabelableObject, testApplication.Name, application.ApplicationTypeLabelKey).Return(fixApplicationTypeLabel(), nil).Once()
		return svc
	}

	successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1 := func() *automock.EntityTypeMappingService {
		etmSvc := &automock.EntityTypeMappingService{}
		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), api1ID, resource.API).Return(fixEntityTypeMappings(api1ID, ""), nil).Once()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, entityTypeMappingID).Return(nil).Once()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, api1ID, fixEntityTypeMappingInput(api1ID, "")).Return(entityTypeMappingID, nil).Once()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event1ID, resource.EventDefinition).Return(fixEntityTypeMappings("", event1ID), nil).Once()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, entityTypeMappingID).Return(nil).Once()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, event1ID, fixEntityTypeMappingInput("", event1ID)).Return(entityTypeMappingID, nil).Once()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), api2ID, resource.API).Return(fixEntityTypeMappingsEmpty(), nil).Once()
		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event2ID, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
		return etmSvc
	}

	successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1 := func() *automock.EntityTypeMappingService {
		etmSvc := &automock.EntityTypeMappingService{}
		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), "", resource.API).Return(fixEntityTypeMappingsEmpty(), nil).Twice()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, entityTypeMappingID).Return(nil).Once()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, "", fixEntityTypeMappingInput(api1ID, "")).Return(entityTypeMappingID, nil).Once()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event1ID, resource.EventDefinition).Return(fixEntityTypeMappings("", event1ID), nil).Once()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, entityTypeMappingID).Return(nil).Once()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, event1ID, fixEntityTypeMappingInput("", event1ID)).Return(entityTypeMappingID, nil).Once()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), event2ID, resource.EventDefinition).Return(fixEntityTypeMappingsEmpty(), nil).Once()
		return etmSvc
	}

	successfulEntityTypeMappingSvcUpdateExistingForAPI1AndCreateNewTwiceForEvent1 := func() *automock.EntityTypeMappingService {
		etmSvc := &automock.EntityTypeMappingService{}
		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), api1ID, resource.API).Return(fixEntityTypeMappings(api1ID, ""), nil).Once()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, entityTypeMappingID).Return(nil).Once()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, api1ID, fixEntityTypeMappingInput(api1ID, "")).Return(entityTypeMappingID, nil).Once()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), "", resource.EventDefinition).Return(fixEntityTypeMappings("", event1ID), nil).Twice()
		etmSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, entityTypeMappingID).Return(nil).Twice()
		etmSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, "", fixEntityTypeMappingInput("", event1ID)).Return(entityTypeMappingID, nil).Twice()

		etmSvc.On("ListByOwnerResourceID", txtest.CtxWithDBMatcher(), api2ID, resource.API).Return(fixEntityTypeMappingsEmpty(), nil).Once()

		return etmSvc
	}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn                func() *automock.ApplicationService
		webhookSvcFn            func() *automock.WebhookService
		webhookConvFn           func() *automock.WebhookConverter
		bundleSvcFn             func() *automock.BundleService
		bundleRefSvcFn          func() *automock.BundleReferenceService
		apiSvcFn                func() *automock.APIService
		eventSvcFn              func() *automock.EventService
		entityTypeSvcFn         func() *automock.EntityTypeService
		entityTypeProcessorFn   func() *automock.EntityTypeProcessor
		entityTypeMappingSvcFn  func() *automock.EntityTypeMappingService
		capabilitySvcFn         func() *automock.CapabilityService
		specSvcFn               func() *automock.SpecService
		fetchReqFn              func() *automock.FetchRequestService
		packageSvcFn            func() *automock.PackageService
		productSvcFn            func() *automock.ProductService
		vendorSvcFn             func() *automock.VendorService
		tombstoneProcessorFn    func() *automock.TombstoneProcessor
		tenantSvcFn             func() *automock.TenantService
		globalRegistrySvcFn     func() *automock.GlobalRegistryService
		appTemplateVersionSvcFn func() *automock.ApplicationTemplateVersionService
		appTemplateSvcFn        func() *automock.ApplicationTemplateService
		labelSvcFn              func() *automock.LabelService
		clientFn                func() *automock.Client
		processFnName           string
		webhookMappings         []application.ORDWebhookMapping
		ExpectedErr             error
	}{
		{
			Name: "Success for Application Template webhook with Static ORD data when resources are already in db and APIs/Events last update fields are newer should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			webhookSvcFn:   successfulStaticWebhookListAppTemplate,
			bundleSvcFn:    successfulBundleUpdateForStaticDoc,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api1ID, *sanitizedStaticDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedStaticDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api2ID, *sanitizedStaticDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventUpdateForStaticDoc,
			entityTypeSvcFn: successfulEntityTypeFetchForAppTemplateVersion,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityUpdateForStaticDoc,
			specSvcFn:              successfulSpecRecreateAndUpdateForStaticDoc,
			fetchReqFn:             successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageSvcFn:           successfulPackageUpdateForStaticDoc,
			productSvcFn:           successfulProductUpdateForStaticDoc,
			vendorSvcFn:            successfulVendorUpdateForStaticDoc,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixORDStaticDocument().Tombstones).Return(fixTombstones(), nil).Once()
				return tombstoneProcessor
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionListAndUpdate,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDoc,
			processFnName:           processApplicationTemplateFnName,
		},
		{
			Name: "Success when resources are already in db and APIs/Events last update fields are newer should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDDocument())).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityUpdate,
			specSvcFn:               successfulSpecRecreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageUpdateForApplication,
			productSvcFn:            successfulProductUpdateForApplication,
			vendorSvcFn:             successfulVendorUpdateForApplication,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			processFnName:           processApplicationFnName,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are already in db and APIs/Events last update fields are NOT newer should Update them and refetch only failed API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoNewerLastUpdate(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoNewerLastUpdate(), nil).Twice()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDDocument())).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability2ID, *sanitizedStaticDoc.Capabilities[1], capability2PreSanitizedHash).Return(nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Twice()
				return capabilitySvc
			},
			specSvcFn:               successfulSpecRefetch,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageUpdateForApplication,
			productSvcFn:            successfulProductUpdateForApplication,
			vendorSvcFn:             successfulVendorUpdateForApplication,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			processFnName:           processApplicationFnName,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are not in db should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDDocument())).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			processFnName:           processApplicationFnName,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when webhook has a proxy URL which should be used to access the document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(34)
			},
			appSvcFn:        successfulAppWithBaseURLSvc,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreationWithProxyURL,
			bundleSvcFn:     successfulBundleCreateForApplicationForProxy,
			apiSvcFn:        successfulAPICreateAndDeleteForProxy,
			eventSvcFn:      successfulEventCreateForProxy,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDocForProxy.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:         successfulCapabilityCreateForProxy,
			specSvcFn:               successfulSpecCreateAndUpdateForProxy,
			fetchReqFn:              successfulFetchRequestFetchAndUpdateForProxy,
			packageSvcFn:            successfulPackageCreateForProxy,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy,
			processFnName:           processApplicationFnName,
			webhookMappings:         []application.ORDWebhookMapping{ordMappingWithProxy},
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are not in db should Create them for a Static document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			bundleSvcFn:  successfulBundleCreateForStaticDoc,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}

				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, nilBundleID, str.Ptr(packageID), *sanitizedStaticDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedStaticDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, nilBundleID, str.Ptr(packageID), *sanitizedStaticDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreateForStaticDoc,
			entityTypeSvcFn: successfulEntityTypeFetchForAppTemplateVersion,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulCapabilityCreateForStaticDoc,
			specSvcFn:              successfulSpecCreateAndUpdateForStaticDoc,
			fetchReqFn:             successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageSvcFn:           successfulPackageCreateForStaticDoc,
			productSvcFn:           successfulProductCreateForStaticDoc,
			vendorSvcFn:            successfulVendorCreateForStaticDoc,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixORDStaticDocument().Tombstones).Return(fixTombstones(), nil).Once()
				return tombstoneProcessor
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionForCreation,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDoc,
			processFnName:           processApplicationTemplateFnName,
		},
		{
			Name: "Error when creating Application Template Version based on the doc",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(4, 3)
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return("", testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when getting Application Template from the webhook ObjectID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			appTemplateSvcFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr)
				return svc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when listing Application Template Version by app template ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when fetching the Application Template for the given dynamic doc",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(5)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(nil, testErr)
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when fetching the Application Template for the given dynamic doc for a second time",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundles(), nil).Once()
				return bundlesSvc
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForAppTemplateVersion,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixCapabilities(), nil).Once()
				return capabilitySvc
			},
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackagesWithHash(), nil).Once()
				return packagesSvc
			},

			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(nil, testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when fetching the packages from the DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return packagesSvc
			},

			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when fetching the bundles from the DB",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Once()
				return packagesSvc
			},
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Success when there is ORD webhook on app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn: successfulAppSvc,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
				return tenantSvc
			},
			webhookConvFn:  successfulWebhookConversion,
			webhookSvcFn:   successfulAppTemplateTenantMappingOnlyCreation,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityUpdate,
			specSvcFn:               successfulSpecRecreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageUpdateForApplication,
			productSvcFn:            successfulProductUpdateForApplication,
			vendorSvcFn:             successfulVendorUpdateForApplication,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionListForAppTemplateFlow,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				testResources := ord.Resource{
					Type:     resource.Application,
					ID:       testApplication.ID,
					Name:     testApplication.Name,
					ParentID: &appTemplateID,
				}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResources, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil).Once()
				return client
			},
			processFnName: processAppInAppTemplateContextFnName,
			labelSvcFn:    successfulLabelGetByKey,
		},
		{
			Name: "Error when synchronizing global resources from global registry should get them from DB and proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", mock.Anything).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
				return globalRegistrySvcFn
			},
			clientFn:      successfulClientFetch,
			processFnName: processApplicationFnName,
			labelSvcFn:    successfulLabelGetByKey,
		},
		{
			Name: "Error when synchronizing global resources from global registry and get them from DB should proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				return globalRegistrySvcFn
			},
			clientFn:      successfulClientFetch,
			processFnName: processApplicationFnName,
			labelSvcFn:    successfulLabelGetByKey,
		},
		{
			Name: "Returns error when list for application fails when processing application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(1)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: successfulTenantSvcOnce,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return whSvc
			},
			processFnName: processApplicationFnName,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when list for application template fails when processing application in context of app template",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return whSvc
			},
			processFnName: processAppInAppTemplateContextFnName,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when list for application template fails when processing app template",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return whSvc
			},
			processFnName: processApplicationTemplateFnName,
			ExpectedErr:   testErr,
		},
		{
			Name: "Returns error when get internal tenant id fails when process application webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			processFnName:       processApplicationFnName,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Returns error when get internal tenant id fails before process application webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			processFnName: processApplicationFnName,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when get tenant fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return tenantSvc
			},
			processFnName: processApplicationFnName,
			ExpectedErr:   testErr,
		},
		{
			Name: "Returns error when application locking fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Twice()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
				return tenantSvc
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			processFnName:       processApplicationFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Does not resync resources when event list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources when api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Returns error when get internal tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(2)
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
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			processFnName:           processAppInAppTemplateContextFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Returns error when get tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(2)
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
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			processFnName:           processAppInAppTemplateContextFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Returns error when application locking fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			tenantSvcFn: successfulTenantSvcOnce,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			processFnName:           processAppInAppTemplateContextFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Skips webhook when ORD documents fetch fails",
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
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(nil, "", testErr)
				return client
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			labelSvcFn:          successfulLabelGetByKey,
			processFnName:       processApplicationFnName,
			ExpectedErr:         testErr,
		},
		{
			Name: "Update application local tenant id when ord local id is unique and application does not have local tenant id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalTenantID")}).Return(nil).Once()
				return appSvc
			},
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityUpdate,
			specSvcFn:              successfulSpecRecreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageUpdateForApplication,
			productSvcFn:           successfulProductUpdateForApplication,
			vendorSvcFn:            successfulVendorUpdateForApplication,
			tombstoneProcessorFn:   successfulTombstoneProcessing,
			globalRegistrySvcFn:    successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Fails to update application local tenant id when ord local id is unique and application does not have local tenant id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalTenantID")}).Return(testErr).Once()
				return appSvc
			},
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				return bundlesSvc
			},
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				return capabilitySvc
			},
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackagesWithHash(), nil).Once()
				return packagesSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources for invalid ORD documents when event resource name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(35)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(34)

				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulOneEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecWithOneEventCreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn:            successfulVendorCreate,
			tombstoneProcessorFn:   successfulTombstoneProcessing,
			globalRegistrySvcFn:    successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Resync resources for invalid ORD documents when bundle name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(34)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)

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
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{}, mock.Anything, "").Return("", nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{}, mock.Anything, "").Return(event1ID, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{}, mock.Anything, "").Return(event2ID, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn:            successfulVendorCreate,
			tombstoneProcessorFn:   successfulTombstoneProcessing,
			globalRegistrySvcFn:    successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Resync resources for invalid ORD documents when vendor ordID is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(35)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(34)

				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				return vendorSvc
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			globalRegistrySvcFn:  successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Resync resources for invalid ORD documents when product title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(35)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(34)

				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			bundleSvcFn:     successfulBundleCreate,
			apiSvcFn:        successfulAPICreateAndDelete,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				return productSvc
			},
			vendorSvcFn:          successfulVendorCreate,
			tombstoneProcessorFn: successfulTombstoneProcessing,
			globalRegistrySvcFn:  successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Products[0].Title = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Resync resources for invalid ORD documents when package title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)

				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulWebhookList,
			webhookConvFn: successfulWebhookConversion,
			bundleSvcFn:   successfulEmptyBundleList,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:        successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcCreateNewForAPI1AndUpdateExistingForEvent1,
			capabilitySvcFn:        successfulEmptyCapabilityList,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return packagesSvc
			},
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				return productSvc
			},
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				return vendorSvc
			},
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				return &automock.TombstoneProcessor{}
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Packages[0].Title = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Does not resync resources if vendor list fails",
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
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list vendors after resync",
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
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID, *sanitizedDoc.Vendors[0]).Return(nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID2, *sanitizedDoc.Vendors[1]).Return(nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if vendor update fails",
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
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID, *sanitizedDoc.Vendors[0]).Return(testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if vendor create fails",
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
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[0]).Return("", testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if product list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(12)

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
			vendorSvcFn:  successfulVendorUpdateForApplication,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return productSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list products after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(14)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(14)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(13)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulEmptyBundleList,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productID, *sanitizedDoc.Products[0]).Return(nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				return productSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if product update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(12)

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
			vendorSvcFn:  successfulVendorUpdateForApplication,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productID, *sanitizedDoc.Products[0]).Return(testErr).Once()
				return productSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if product create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(12)

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
			vendorSvcFn:  successfulVendorUpdateForApplication,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Products[0]).Return("", testErr).Once()
				return productSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if package list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(15)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(15)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(14)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn:  successfulEmptyBundleList,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list packages after resync",
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
			bundleSvcFn:  successfulEmptyBundleList,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if package update fails",
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
			bundleSvcFn:  successfulEmptyBundleList,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(testErr).Once()
				return packagesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if package create fails",
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
			bundleSvcFn:  successfulEmptyBundleList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", testErr).Once()
				return packagesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle list fails",
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
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: successfulPackageUpdateForApplication,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list bundles after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(20)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductUpdateForApplication,
			vendorSvcFn:   successfulVendorUpdateForApplication,
			packageSvcFn:  successfulPackageUpdateForApplication,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Twice()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: successfulPackageUpdateForApplication,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(18)
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
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle have different tenant mapping configuration",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithMultipleSameTypesFormat, credentialExchangeStrategyType, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Does not resync resources if webhooks could not be enriched",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(nil, testErr).Once()
				return whSvc
			},
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be listed for application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				return whSvc
			},
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be converted from graphql input to model input",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()

				return whSvc
			},
			webhookConvFn: func() *automock.WebhookConverter {
				whConv := &automock.WebhookConverter{}
				whConv.On("InputFromGraphQL", fixTenantMappingWebhookGraphQLInput()).Return(nil, testErr).Once()
				return whConv
			},
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be created",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("", testErr).Once()

				return whSvc
			},
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn:   successfulVendorCreate,
			packageSvcFn:  successfulPackageCreate,
			bundleSvcFn:   successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources if webhooks can be created successfully",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(36)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(34)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("", nil).Once()

				return whSvc
			},
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn:   successfulVendorCreate,
			packageSvcFn:  successfulPackageCreate,
			bundleSvcFn:   successfulBundleCreateWithGenericParam,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(testErr).Once()
				return apiSvc
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			eventSvcFn:           successfulEventCreate,
			entityTypeSvcFn:      successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not recreate tenant mapping webhooks if there are no differences",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(36)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(34)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(fixTenantMappingWebhooksForApplication(), nil).Once()

				return whSvc
			},
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn:   successfulVendorCreate,
			packageSvcFn:  successfulPackageCreate,
			bundleSvcFn:   successfulBundleCreateWithGenericParam,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(testErr).Once()
				return apiSvc
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			eventSvcFn:           successfulEventCreate,
			entityTypeSvcFn:      successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does recreate of tenant mapping webhooks when there are differences",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(36)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(35)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(35)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				webhooks := fixTenantMappingWebhooksForApplication()
				webhooks[0].URL = str.Ptr("old")
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(webhooks, nil).Once()
				whSvc.On("Delete", txtest.CtxWithDBMatcher(), webhookID, model.ApplicationWebhookReference).Return(nil).Once()
				whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return(webhookID, nil).Once()

				return whSvc
			},
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn:   successfulVendorCreate,
			packageSvcFn:  successfulPackageCreate,
			bundleSvcFn:   successfulBundleCreateWithGenericParam,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			eventSvcFn:           successfulEventCreate,
			entityTypeSvcFn:      successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Does not recreate of tenant mapping webhooks when there are differences but deletion fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(34)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(20)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
				webhooks := fixTenantMappingWebhooksForApplication()
				webhooks[0].URL = str.Ptr("old")
				whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return(webhooks, nil).Once()
				whSvc.On("Delete", txtest.CtxWithDBMatcher(), webhookID, model.ApplicationWebhookReference).Return(testErr).Once()

				return whSvc
			},
			webhookConvFn: successfulWebhookConversion,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn:   successfulVendorCreate,
			packageSvcFn:  successfulPackageCreate,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
				return bundlesSvc
			},
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				return &automock.TombstoneProcessor{}
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			capabilitySvcFn: successfulEmptyCapabilityList,
			specSvcFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: successfulPackageUpdateForApplication,
			bundleSvcFn:  successfulBundleUpdateForApplication,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn:                successfulClientFetch,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list apis after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(25)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(24)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			specSvcFn:      successfulAPISpecUpdate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return apiSvc
			},
			clientFn:                successfulClientFetch,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if fetching bundle ids for api fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: successfulPackageUpdateForApplication,
			bundleSvcFn:  successfulBundleUpdateForApplication,
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
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(testErr).Once()
				return apiSvc
			},
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			packageSvcFn: successfulPackageCreate,
			bundleSvcFn:  successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return("", testErr).Once()
				return apiSvc
			},
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(testErr).Once()
				return specSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, testErr).Once()
				return specSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api spec get fetch request fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(22)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(23)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			productSvcFn:   successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:              successfulEmptyEventList,
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources returns error if api spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(35)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(35)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(35)
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoNewerLastUpdate(), nil).Times(3)
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(fixAPI1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixAPI1IDs(), model.APISpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest(), fixFailedFetchRequest(), fixFailedFetchRequest()}, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(fixAPI2IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, []string{api2spec1ID, api2spec2ID}, model.APISpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest(), fixFailedFetchRequest(), fixFailedFetchRequest()}, nil).Once()

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs(baseURL)[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Twice()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs()))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs()))

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(2 * len(fixCapabilitySpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(2 * len(fixCapabilitySpecInputs()))

				return fetchReqSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()

				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndCreateNewTwiceForEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return("", nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[1], ([]*model.SpecInput)(nil), mock.Anything).Return("", nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()

				return capabilitySvc
			},
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Does not resync resources if event list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(25)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn:      successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list events after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(28)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(29)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], resource.Application, model.EventSpecReference, event1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs(baseURL)[0], resource.Application, model.EventSpecReference, event2ID).Return("", nil, nil).Once()

				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Twice()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if fetching bundle ids for event fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			packageSvcFn: successfulPackageUpdateForApplication,
			bundleSvcFn:  successfulBundleUpdateForApplication,
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
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn:      successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(testErr).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync specification resources if event create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductCreate,
			vendorSvcFn:    successfulVendorCreate,
			packageSvcFn:   successfulPackageCreate,
			bundleSvcFn:    successfulBundleCreate,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:       successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return("", testErr).Once()
				return eventSvc
			},
			entityTypeSvcFn:        successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulEmptyCapabilityList,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)))

				return fetchReqSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			clientFn:                successfulClientFetch,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], resource.Application, model.EventSpecReference, event1ID).Return("", nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoNewerLastUpdate(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event spec get fetch request fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(26)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(27)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoNewerLastUpdate(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			entityTypeSvcFn:         successfulEntityTypeFetchForApplication,
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulEmptyCapabilityList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources returns error if event spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(35)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(35)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(35)
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(fixEvent1IDs(), nil).Once()
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(fixEvent2IDs(), nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent1IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()
				specSvc.On("ListFetchRequestsByReferenceObjectIDs", txtest.CtxWithDBMatcher(), tenantID, fixEvent2IDs(), model.EventSpecReference).Return([]*model.FetchRequest{fixFailedFetchRequest()}, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixAPI1SpecInputs(baseURL)))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))

				return specSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoNewerLastUpdate(), nil).Times(3)
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityUpdate,
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(2 * len(fixCapabilitySpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(2 * len(fixCapabilitySpecInputs()))

				return fetchReqSvc
			},
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Does not resync resources if capability list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(29)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			productSvcFn:    successfulProductUpdateForApplication,
			vendorSvcFn:     successfulVendorUpdateForApplication,
			packageSvcFn:    successfulPackageUpdateForApplication,
			bundleSvcFn:     successfulBundleUpdateForApplication,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn:       successfulSpecRecreateAndUpdateForApisAndEvents,
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list capabilities after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(32)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs(baseURL)[0]

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability1ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, capability2ID).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				return specSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability2ID, *sanitizedStaticDoc.Capabilities[1], capability2PreSanitizedHash).Return(nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Twice()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			productSvcFn:    successfulProductUpdateForApplication,
			vendorSvcFn:     successfulVendorUpdateForApplication,
			packageSvcFn:    successfulPackageUpdateForApplication,
			bundleSvcFn:     successfulBundleUpdateForApplication,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			specSvcFn:       successfulSpecRecreateAndUpdateForApisAndEvents,
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(testErr).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync specification resources if capability create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:        successfulAppGet,
			tenantSvcFn:     successfulTenantSvc,
			webhookSvcFn:    successfulTenantMappingOnlyCreation,
			webhookConvFn:   successfulWebhookConversion,
			productSvcFn:    successfulProductCreate,
			vendorSvcFn:     successfulVendorCreate,
			packageSvcFn:    successfulPackageCreate,
			bundleSvcFn:     successfulBundleCreate,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return("", testErr).Once()
				return capabilitySvc
			},
			specSvcFn: successfulSpecCreateAndUpdateForApisAndEvents,
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				return fetchReqSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			clientFn:                successfulClientFetch,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			productSvcFn:   successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], resource.Application, model.EventSpecReference, event1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs(baseURL)[0], resource.Application, model.EventSpecReference, event2ID).Return("", nil, nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixCapabilitySpecInputs()[0], resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", nil, testErr).Once()

				return specSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return("", nil).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			productSvcFn:   successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], resource.Application, model.EventSpecReference, event1ID).Return("", nil, nil).Once()
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs(baseURL)[0], resource.Application, model.EventSpecReference, event2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(testErr).Once()

				return specSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(30)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			productSvcFn:   successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI1SpecInputs(baseURL)[2], resource.Application, model.APISpecReference, api1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[0], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixAPI2SpecInputs(baseURL)[1], resource.Application, model.APISpecReference, api2ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event1ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent1SpecInputs()[0], resource.Application, model.EventSpecReference, event1ID).Return("", nil, nil).Once()

				specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.EventSpecReference, event2ID).Return(nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *fixEvent2SpecInputs(baseURL)[0], resource.Application, model.EventSpecReference, event2ID).Return("", nil, nil).Once()

				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.CapabilitySpecReference, capability1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilitiesNoNewerLastUpdate(), nil).Twice()
				capabilitySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, capability1ID, *sanitizedStaticDoc.Capabilities[0], capability1PreSanitizedHash).Return(nil).Once()
				return capabilitySvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if tombstone processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(32)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:     successfulVendorUpdateForApplication,
			packageSvcFn:    successfulPackageUpdateForApplication,
			bundleSvcFn:     successfulBundleUpdateForApplication,
			bundleRefSvcFn:  successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn:        successfulAPIUpdate,
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityUpdate,
			specSvcFn:              successfulSpecRecreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixORDDocument().Tombstones).Return(nil, testErr).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if api resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(testErr).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if package resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreate,
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Packages[0], mock.Anything).Return("", nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, packageID).Return(testErr).Once()
				return packagesSvc
			},
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = packageORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = packageORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = packageORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if event resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: successfulAPICreate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[0], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event1ID, nil).Once()
				eventSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.EventResources[1], ([]*model.SpecInput)(nil), []string{bundleID}, mock.Anything, "").Return(event2ID, nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, event1ID).Return(testErr).Once()
				return eventSvc
			},
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn:            successfulVendorCreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixSanitizedORDDocument()
				ts := doc.Tombstones[0]
				ts.OrdID = event1ORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = event1ORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = event1ORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn: func() *automock.CapabilityService {
				capabilitySvc := &automock.CapabilityService{}
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[0], ([]*model.SpecInput)(nil), mock.Anything).Return(capability1ID, nil).Once()
				capabilitySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *sanitizedDoc.Capabilities[1], ([]*model.SpecInput)(nil), mock.Anything).Return(capability2ID, nil).Once()
				capabilitySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixCapabilities(), nil).Once()
				capabilitySvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, capability1ID).Return(testErr).Once()
				return capabilitySvc
			},
			specSvcFn:    successfulSpecCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixSanitizedORDDocument()
				ts := doc.Tombstones[0]
				ts.OrdID = capability1ORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = capability1ORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = capability1ORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if vendor resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[0]).Return("", nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, vendorID).Return(testErr).Once()
				return vendorSvc
			},
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixSanitizedORDDocument()
				ts := doc.Tombstones[0]
				ts.OrdID = vendorORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = vendorORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = vendorORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if product resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Products[0]).Return("", nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, productID).Return(testErr).Once()
				return productSvc
			},
			vendorSvcFn: successfulVendorCreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixSanitizedORDDocument()
				ts := doc.Tombstones[0]
				ts.OrdID = productORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = productORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = productORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(34)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.ConsumptionBundles[0], mock.Anything).Return("", nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, bundleID).Return(testErr).Once()
				return bundlesSvc
			},
			apiSvcFn:        successfulAPICreate,
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn:            successfulVendorCreate,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				doc := fixSanitizedORDDocument()
				ts := doc.Tombstones[0]
				ts.OrdID = bundleORDID
				tombstones := fixTombstones()
				tombstones[0].OrdID = bundleORDID
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Tombstones).Return(tombstones, nil).Once()
				return tombstoneProcessor
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = bundleORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             testErr,
		},
		{
			Name: "Returns error when failing to open final transaction to commit fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(34)
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn:  successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:         successfulCapabilityCreate,
			specSvcFn:               successfulSpecCreate,
			fetchReqFn:              successfulFetchRequestFetch,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Returns error when failing to find spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(34)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs(baseURL)[0]

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Twice()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(nil, testErr).Once()

				return specSvc
			},
			fetchReqFn:              successfulFetchRequestFetch,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Returns error when failing to update spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(34)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs(baseURL)[0]

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Twice()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(testErr).Once()

				return specSvc
			},
			fetchReqFn:              successfulFetchRequestFetch,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Returns error when failing to update fetch request in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(34)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
				api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
				api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

				api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
				api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

				event1Spec := fixEvent1SpecInputs()[0]
				event2Spec := fixEvent2SpecInputs(baseURL)[0]

				capabilitySpec := fixCapabilitySpecInputs()[0]

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *capabilitySpec, resource.Application, model.CapabilitySpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.CapabilitySpecFetchRequestReference, ""), nil).Twice()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Once()

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, &sync.Map{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs()))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(testErr).Once()

				return fetchReqSvc
			},
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneProcessorFn:    successfulTombstoneProcessing,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			ExpectedErr:             processingORDDocsErr,
		},
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[0], ([]*model.SpecInput)(nil), map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, mock.Anything, "").Return(api1ID, nil).Once()
				apiSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, nilBundleID, str.Ptr(packageID), *sanitizedDoc.APIResources[1], ([]*model.SpecInput)(nil), map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, mock.Anything, "").Return(api2ID, nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventCreate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityCreate,
			specSvcFn:              successfulSpecCreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageCreate,
			productSvcFn:           successfulProductCreate,
			vendorSvcFn:            successfulEmptyVendorList,
			tombstoneProcessorFn:   successfulTombstoneProcessing,
			globalRegistrySvcFn:    successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn:      successfulEventUpdate,
			entityTypeSvcFn: successfulEntityTypeFetchForApplication,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				enityTypeProcessor := &automock.EntityTypeProcessor{}
				enityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return enityTypeProcessor
			},
			entityTypeMappingSvcFn: successfulEntityTypeMappingSvcUpdateExistingForAPI1AndEvent1,
			capabilitySvcFn:        successfulCapabilityUpdate,
			specSvcFn:              successfulSpecRecreateAndUpdate,
			fetchReqFn:             successfulFetchRequestFetchAndUpdate,
			packageSvcFn:           successfulPackageUpdateForApplication,
			productSvcFn:           successfulProductUpdateForApplication,
			vendorSvcFn:            successfulEmptyVendorList,
			tombstoneProcessorFn:   successfulTombstoneProcessing,
			globalRegistrySvcFn:    successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
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
			entityTypeSvc := &automock.EntityTypeService{}
			if test.entityTypeSvcFn != nil {
				entityTypeSvc = test.entityTypeSvcFn()
			}
			entityTypeProcessor := &automock.EntityTypeProcessor{}
			if test.entityTypeProcessorFn != nil {
				entityTypeProcessor = test.entityTypeProcessorFn()
			}
			entityTypeMappingSvc := &automock.EntityTypeMappingService{}
			if test.entityTypeMappingSvcFn != nil {
				entityTypeMappingSvc = test.entityTypeMappingSvcFn()
			}
			capabilitySvc := &automock.CapabilityService{}
			if test.capabilitySvcFn != nil {
				capabilitySvc = test.capabilitySvcFn()
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
			tombstoneProcessor := &automock.TombstoneProcessor{}
			if test.tombstoneProcessorFn != nil {
				tombstoneProcessor = test.tombstoneProcessorFn()
			}
			tenantSvc := &automock.TenantService{}
			if test.tenantSvcFn != nil {
				tenantSvc = test.tenantSvcFn()
			}
			globalRegistrySvcFn := &automock.GlobalRegistryService{}
			if test.globalRegistrySvcFn != nil {
				globalRegistrySvcFn = test.globalRegistrySvcFn()
			}
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}
			whConverter := &automock.WebhookConverter{}
			if test.webhookConvFn != nil {
				whConverter = test.webhookConvFn()
			}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			if test.appTemplateVersionSvcFn != nil {
				appTemplateVersionSvc = test.appTemplateVersionSvcFn()
			}
			appTemplateSvc := &automock.ApplicationTemplateService{}
			if test.appTemplateSvcFn != nil {
				appTemplateSvc = test.appTemplateSvcFn()
			}
			labelSvc := &automock.LabelService{}
			if test.labelSvcFn != nil {
				labelSvc = test.labelSvcFn()
			}
			ordWebhookMappings := []application.ORDWebhookMapping{}
			if test.webhookMappings != nil {
				ordWebhookMappings = test.webhookMappings
			}

			metrixCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, entityTypeSvc, entityTypeProcessor, entityTypeMappingSvc, capabilitySvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, ordWebhookMappings, nil)

			var err error
			switch test.processFnName {
			case processApplicationFnName:
				err = svc.ProcessApplication(context.TODO(), appID)
			case processAppInAppTemplateContextFnName:
				err = svc.ProcessAppInAppTemplateContext(context.TODO(), appTemplateID, appID)
			case processApplicationTemplateFnName:
				err = svc.ProcessApplicationTemplate(context.TODO(), appTemplateID)
			}
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, capabilitySvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func TestService_ProcessApplication(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyORDMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: &sync.Map{}}

	testApplication := fixApplications()[0]
	testResource := ord.Resource{
		Type:          resource.Application,
		ID:            testApplication.ID,
		Name:          testApplication.Name,
		LocalTenantID: testApplication.LocalTenantID,
		ParentID:      &appTemplateID,
	}
	testWebhookForApplication := fixWebhooksForApplication()[0]

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil)
		return client
	}

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn            func() *automock.ApplicationService
		webhookSvcFn        func() *automock.WebhookService
		webhookConvFn       func() *automock.WebhookConverter
		tenantSvcFn         func() *automock.TenantService
		globalRegistrySvcFn func() *automock.GlobalRegistryService
		labelSvcFn          func() *automock.LabelService
		clientFn            func() *automock.Client
		appID               string
		ExpectedErr         error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			appSvcFn: successfulAppGet,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Times(3)
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Times(3)
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetch,
			appID:               appID,
			labelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), tenantID, model.ApplicationLabelableObject, testApplication.Name, application.ApplicationTypeLabelKey).Return(fixApplicationTypeLabel(), nil).Once()
				return svc
			},
		},
		{
			Name: "Error while listing webhooks for application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: successfulTenantSvcOnce,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return whSvc
			},
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				return &automock.GlobalRegistryService{}
			},
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:       appID,
			ExpectedErr: testErr,
		},
		{
			Name: "Error while retrieving application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:       appID,
			ExpectedErr: testErr,
		},
		{
			Name: "Error while getting lowest owner of resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Once()
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return("", testErr).Once()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:       appID,
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

			whConverter := &automock.WebhookConverter{}
			if test.webhookConvFn != nil {
				whConverter = test.webhookConvFn()
			}

			labelSvc := &automock.LabelService{}
			if test.labelSvcFn != nil {
				labelSvc = test.labelSvcFn()
			}

			tenantSvc := &automock.TenantService{}
			if test.tenantSvcFn != nil {
				tenantSvc = test.tenantSvcFn()
			}

			globalRegistrySvcFn := &automock.GlobalRegistryService{}
			if test.globalRegistrySvcFn != nil {
				globalRegistrySvcFn = test.globalRegistrySvcFn()
			}

			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}

			bndlSvc := &automock.BundleService{}
			bndlRefSvc := &automock.BundleReferenceService{}
			apiSvc := &automock.APIService{}
			eventSvc := &automock.EventService{}
			entityTypeSvc := &automock.EntityTypeService{}
			capabilitySvc := &automock.CapabilityService{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneProcessor := &automock.TombstoneProcessor{}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			appTemplateSvc := &automock.ApplicationTemplateService{}

			metrixCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, entityTypeSvc, nil, nil, capabilitySvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{}, nil)
			err := svc.ProcessApplication(context.TODO(), test.appID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func TestService_ProcessApplicationTemplate(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyORDMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: &sync.Map{}}

	testResourceAppTemplate := ord.Resource{
		Type: resource.ApplicationTemplate,
		ID:   appTemplateID,
		Name: appTemplateName,
	}
	testWebhookForAppTemplate := fixStaticOrdWebhooksForAppTemplate()[0]

	successfulClientFetchForAppTemplate := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceAppTemplate, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil).Once()
		return client
	}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn                func() *automock.ApplicationService
		webhookSvcFn            func() *automock.WebhookService
		webhookConvFn           func() *automock.WebhookConverter
		tenantSvcFn             func() *automock.TenantService
		globalRegistrySvcFn     func() *automock.GlobalRegistryService
		appTemplateSvcFn        func() *automock.ApplicationTemplateService
		appTemplateVersionSvcFn func() *automock.ApplicationTemplateVersionService
		labelSvcFn              func() *automock.LabelService
		clientFn                func() *automock.Client
		appTemplateID           string
		ExpectedErr             error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn: func() *automock.ApplicationService {
				return &automock.ApplicationService{}
			},
			tenantSvcFn: func() *automock.TenantService {
				return &automock.TenantService{}
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixStaticOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForAppTemplate,
			appTemplateID:       appTemplateID,
			appTemplateSvcFn:    successAppTemplateGetSvc,
			labelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
		},
		{
			Name: "Error while listing webhooks for application templates",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
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
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				return &automock.GlobalRegistryService{}
			},
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appTemplateID: appTemplateID,
			ExpectedErr:   testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			bndlSvc := &automock.BundleService{}
			bndlRefSvc := &automock.BundleReferenceService{}
			apiSvc := &automock.APIService{}
			eventSvc := &automock.EventService{}
			entityTypeSvc := &automock.EntityTypeService{}
			capabilitySvc := &automock.CapabilityService{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneProcessor := &automock.TombstoneProcessor{}

			appSvc := &automock.ApplicationService{}
			if test.appSvcFn != nil {
				appSvc = test.appSvcFn()
			}
			whSvc := &automock.WebhookService{}
			if test.webhookSvcFn != nil {
				whSvc = test.webhookSvcFn()
			}
			whConverter := &automock.WebhookConverter{}
			if test.webhookConvFn != nil {
				whConverter = test.webhookConvFn()
			}
			tenantSvc := &automock.TenantService{}
			if test.tenantSvcFn != nil {
				tenantSvc = test.tenantSvcFn()
			}
			globalRegistrySvcFn := &automock.GlobalRegistryService{}
			if test.globalRegistrySvcFn != nil {
				globalRegistrySvcFn = test.globalRegistrySvcFn()
			}
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}
			appTemplateSvc := &automock.ApplicationTemplateService{}
			if test.appTemplateSvcFn != nil {
				appTemplateSvc = test.appTemplateSvcFn()
			}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			if test.appTemplateVersionSvcFn != nil {
				appTemplateVersionSvc = test.appTemplateVersionSvcFn()
			}
			labelSvc := &automock.LabelService{}
			if test.labelSvcFn != nil {
				labelSvc = test.labelSvcFn()
			}
			metricsCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metricsCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, entityTypeSvc, nil, nil, capabilitySvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{}, nil)
			err := svc.ProcessApplicationTemplate(context.TODO(), test.appTemplateID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func TestService_ProcessAppInAppTemplateContext(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyORDMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: &sync.Map{}}

	testApplication := fixApplications()[0]
	testResourceApp := ord.Resource{
		Type:          resource.Application,
		ID:            testApplication.ID,
		Name:          testApplication.Name,
		LocalTenantID: testApplication.LocalTenantID,
		ParentID:      &appTemplateID,
	}
	testWebhookForAppTemplate := fixOrdWebhooksForAppTemplate()[0]

	successfulClientFetchForAppTemplate := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceApp, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil).Once()
		return client
	}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn                func() *automock.ApplicationService
		webhookSvcFn            func() *automock.WebhookService
		webhookConvFn           func() *automock.WebhookConverter
		tenantSvcFn             func() *automock.TenantService
		globalRegistrySvcFn     func() *automock.GlobalRegistryService
		appTemplateSvcFn        func() *automock.ApplicationTemplateService
		appTemplateVersionSvcFn func() *automock.ApplicationTemplateVersionService
		labelSvcFn              func() *automock.LabelService
		clientFn                func() *automock.Client
		appID                   string
		appTemplateID           string
		ExpectedErr             error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			appSvcFn: successfulAppSvc,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
				return tenantSvc
			},
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForAppTemplate,
			appID:               appID,
			appTemplateID:       appTemplateID,
			appTemplateSvcFn:    successAppTemplateGetSvc,
			labelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), tenantID, model.ApplicationLabelableObject, testApplication.Name, application.ApplicationTypeLabelKey).Return(fixApplicationTypeLabel(), nil).Once()
				return svc
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
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				return &automock.GlobalRegistryService{}
			},
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:         appID,
			appTemplateID: appTemplateID,
			ExpectedErr:   testErr,
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:            appID,
			appTemplateID:    appTemplateID,
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
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
			tenantSvcFn: successfulTenantSvcOnce,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:            appID,
			appTemplateID:    appTemplateID,
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:            appID,
			appTemplateID:    appTemplateID,
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
		},
		{
			Name: "Error when cannot find application from the given app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.Application{}, nil).Once()
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				return &automock.Client{}
			},
			appID:            appID,
			appTemplateID:    appTemplateID,
			appTemplateSvcFn: successAppTemplateGetSvc,
			labelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ExpectedErr: errors.New("cannot find application"),
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			bndlSvc := &automock.BundleService{}
			bndlRefSvc := &automock.BundleReferenceService{}
			apiSvc := &automock.APIService{}
			eventSvc := &automock.EventService{}
			entityTypeScv := &automock.EntityTypeService{}
			capabilitySvc := &automock.CapabilityService{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneProcessor := &automock.TombstoneProcessor{}

			appSvc := &automock.ApplicationService{}
			if test.appSvcFn != nil {
				appSvc = test.appSvcFn()
			}
			whSvc := &automock.WebhookService{}
			if test.webhookSvcFn != nil {
				whSvc = test.webhookSvcFn()
			}
			whConverter := &automock.WebhookConverter{}
			if test.webhookConvFn != nil {
				whConverter = test.webhookConvFn()
			}
			tenantSvc := &automock.TenantService{}
			if test.tenantSvcFn != nil {
				tenantSvc = test.tenantSvcFn()
			}
			globalRegistrySvcFn := &automock.GlobalRegistryService{}
			if test.globalRegistrySvcFn != nil {
				globalRegistrySvcFn = test.globalRegistrySvcFn()
			}
			client := &automock.Client{}
			if test.clientFn != nil {
				client = test.clientFn()
			}
			appTemplateSvc := &automock.ApplicationTemplateService{}
			if test.appTemplateSvcFn != nil {
				appTemplateSvc = test.appTemplateSvcFn()
			}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			if test.appTemplateVersionSvcFn != nil {
				appTemplateVersionSvc = test.appTemplateVersionSvcFn()
			}
			labelSvc := &automock.LabelService{}
			if test.labelSvcFn != nil {
				labelSvc = test.labelSvcFn()
			}
			metrixCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, entityTypeScv, nil, nil, capabilitySvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{}, nil)
			err := svc.ProcessAppInAppTemplateContext(context.TODO(), test.appTemplateID, test.appID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func successfulGlobalRegistrySvc() *automock.GlobalRegistryService {
	globalRegistrySvcFn := &automock.GlobalRegistryService{}
	globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
	return globalRegistrySvcFn
}

func successfulTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Times(3)
	tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Times(3)
	return tenantSvc
}

func successfulTenantSvcOnce() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Once()
	tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Once()
	return tenantSvc
}

func successfulAppGet() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
	return appSvc
}

func successfulAppGetOnce() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Once()
	return appSvc
}

func successfulAppSvc() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixApplications(), nil).Once()
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()

	return appSvc
}

func successfulAppWithBaseURLSvc() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplicationsWithBaseURL()[0], nil).Twice()

	return appSvc
}

func successAppTemplateGetSvc() *automock.ApplicationTemplateService {
	svc := &automock.ApplicationTemplateService{}
	svc.On("Get", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplate(), nil)
	return svc
}
