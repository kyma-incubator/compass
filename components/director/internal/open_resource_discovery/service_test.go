package ord_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"net/http"
	"testing"

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
	testApplicationType = "testApplicationType"
)

func TestService_SyncORDDocuments(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	ordMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: http.Header{}}

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
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
		return whSvc
	}

	successfulWebhookListAppTemplate := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
		return whSvc
	}

	successfulTenantMappingOnlyCreation := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
		whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
		whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return([]*model.Webhook{}, nil).Once()
		whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("id", nil).Once()
		return whSvc
	}

	successfulTenantMappingOnlyCreationWithProxyURL := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return([]*model.Webhook{fixWebhookForApplicationWithProxyURL()}, nil).Once()
		return whSvc
	}

	successfulAppTemplateTenantMappingOnlyCreation := func() *automock.WebhookService {
		whSvc := &automock.WebhookService{}
		whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
		whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
		whSvc.On("EnrichWebhooksWithTenantMappingWebhooks", whInputs).Return(whInputs, nil).Once()
		whSvc.On("ListForApplicationGlobal", txtest.CtxWithDBMatcher(), appID).Return([]*model.Webhook{}, nil).Once()
		whSvc.On("Create", txtest.CtxWithDBMatcher(), appID, *fixTenantMappingWebhookModelInput(), model.ApplicationWebhookReference).Return("id", nil).Once()
		return whSvc
	}

	successfulTombstoneCreate := func() *automock.TombstoneService {
		tombstoneSvc := &automock.TombstoneService{}
		tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Tombstones[0]).Return("", nil).Once()
		tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
		return tombstoneSvc
	}

	successfulTombstoneCreateForStaticDoc := func() *automock.TombstoneService {
		tombstoneSvc := &automock.TombstoneService{}
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *sanitizedStaticDoc.Tombstones[0]).Return("", nil).Once()
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
		return tombstoneSvc
	}

	successfulTombstoneUpdateForStaticDoc := func() *automock.TombstoneService {
		tombstoneSvc := &automock.TombstoneService{}
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
		tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, tombstoneID, *sanitizedStaticDoc.Tombstones[0]).Return(nil).Once()
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Once()
		return tombstoneSvc
	}

	successfulTombstoneUpdateForStaticDocWithApplication := func() *automock.TombstoneService {
		tombstoneSvc := &automock.TombstoneService{}
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Times(2)
		tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, tombstoneID, *sanitizedStaticDoc.Tombstones[0]).Return(nil).Times(2)
		tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixTombstones(), nil).Times(2)
		return tombstoneSvc
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

	successfulBundleUpdateForStaticDocWithApplication := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, bundleID, bundleUpdateInputFromCreateInput(*sanitizedStaticDoc.ConsumptionBundles[0]), bundlePreSanitizedHashStaticDoc).Return(nil).Times(2)
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Times(4)

		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
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

	successfulVendorUpdateForStaticDocWithApplication := func() *automock.VendorService {
		vendorSvc := &automock.VendorService{}
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixVendors(), nil).Times(2)
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, vendorID, *sanitizedStaticDoc.Vendors[0]).Return(nil).Times(2)
		vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, vendorID2, *sanitizedStaticDoc.Vendors[1]).Return(nil).Times(2)
		vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixVendors(), nil).Times(2)
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

	successfulProductUpdateForStaticDocWithApplication := func() *automock.ProductService {
		productSvc := &automock.ProductService{}
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixProducts(), nil).Times(2)
		productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, productID, *sanitizedStaticDoc.Products[0]).Return(nil).Times(2)
		productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixProducts(), nil).Times(2)
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

	successfulPackageUpdateForStaticDocWithApplication := func() *automock.PackageService {
		packagesSvc := &automock.PackageService{}
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackagesWithHash(), nil).Once()
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Once()
		packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, packageID, *sanitizedStaticDoc.Packages[0], packagePreSanitizedHash).Return(nil).Times(2)
		packagesSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixPackages(), nil).Times(4)

		packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackagesWithHash(), nil).Once()
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

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

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

	successfulSpecCreateAndUpdateForProxy := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(customWebhookConfigURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(customWebhookConfigURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(customWebhookConfigURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(customWebhookConfigURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(customWebhookConfigURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(customWebhookConfigURL)[0]

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(customWebhookConfigURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(customWebhookConfigURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByIDGlobal", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnlyGlobal", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.Application, model.APISpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Once()

		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.Application, model.EventSpecReference, mock.Anything).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Once()

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("GetByIDGlobal", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnlyGlobal", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		return specSvc
	}

	successfulSpecRecreateAndUpdateForStaticDocWithApplication := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

		api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
		api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
		api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

		api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
		api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

		event1Spec := fixEvent1SpecInputs()[0]
		event2Spec := fixEvent2SpecInputs(baseURL)[0]

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return(nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api1SpecInput3, resource.ApplicationTemplateVersion, model.APISpecReference, api1ID).Return("", fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Times(2)
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return(nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput1, resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *api2SpecInput2, resource.ApplicationTemplateVersion, model.APISpecReference, api2ID).Return("", fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, ""), nil).Times(2)

		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.EventSpecReference, event1ID).Return(nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event1Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, event1ID).Return("", fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Times(2)
		specSvc.On("DeleteByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, model.EventSpecReference, event2ID).Return(nil).Times(2)
		specSvc.On("CreateByReferenceObjectIDWithDelayedFetchRequest", txtest.CtxWithDBMatcher(), *event2Spec, resource.ApplicationTemplateVersion, model.EventSpecReference, event2ID).Return("", fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, ""), nil).Times(2)

		specSvc.On("GetByIDGlobal", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times((len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) * 2) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnlyGlobal", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times((len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) * 2) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Times(3)

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Times(3)

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
		headers := http.Header{
			"Target_host": []string{baseURL},
		}
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, headers).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(customWebhookConfigURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(customWebhookConfigURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(customWebhookConfigURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(customWebhookConfigURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		return fetchReqSvc
	}

	successfulFetchRequestFetch := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdate := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdateForStaticDoc := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		fetchReqSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		return fetchReqSvc
	}

	successfulFetchRequestFetchAndUpdateForStaticDocForApplication := func() *automock.FetchRequestService {
		fetchReqSvc := &automock.FetchRequestService{}
		fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
			Times((len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) * 2) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

		fetchReqSvc.On("UpdateGlobal", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
			return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
		})).Return(nil).
			Times((len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL))) * 2) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones

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

	successfulEventUpdateForStaticDocWithApplication := func() *automock.EventService {
		eventSvc := &automock.EventService{}
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Once()
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, event1ID, *sanitizedStaticDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Times(2)
		eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, event2ID, *sanitizedStaticDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Times(2)
		eventSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixEvents(), nil).Times(5)

		eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
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

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
		return client
	}

	successfulClientFetchForStaticDoc := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDStaticDocument()}, baseURL, nil)
		return client
	}

	successfulClientFetchForStaticDocOnAppTemplateWithApplications := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDStaticDocument()}, baseURL, nil).Once()
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDStaticDocument()}, baseURL, nil).Once()
		return client
	}

	successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy := func() *automock.Client {
		ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{
			Headers:     http.Header{"Target_host": []string{baseURL}},
			Application: webhook.Application{BaseURL: baseURL},
		}

		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, fixWebhookForApplicationWithProxyURL(), ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocumentWithoutCredentialExchanges()}, baseURL, nil)
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

	successfulAppTemplateVersionListAndUpdateForApplication := func() *automock.ApplicationTemplateVersionService {
		svc := &automock.ApplicationTemplateVersionService{}
		svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Times(4)
		svc.On("Update", txtest.CtxWithDBMatcher(), appTemplateVersionID, appTemplateID, *fixAppTemplateVersionInput()).Return(nil).Twice()
		svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(fixAppTemplateVersion(), nil).Times(4)
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
		specSvcFn               func() *automock.SpecService
		fetchReqFn              func() *automock.FetchRequestService
		packageSvcFn            func() *automock.PackageService
		productSvcFn            func() *automock.ProductService
		vendorSvcFn             func() *automock.VendorService
		tombstoneSvcFn          func() *automock.TombstoneService
		tenantSvcFn             func() *automock.TenantService
		globalRegistrySvcFn     func() *automock.GlobalRegistryService
		appTemplateVersionSvcFn func() *automock.ApplicationTemplateVersionService
		appTemplateSvcFn        func() *automock.ApplicationTemplateService
		labelSvcFn              func() *automock.LabelService
		clientFn                func() *automock.Client
		ExpectedErr             error
	}{
		{
			Name: "Success for Application Template webhook with Static ORD data when resources are already in db and APIs/Events versions are incremented should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:       successfulAppTemplateNoAppsAppSvc,
			webhookSvcFn:   successfulWebhookListAppTemplate,
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
			eventSvcFn:              successfulEventUpdateForStaticDoc,
			specSvcFn:               successfulSpecRecreateAndUpdateForStaticDoc,
			fetchReqFn:              successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageSvcFn:            successfulPackageUpdateForStaticDoc,
			productSvcFn:            successfulProductUpdateForStaticDoc,
			vendorSvcFn:             successfulVendorUpdateForStaticDoc,
			tombstoneSvcFn:          successfulTombstoneUpdateForStaticDoc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionListAndUpdate,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDoc,
		},
		{
			Name: "Success for Application Template and Applications webhook with Static ORD data when resources are already in db and APIs/Events versions are incremented should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(69)
			},
			tenantSvcFn:    successfulTenantSvc,
			appSvcFn:       successfulAppSvc,
			webhookSvcFn:   successfulWebhookListAppTemplate,
			bundleSvcFn:    successfulBundleUpdateForStaticDocWithApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api1ID, *sanitizedStaticDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedStaticDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Times(2)
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api2ID, *sanitizedStaticDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Times(2)
				apiSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixAPIs(), nil).Times(5)
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, api2ID).Return(nil).Times(2)

				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIs(), nil).Once()
				return apiSvc
			},
			eventSvcFn:              successfulEventUpdateForStaticDocWithApplication,
			specSvcFn:               successfulSpecRecreateAndUpdateForStaticDocWithApplication,
			fetchReqFn:              successfulFetchRequestFetchAndUpdateForStaticDocForApplication,
			packageSvcFn:            successfulPackageUpdateForStaticDocWithApplication,
			productSvcFn:            successfulProductUpdateForStaticDocWithApplication,
			vendorSvcFn:             successfulVendorUpdateForStaticDocWithApplication,
			tombstoneSvcFn:          successfulTombstoneUpdateForStaticDocWithApplication,
			appTemplateVersionSvcFn: successfulAppTemplateVersionListAndUpdateForApplication,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDocOnAppTemplateWithApplications,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are already in db and APIs/Events versions are incremented should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
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
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageUpdateForApplication,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are already in db and APIs/Events versions are NOT incremented should Update them and refetch only failed API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api2ID, *sanitizedDoc.APIResources[1], nilSpecInput, map[string]string{bundleID: "http://localhost:8080/some-api/v1"}, map[string]string{}, []string{}, api2PreSanitizedHash, "").Return(nil).Once()
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, api2ID).Return(nil).Once()
				return apiSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				return eventSvc
			},
			specSvcFn:    successfulSpecRefetch,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageUpdateForApplication,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are not in db should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:                successfulAppGet,
			tenantSvcFn:             successfulTenantSvc,
			webhookSvcFn:            successfulTenantMappingOnlyCreation,
			webhookConvFn:           successfulWebhookConversion,
			bundleSvcFn:             successfulBundleCreate,
			apiSvcFn:                successfulAPICreateAndDelete,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when webhook has a proxy URL which should be used to access the document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(32)
			},
			appSvcFn:                successfulAppWithBaseURLSvc,
			tenantSvcFn:             successfulTenantSvc,
			webhookSvcFn:            successfulTenantMappingOnlyCreationWithProxyURL,
			bundleSvcFn:             successfulBundleCreateForApplicationForProxy,
			apiSvcFn:                successfulAPICreateAndDeleteForProxy,
			eventSvcFn:              successfulEventCreateForProxy,
			specSvcFn:               successfulSpecCreateAndUpdateForProxy,
			fetchReqFn:              successfulFetchRequestFetchAndUpdateForProxy,
			packageSvcFn:            successfulPackageCreateForProxy,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are not in db should Create them for a Static document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(35)
			},
			appSvcFn:     successfulAppTemplateNoAppsAppSvc,
			webhookSvcFn: successfulWebhookListAppTemplate,
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
			eventSvcFn:              successfulEventCreateForStaticDoc,
			specSvcFn:               successfulSpecCreateAndUpdateForStaticDoc,
			fetchReqFn:              successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageSvcFn:            successfulPackageCreateForStaticDoc,
			productSvcFn:            successfulProductCreateForStaticDoc,
			vendorSvcFn:             successfulVendorCreateForStaticDoc,
			tombstoneSvcFn:          successfulTombstoneCreateForStaticDoc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionForCreation,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDoc,
		},
		{
			Name: "Error when creating Application Template Version based on the doc",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(4, 3)
			},
			webhookSvcFn: successfulWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return("", testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
		},
		{
			Name: "Error when getting Application Template from the webhook ObjectID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			webhookSvcFn: successfulWebhookListAppTemplate,
			appTemplateSvcFn: func() *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr)
				return svc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
		},
		{
			Name: "Error when listing Application Template Version by app template ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
			},
			webhookSvcFn: successfulWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
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
			webhookSvcFn: successfulWebhookListAppTemplate,
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
			webhookSvcFn: successfulWebhookListAppTemplate,
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
			webhookSvcFn: successfulWebhookListAppTemplate,
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
			webhookSvcFn: successfulWebhookListAppTemplate,
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
		},
		{
			Name: "Success when there is ORD webhook on app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(38)
			},
			appSvcFn:       successfulAppSvc,
			tenantSvcFn:    successfulTenantSvc,
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
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageUpdateForApplication,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil).Once()
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResources, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil).Once()
				return client
			},
			labelSvcFn: successfulLabelGetByKey,
		},
		{
			Name: "Error when synchronizing global resources from global registry should get them from DB and proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:                successfulAppGet,
			tenantSvcFn:             successfulTenantSvc,
			webhookSvcFn:            successfulTenantMappingOnlyCreation,
			webhookConvFn:           successfulWebhookConversion,
			bundleSvcFn:             successfulBundleCreate,
			apiSvcFn:                successfulAPICreateAndDelete,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", context.TODO()).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
				return globalRegistrySvcFn
			},
			clientFn:   successfulClientFetch,
			labelSvcFn: successfulLabelGetByKey,
		},
		{
			Name: "Error when synchronizing global resources from global registry and get them from DB should proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn:                successfulAppGet,
			tenantSvcFn:             successfulTenantSvc,
			webhookSvcFn:            successfulTenantMappingOnlyCreation,
			webhookConvFn:           successfulWebhookConversion,
			bundleSvcFn:             successfulBundleCreate,
			apiSvcFn:                successfulAPICreateAndDelete,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", context.TODO()).Return(nil, errors.New("error")).Once()
				return globalRegistrySvcFn
			},
			clientFn:   successfulClientFetch,
			labelSvcFn: successfulLabelGetByKey,
		},
		{
			Name:            "Returns error when list by webhook type fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			ExpectedErr:         testErr,
		},
		{
			Name:                "Returns error when transaction opening fails",
			TransactionerFn:     txGen.ThatFailsOnBegin,
			ExpectedErr:         testErr,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
		},
		{
			Name:                "Returns error when first transaction commit fails",
			TransactionerFn:     txGen.ThatFailsOnCommit,
			webhookSvcFn:        successfulWebhookList,
			ExpectedErr:         testErr,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			appSvcFn:            successfulAppGetOnce,
			tenantSvcFn:         successfulTenantSvcOnce,
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			labelSvcFn:          successfulLabelGetByKey,
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			tenantSvcFn: successfulTenantSvcOnce,
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return appSvc
			},
			webhookSvcFn:        successfulWebhookList,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
		},
		{
			Name: "Does not resync resources when event list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
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
		},
		{
			Name: "Does not resync resources when api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
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
		},
		{
			Name: "Returns error when list all applications by app template id fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(5)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
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
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when get internal tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
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
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when get tenant id fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
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
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
		},
		{
			Name: "Returns error when application locking fails for ORD webhook for app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)
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
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
		},
		{
			Name: "Skips webhook when ORD documents fetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(3)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(3)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(nil, "", testErr)
				return client
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			labelSvcFn:          successfulLabelGetByKey,
		},
		{
			Name: "Update application local tenant id when ord local id is unique and application does not have local tenant id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(33)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalID")}).Return(nil).Once()
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
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageUpdateForApplication,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulVendorUpdateForApplication,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to update application local tenant id when ord local id is unique and application does not have local tenant id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalID")}).Return(testErr).Once()
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
			packageSvcFn: func() *automock.PackageService {
				packagesSvc := &automock.PackageService{}
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackagesWithHash(), nil).Once()
				return packagesSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources for invalid ORD documents when event resource name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)

				return persistTx, transact
			},
			appSvcFn:            successfulAppGet,
			tenantSvcFn:         successfulTenantSvc,
			webhookSvcFn:        successfulTenantMappingOnlyCreation,
			webhookConvFn:       successfulWebhookConversion,
			bundleSvcFn:         successfulBundleCreate,
			apiSvcFn:            successfulAPICreateAndDelete,
			eventSvcFn:          successfulOneEventCreate,
			specSvcFn:           successfulSpecWithOneEventCreateAndUpdate,
			fetchReqFn:          successfulFetchRequestFetchAndUpdate,
			packageSvcFn:        successfulPackageCreate,
			productSvcFn:        successfulProductCreate,
			vendorSvcFn:         successfulVendorCreate,
			tombstoneSvcFn:      successfulTombstoneCreate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources for invalid ORD documents when bundle name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(32)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)

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
			specSvcFn:           successfulSpecCreateAndUpdate,
			fetchReqFn:          successfulFetchRequestFetchAndUpdate,
			packageSvcFn:        successfulPackageCreate,
			productSvcFn:        successfulProductCreate,
			vendorSvcFn:         successfulVendorCreate,
			tombstoneSvcFn:      successfulTombstoneCreate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources for invalid ORD documents when vendor ordID is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)

				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion,
			bundleSvcFn:   successfulBundleCreate,
			apiSvcFn:      successfulAPICreateAndDelete,
			eventSvcFn:    successfulEventCreate,
			specSvcFn:     successfulSpecCreateAndUpdate,
			fetchReqFn:    successfulFetchRequestFetchAndUpdate,
			packageSvcFn:  successfulPackageCreate,
			productSvcFn:  successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				return vendorSvc
			},
			tombstoneSvcFn:      successfulTombstoneCreate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources for invalid ORD documents when product title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)

				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion,
			bundleSvcFn:   successfulBundleCreate,
			apiSvcFn:      successfulAPICreateAndDelete,
			eventSvcFn:    successfulEventCreate,
			specSvcFn:     successfulSpecCreateAndUpdate,
			fetchReqFn:    successfulFetchRequestFetchAndUpdate,
			packageSvcFn:  successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Twice()
				return productSvc
			},
			vendorSvcFn:         successfulVendorCreate,
			tombstoneSvcFn:      successfulTombstoneCreate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Products[0].Title = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources for invalid ORD documents when package title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(6)

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
			fetchReqFn: successfulFetchRequestFetchAndUpdate,
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
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Packages[0].Title = ""
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if vendor list fails",
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
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list vendors after resync",
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
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if vendor update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

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
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorID, *sanitizedDoc.Vendors[0]).Return(testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if vendor create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

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
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[0]).Return("", testErr).Once()
				return vendorSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if product list fails",
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
			vendorSvcFn:  successfulVendorUpdateForApplication,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return productSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list products after resync",
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
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if product update fails",
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
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if product create fails",
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
			packageSvcFn:            successfulEmptyPackageList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if package list fails",
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list packages after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(16)

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
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				packagesSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageID, *sanitizedDoc.Packages[0], packagePreSanitizedHash).Return(nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixPackages(), nil).Once()
				packagesSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return packagesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if package update fails",
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if package create fails",
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if bundle list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(16)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(16)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list bundles after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(19)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(20)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if bundle update fails",
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
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			apiSvcFn:                successfulEmptyAPIList,
			eventSvcFn:              successfulEmptyEventList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if bundle create fails",
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if bundle have different tenant mapping configuration",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

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
			bundleSvcFn:  successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithMultipleSameTypesFormat, credentialExchangeStrategyType, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if webhooks could not be enriched",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(18)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(17)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if webhooks cannot be listed for application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

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
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if webhooks cannot be converted from graphql input to model input",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

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
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if webhooks cannot be created",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(18)

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
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources if webhooks can be created successfully",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(34)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			tombstoneSvcFn:          successfulTombstoneCreate,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not recreate tenant mapping webhooks if there are no differences",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(34)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(32)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			tombstoneSvcFn:          successfulTombstoneCreate,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreate,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does recreate of tenant mapping webhooks when there are differences",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(34)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
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
			tombstoneSvcFn:          successfulTombstoneCreate,
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreateAndUpdate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not recreate of tenant mapping webhooks when there are differences but deletion fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(19)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(19)
				return persistTx, transact
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whInputs := []*graphql.WebhookInput{fixTenantMappingWebhookGraphQLInput()}
				whSvc.On("ListByWebhookType", txtest.CtxWithDBMatcher(), model.WebhookTypeOpenResourceDiscovery).Return(fixWebhooksForApplication(), nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, baseURL, nil)
				return client
			},
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return apiSvc
			},
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				return tombstoneSvc
			},
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				return eventSvc
			},
			specSvcFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(21)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(20)
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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list apis after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(24)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(24)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(23)
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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if fetching bundle ids for api fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(22)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(21)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

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
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(21)

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
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
				apiSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, api1ID, *sanitizedDoc.APIResources[0], nilSpecInput, map[string]string{bundleID: sanitizedDoc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL}, map[string]string{}, []string{}, api1PreSanitizedHash, "").Return(nil).Once()
				return apiSvc
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("ListIDByReferenceObjectID", txtest.CtxWithDBMatcher(), resource.Application, model.APISpecReference, api1ID).Return(nil, testErr).Once()
				return specSvc
			},
			eventSvcFn:              successfulEmptyEventList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api spec get fetch request fails",
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
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			productSvcFn:   successfulProductUpdateForApplication,
			vendorSvcFn:    successfulVendorUpdateForApplication,
			packageSvcFn:   successfulPackageUpdateForApplication,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn: func() *automock.APIService {
				apiSvc := &automock.APIService{}
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Twice()
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources returns error if api spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
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
				apiSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixAPIsNoVersionBump(), nil).Times(3)
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

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

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
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event list fails",
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
			bundleRefSvcFn: successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiSvcFn:       successfulAPIUpdate,
			specSvcFn:      successfulAPISpecUpdate,
			eventSvcFn: func() *automock.EventService {
				eventSvc := &automock.EventService{}
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEvents(), nil).Once()
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return eventSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list events after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(27)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(28)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(27)
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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if fetching bundle ids for event fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync specification resources if event create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(26)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(25)
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
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(nil, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
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
		},
		{
			Name: "Does not resync resources if event spec delete fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event spec create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

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
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event spec list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(24)

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
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event spec get fetch request fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(25)

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
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Twice()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Resync resources returns error if event spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(33)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(33)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(33)
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
				eventSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixEventsNoVersionBump(), nil).Times(3)
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event1ID, *sanitizedDoc.EventResources[0], nilSpecInput, []string{bundleID}, []string{}, []string{}, event1PreSanitizedHash, "").Return(nil).Once()
				eventSvc.On("UpdateInManyBundles", txtest.CtxWithDBMatcher(), resource.Application, event2ID, *sanitizedDoc.EventResources[1], nilSpecInput, []string{bundleID}, []string{}, []string{}, event2PreSanitizedHash, "").Return(nil).Once()
				return eventSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionFailed}).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))
				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionFailed
				})).Return(nil).
					Times(len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				return fetchReqSvc
			},
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if tombstone list fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

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
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecRecreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Fails to list tombstones after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(30)
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
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecRecreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if tombstone update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(29)
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
			eventSvcFn:     successfulEventUpdate,
			specSvcFn:      successfulSpecRecreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if tombstone create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(29)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(30)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(29)
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
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Tombstones[0]).Return("", testErr).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if api resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
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
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreate,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if package resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:   successfulAPICreate,
			eventSvcFn: successfulEventCreate,
			specSvcFn:  successfulSpecCreate,
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
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = packageORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = packageORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = packageORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if event resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
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
			specSvcFn:    successfulSpecCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn:  successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = event1ORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = event1ORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = event1ORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if vendor resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: successfulProductCreate,
			vendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[0]).Return("", nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Vendors[1]).Return("", nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixVendors(), nil).Once()
				vendorSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, vendorID).Return(testErr).Once()
				return vendorSvc
			},
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = vendorORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = vendorORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = vendorORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if product resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiSvcFn:     successfulAPICreate,
			eventSvcFn:   successfulEventCreate,
			specSvcFn:    successfulSpecCreate,
			packageSvcFn: successfulPackageCreate,
			productSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *sanitizedDoc.Products[0]).Return("", nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixProducts(), nil).Once()
				productSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, productID).Return(testErr).Once()
				return productSvc
			},
			vendorSvcFn: successfulVendorCreate,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				ts := fixSanitizedORDDocument().Tombstones[0]
				ts.OrdID = productORDID
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = productORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = productORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Does not resync resources if bundle resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(31)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(32)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(31)
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
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *ts).Return("", nil).Once()
				tombstones := fixTombstones()
				tombstones[0].OrdID = bundleORDID
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(tombstones, nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = bundleORDID
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Returns error when failing to open final transaction to commit fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(32)
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
			eventSvcFn:              successfulEventCreate,
			specSvcFn:               successfulSpecCreate,
			fetchReqFn:              successfulFetchRequestFetch,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Returns error when failing to find spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(32)
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
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
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

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(nil, testErr).Once()

				return specSvc
			},
			fetchReqFn:              successfulFetchRequestFetch,
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Returns error when failing to update spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(32)
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
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
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
			tombstoneSvcFn:          successfulTombstoneCreate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Returns error when failing to update fetch request in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(32)
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
			eventSvcFn: successfulEventCreate,
			specSvcFn: func() *automock.SpecService {
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

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Once()

				return specSvc
			},
			fetchReqFn: func() *automock.FetchRequestService {
				fetchReqSvc := &automock.FetchRequestService{}
				fetchReqSvc.On("FetchSpec", txtest.CtxWithDBMatcher(), mock.Anything, http.Header{}).Return(&testSpecData, &model.FetchRequestStatus{Condition: model.FetchRequestStatusConditionSucceeded}).
					Times(len(fixAPI1SpecInputs(baseURL)) + len(fixAPI2SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)))

				fetchReqSvc.On("Update", txtest.CtxWithDBMatcher(), mock.MatchedBy(func(actual *model.FetchRequest) bool {
					return actual.Status.Condition == model.FetchRequestStatusConditionSucceeded
				})).Return(testErr).Once()

				return fetchReqSvc
			},
			packageSvcFn:            successfulPackageCreate,
			productSvcFn:            successfulProductCreate,
			vendorSvcFn:             successfulVendorCreate,
			tombstoneSvcFn:          successfulTombstoneCreate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(31)
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
			eventSvcFn:          successfulEventCreate,
			specSvcFn:           successfulSpecCreateAndUpdate,
			fetchReqFn:          successfulFetchRequestFetchAndUpdate,
			packageSvcFn:        successfulPackageCreate,
			productSvcFn:        successfulProductCreate,
			vendorSvcFn:         successfulEmptyVendorList,
			tombstoneSvcFn:      successfulTombstoneCreate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(31)
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
			eventSvcFn:   successfulEventUpdate,
			specSvcFn:    successfulSpecRecreateAndUpdate,
			fetchReqFn:   successfulFetchRequestFetchAndUpdate,
			packageSvcFn: successfulPackageUpdateForApplication,
			productSvcFn: successfulProductUpdateForApplication,
			vendorSvcFn:  successfulEmptyVendorList,
			tombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstoneID, *sanitizedDoc.Tombstones[0]).Return(nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(fixTombstones(), nil).Once()
				return tombstoneSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{doc}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
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

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{})
			err := svc.SyncORDDocuments(context.TODO(), ord.MetricsConfig{})
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func TestService_ProcessApplications(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	ordMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: http.Header{}}

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
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, ordMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil)
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
		appIDs              func() []string
		ExpectedErr         error
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
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
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
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			appSvcFn:    successfulAppGet,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetch,
			appIDs: func() []string {
				return []string{appID}
			},
			labelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", txtest.CtxWithDBMatcher(), tenantID, model.ApplicationLabelableObject, testApplication.Name, application.ApplicationTypeLabelKey).Return(fixApplicationTypeLabel(), nil).Once()
				return svc
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			tenantSvcFn: successfulTenantSvcOnce,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplication", txtest.CtxWithDBMatcher(), appID).Return(fixWebhooksForApplication(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageSvc := &automock.PackageService{}
			productSvc := &automock.ProductService{}
			vendorSvc := &automock.VendorService{}
			tombstoneSvc := &automock.TombstoneService{}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			appTemplateSvc := &automock.ApplicationTemplateService{}

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{})
			err := svc.ProcessApplications(context.TODO(), ord.MetricsConfig{}, test.appIDs())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func TestService_ProcessApplicationTemplates(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	ordMapping := application.ORDWebhookMapping{}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: http.Header{}}

	testApplication := fixApplications()[0]
	testResourceApp := ord.Resource{
		Type:          resource.Application,
		ID:            testApplication.ID,
		Name:          testApplication.Name,
		LocalTenantID: testApplication.LocalTenantID,
		ParentID:      &appTemplateID,
	}
	testResourceAppTemplate := ord.Resource{
		Type: resource.ApplicationTemplate,
		ID:   appTemplateID,
		Name: appTemplateName,
	}
	testWebhookForAppTemplate := fixOrdWebhooksForAppTemplate()[0]

	successfulClientFetchForAppTemplate := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil).Once()
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceApp, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil).Once()
		return client
	}

	successfulClientFetchForOnlyAppTemplate := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceAppTemplate, testWebhookForAppTemplate, ordMapping, ordRequestObject).Return(ord.Documents{}, baseURL, nil).Once()
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
		appTemplateIDs          func() []string
		ExpectedErr             error
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
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
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
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			appSvcFn:    successfulAppSvc,
			tenantSvcFn: successfulTenantSvc,
			webhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixOrdWebhooksForAppTemplate(), nil).Once()
				return whSvc
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForAppTemplate,
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
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
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
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
				return txGen.ThatSucceedsMultipleTimes(3)
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
			clientFn:            successfulClientFetchForOnlyAppTemplate,
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
		},
		{
			Name: "Error while getting application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
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
			clientFn:            successfulClientFetchForOnlyAppTemplate,
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
		},
		{
			Name: "Error while getting lowest owner of resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
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
			clientFn:            successfulClientFetchForOnlyAppTemplate,
			appTemplateIDs: func() []string {
				return []string{appTemplateID}
			},
			appTemplateSvcFn: successAppTemplateGetSvc,
			ExpectedErr:      testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

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

			appSvc := &automock.ApplicationService{}
			if test.appSvcFn != nil {
				appSvc = test.appSvcFn()
			}
			whSvc := &automock.WebhookService{}
			if test.webhookSvcFn != nil {
				whSvc = test.webhookSvcFn()
			}
			whConv := &automock.WebhookConverter{}
			if test.webhookConvFn != nil {
				whConv = test.webhookConvFn()
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

			ordCfg := ord.NewServiceConfig(4, 100, 0, "", false, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiSvc, eventSvc, specSvc, fetchReqSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, whConv, appTemplateVersionSvc, appTemplateSvc, labelSvc, []application.ORDWebhookMapping{})
			err := svc.ProcessApplicationTemplates(context.TODO(), ord.MetricsConfig{}, test.appTemplateIDs())
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, apiSvc, eventSvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func successfulGlobalRegistrySvc() *automock.GlobalRegistryService {
	globalRegistrySvcFn := &automock.GlobalRegistryService{}
	globalRegistrySvcFn.On("SyncGlobalResources", context.TODO()).Return(map[string]bool{ord.SapVendor: true}, nil).Once()
	return globalRegistrySvcFn
}

func successfulTenantSvc() *automock.TenantService {
	tenantSvc := &automock.TenantService{}
	tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
	tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
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

func successfulAppTemplateNoAppsAppSvc() *automock.ApplicationService {
	appSvc := &automock.ApplicationService{}
	appSvc.On("ListAllByApplicationTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(nil, nil).Once()

	return appSvc
}

func successAppTemplateGetSvc() *automock.ApplicationTemplateService {
	svc := &automock.ApplicationTemplateService{}
	svc.On("Get", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplate(), nil)
	return svc
}
