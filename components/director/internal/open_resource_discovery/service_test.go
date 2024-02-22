package ord_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"

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
	sapVendor                            = "sap:vendor:SAP:"
)

func TestService_Processing(t *testing.T) {
	testErr := errors.New("Test error")
	validatingORDDocsErr := []*ord.ValidationError{{OrdID: "ns:eventResource:test:v1"}}
	txGen := txtest.NewTransactionContextGenerator(testErr)

	emptyORDMapping := application.ORDWebhookMapping{}
	ordMappingWithProxy := application.ORDWebhookMapping{ProxyURL: proxyURL, Type: applicationTypeLabelValue}
	ordRequestObject := webhook.OpenResourceDiscoveryWebhookRequestObject{Headers: &sync.Map{}}

	sanitizedDoc := fixSanitizedORDDocument()
	sanitizedDocForProxy := fixSanitizedORDDocumentForProxyURL()
	sanitizedStaticDoc := fixSanitizedStaticORDDocument()
	var testSpecData = "{}"
	var testSpec = model.Spec{}
	var bundles []*model.Bundle

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

	successfulIntegrationDependencyProcessing := func() *automock.IntegrationDependencyProcessor {
		integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
		integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(fixORDDocument())).Return(fixIntegrationDependencies(), nil).Once()
		return integrationDependencyProcessor
	}

	successfulDataProductProcessing := func() *automock.DataProductProcessor {
		dataProductProcessor := &automock.DataProductProcessor{}
		dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(fixORDDocument())).Return(fixDataProducts(), nil).Once()
		return dataProductProcessor
	}

	successfulTombstoneProcessing := func() *automock.TombstoneProcessor {
		tombstoneProcessor := &automock.TombstoneProcessor{}
		tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixORDDocument().Tombstones).Return(fixTombstones(), nil).Once()
		return tombstoneProcessor
	}

	successfulBundleCreateForApplicationForProxy := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		currentTime := time.Now().Format(time.RFC3339)
		sanitizedDocForProxy.ConsumptionBundles[0].LastUpdate = &currentTime
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleUpdateForApplication := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleUpdateForStaticDoc := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, bundleID, bundleUpdateInputFromCreateInput(*sanitizedStaticDoc.ConsumptionBundles[0]), bundlePreSanitizedHashStaticDoc).Return(nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundlesWithCredentialExchangeStrategies(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreate := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreateForStaticDoc := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, mock.Anything, mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationTemplateVersionIDNoPaging", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulBundleCreateWithGenericParam := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
		bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
		bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
		return bundlesSvc
	}

	successfulListTwiceAndCreateBundle := func() *automock.BundleService {
		bundlesSvc := &automock.BundleService{}
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

	successfulVendorProcess := func() *automock.VendorProcessor {
		vendorProcessor := &automock.VendorProcessor{}
		vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Vendors).Return(fixVendors(), nil).Once()
		return vendorProcessor
	}

	successfulVendorProcessForStaticDoc := func() *automock.VendorProcessor {
		vendorProcessor := &automock.VendorProcessor{}
		vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.Vendors).Return(fixVendors(), nil).Once()
		return vendorProcessor
	}

	successfulProductProcess := func() *automock.ProductProcessor {
		productProcessor := &automock.ProductProcessor{}
		productProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Products).Return(fixProducts(), nil).Once()
		return productProcessor
	}

	successfulProductProcessForStaticDoc := func() *automock.ProductProcessor {
		productProcessor := &automock.ProductProcessor{}
		productProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.Products).Return(fixProducts(), nil).Once()
		return productProcessor
	}

	successfulPackageProcess := func() *automock.PackageProcessor {
		packageProcessor := &automock.PackageProcessor{}
		packageProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Packages, mock.Anything).Return(fixPackages(), nil).Once()
		return packageProcessor
	}

	successfulPackageProcessForStaticDoc := func() *automock.PackageProcessor {
		packageProcessor := &automock.PackageProcessor{}
		packageProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.Packages, mock.Anything).Return(fixPackages(), nil).Once()
		return packageProcessor
	}

	successfulAPIProcess := func() *automock.APIProcessor {
		apiProcessor := &automock.APIProcessor{}
		apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.APIResources, mock.Anything).Return(fixAPIs(), fixAPIsFetchRequests(), nil).Once()
		return apiProcessor
	}

	successfulEventProcess := func() *automock.EventProcessor {
		eventProcessor := &automock.EventProcessor{}
		eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.EventResources, mock.Anything).Return(fixEvents(), fixEventsFetchRequests(), nil).Once()
		return eventProcessor
	}

	successfulCapabilityProcess := func() *automock.CapabilityProcessor {
		capabilityProcessor := &automock.CapabilityProcessor{}
		capabilityProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.Capabilities, mock.Anything).Return(fixCapabilities(), fixCapabilityFetchRequests(), nil).Once()
		return capabilityProcessor
	}

	successfulCapabilityProcessForProxy := func() *automock.CapabilityProcessor {
		capabilityProcessor := &automock.CapabilityProcessor{}
		capabilityProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDocForProxy.Capabilities, mock.Anything).Return(fixCapabilities(), fixCapabilityFetchRequests(), nil).Once()
		return capabilityProcessor
	}

	successfulCapabilityProcessForStaticDoc := func() *automock.CapabilityProcessor {
		capabilityProcessor := &automock.CapabilityProcessor{}
		capabilityProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixPackages(), sanitizedStaticDoc.Capabilities, mock.Anything).Return(fixCapabilities(), fixCapabilityFetchRequests(), nil).Once()
		return capabilityProcessor
	}

	successfulEntityTypeProcess := func() *automock.EntityTypeProcessor {
		entityTypeProcessor := &automock.EntityTypeProcessor{}
		entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
		return entityTypeProcessor
	}

	successfulSpecCreateAndUpdateForProxy := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
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
		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecCreateAndUpdateForStaticDoc := func() *automock.SpecService {
		specSvc := &automock.SpecService{}
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
		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
			Times(len(fixAPI1SpecInputs(baseURL)) + len(fixEvent1SpecInputs()) + len(fixEvent2SpecInputs(baseURL)) + 2*len(fixCapabilitySpecInputs())) // len(fixAPI2SpecInputs(baseURL)) is excluded because it's API is part of tombstones, 2 * len(fixCapabilitySpecInputs(), because it is used twice

		return specSvc
	}

	successfulSpecRecreateAndUpdateForStaticDoc := func() *automock.SpecService {
		specSvc := &automock.SpecService{}

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
		specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Times(5)

		expectedSpecToUpdate := testSpec
		expectedSpecToUpdate.Data = &testSpecData
		specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).Times(5)

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

	successfulClientFetch := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, []string{}, baseURL, nil)
		return client
	}

	successfulClientFetchForStaticDoc := func() *automock.Client {
		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceForAppTemplate, testStaticWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDStaticDocument()}, []string{}, baseURL, nil)
		return client
	}

	successfulTombstoneResDeleter := func() *automock.TombstonedResourcesDeleter {
		tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
		fetchReq := fixAPIsFetchRequests()[0:3]
		fetchReq = append(fetchReq, fixEventsFetchRequests()...)
		fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
		tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, fixVendors(), fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
		return tombstonedResourcesDeleterFn
	}

	successfulTombstoneResDeleterForStaticDoc := func() *automock.TombstonedResourcesDeleter {
		tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
		fetchReq := fixAPIsFetchRequests()[0:3]
		fetchReq = append(fetchReq, fixEventsFetchRequests()...)
		fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
		tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, fixVendors(), fixProducts(), fixPackages(), fixBundlesWithCredentialExchangeStrategies(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
		return tombstonedResourcesDeleterFn
	}

	successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy := func() *automock.Client {
		headerMatcher := func() interface{} {
			return mock.MatchedBy(func(ordRequestObject webhook.OpenResourceDiscoveryWebhookRequestObject) bool {
				value, ok := ordRequestObject.Headers.Load("target_host")
				return ok && value == baseURL && ordRequestObject.Application.BaseURL == baseURL
			})
		}

		client := &automock.Client{}
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, fixWebhookForApplicationWithProxyURL(), ordMappingWithProxy, headerMatcher()).Return(ord.Documents{fixORDDocumentWithoutCredentialExchanges()}, []string{}, baseURL, nil)
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

	successfulDocumentValidatorForStaticDocFn := func() *automock.Validator {
		docValidator := &automock.Validator{}
		docValidator.On("Validate", mock.Anything, []*ord.Document{fixORDStaticDocument()}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
		return docValidator
	}

	successfulDocumentValidatorForApplicationFn := func() *automock.Validator {
		docValidator := &automock.Validator{}
		docValidator.On("Validate", mock.Anything, []*ord.Document{fixORDDocument()}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
		return docValidator
	}

	successfulDocumentValidatorForProxyFn := func() *automock.Validator {
		docValidator := &automock.Validator{}
		docValidator.On("Validate", mock.Anything, []*ord.Document{fixORDDocumentWithoutCredentialExchanges()}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
		return docValidator
	}

	testCases := []struct {
		Name                             string
		TransactionerFn                  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		appSvcFn                         func() *automock.ApplicationService
		webhookSvcFn                     func() *automock.WebhookService
		webhookConvFn                    func() *automock.WebhookConverter
		bundleSvcFn                      func() *automock.BundleService
		bundleRefSvcFn                   func() *automock.BundleReferenceService
		apiProcessorFn                   func() *automock.APIProcessor
		eventProcessorFn                 func() *automock.EventProcessor
		capabilityProcessorFn            func() *automock.CapabilityProcessor
		entityTypeProcessorFn            func() *automock.EntityTypeProcessor
		integrationDependencyProcessorFn func() *automock.IntegrationDependencyProcessor
		dataProductProcessorFn           func() *automock.DataProductProcessor
		specSvcFn                        func() *automock.SpecService
		fetchReqFn                       func() *automock.FetchRequestService
		packageProcessorFn               func() *automock.PackageProcessor
		productProcessorFn               func() *automock.ProductProcessor
		vendorProcessorFn                func() *automock.VendorProcessor
		tombstoneProcessorFn             func() *automock.TombstoneProcessor
		tenantSvcFn                      func() *automock.TenantService
		globalRegistrySvcFn              func() *automock.GlobalRegistryService
		appTemplateVersionSvcFn          func() *automock.ApplicationTemplateVersionService
		appTemplateSvcFn                 func() *automock.ApplicationTemplateService
		tombstonedResourcesDeleterFn     func() *automock.TombstonedResourcesDeleter
		labelSvcFn                       func() *automock.LabelService
		clientFn                         func() *automock.Client
		processFnName                    string
		webhookMappings                  []application.ORDWebhookMapping
		documentValidatorFn              func() *automock.Validator
		ExpectedErr                      error
	}{
		{
			Name: "Success for Application Template webhook with Static ORD data when resources are already in db and APIs/Events last update fields are newer should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			webhookSvcFn:   successfulStaticWebhookListAppTemplate,
			bundleSvcFn:    successfulBundleUpdateForStaticDoc,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixBundlesWithCredentialExchangeStrategies(), fixPackages(), sanitizedStaticDoc.APIResources, mock.Anything).Return(fixAPIs(), fixAPIsFetchRequests(), nil).Once()
				return apiProcessor
			},
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixBundlesWithCredentialExchangeStrategies(), fixPackages(), sanitizedStaticDoc.EventResources, mock.Anything).Return(fixEvents(), fixEventsFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				entityTypeProcessor := &automock.EntityTypeProcessor{}
				entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixEntityTypes(), nil).Once()
				return entityTypeProcessor
			},
			capabilityProcessorFn: successfulCapabilityProcessForStaticDoc,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixPackages(), sanitizedStaticDoc.IntegrationDependencies, fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			specSvcFn:          successfulSpecRecreateAndUpdateForStaticDoc,
			fetchReqFn:         successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageProcessorFn: successfulPackageProcessForStaticDoc,
			productProcessorFn: successfulProductProcessForStaticDoc,
			vendorProcessorFn:  successfulVendorProcessForStaticDoc,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixORDStaticDocument().Tombstones).Return(fixTombstones(), nil).Once()
				return tombstoneProcessor
			},
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleterForStaticDoc,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionListAndUpdate,
			appTemplateSvcFn:             successAppTemplateGetSvc,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetchForStaticDoc,
			processFnName:                processApplicationTemplateFnName,
			documentValidatorFn:          successfulDocumentValidatorForStaticDocFn,
		},
		{
			Name: "Success when resources are already in db and APIs/Events last update fields are newer should Update them and resync API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			bundleSvcFn:                      successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecRecreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn:                         successfulClientFetch,
			processFnName:                    processApplicationFnName,
			labelSvcFn:                       successfulLabelGetByKey,
			documentValidatorFn:              successfulDocumentValidatorForApplicationFn,
		},
		{
			Name: "Success when resources are already in db and APIs/Events last update fields are NOT newer should Update them and refetch only failed API/Event specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleUpdateForApplication,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.APIResources, mock.Anything).Return(fixAPIsNoNewerLastUpdate(), fixFailedAPIFetchRequests(), nil).Once()
				return apiProcessor
			},
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.EventResources, mock.Anything).Return(fixEventsNoNewerLastUpdate(), fixFailedEventsFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				entityTypeProcessor := &automock.EntityTypeProcessor{}
				entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDDocument())).Return(fixEntityTypes(), nil).Once()
				return entityTypeProcessor
			},
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecRefetch,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixFailedAPIFetchRequests()[:1]
				fetchReq = append(fetchReq, fixFailedEventsFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, fixVendors(), fixProducts(), fixPackages(), fixBundles(), fixAPIsNoNewerLastUpdate(), fixEventsNoNewerLastUpdate(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			processFnName:           processApplicationFnName,
			labelSvcFn:              successfulLabelGetByKey,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
		},
		{
			Name: "Success when resources are not in db should Create them",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			bundleSvcFn:                      successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecCreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn:                         successfulClientFetch,
			processFnName:                    processApplicationFnName,
			labelSvcFn:                       successfulLabelGetByKey,
			documentValidatorFn:              successfulDocumentValidatorForApplicationFn,
		},
		{
			Name: "Success when webhook has a proxy URL which should be used to access the document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:     successfulAppWithBaseURLSvc,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulTenantMappingOnlyCreationWithProxyURL,
			bundleSvcFn:  successfulBundleCreateForApplicationForProxy,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDocForProxy.APIResources, mock.Anything).Return(fixAPIs(), fixAPIsFetchRequests(), nil).Once()
				return apiProcessor
			},
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDocForProxy.EventResources, mock.Anything).Return(fixEvents(), fixEventsFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				entityTypeProcessor := &automock.EntityTypeProcessor{}
				entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDocForProxy.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return entityTypeProcessor
			},
			capabilityProcessorFn: successfulCapabilityProcessForProxy,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDocForProxy.IntegrationDependencies, mock.Anything).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductsProcessor := &automock.DataProductProcessor{}
				dataProductsProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDocForProxy.DataProducts, mock.Anything).Return(fixDataProducts(), nil).Once()
				return dataProductsProcessor
			},
			specSvcFn:  successfulSpecCreateAndUpdateForProxy,
			fetchReqFn: successfulFetchRequestFetchAndUpdateForProxy,
			packageProcessorFn: func() *automock.PackageProcessor {
				packageProcessor := &automock.PackageProcessor{}
				packageProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDocForProxy.Packages, mock.Anything).Return(fixPackages(), nil).Once()
				return packageProcessor
			},
			productProcessorFn:           successfulProductProcess,
			vendorProcessorFn:            successfulVendorProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionList,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetchForDocWithoutCredentialExchangeStrategiesWithProxy,
			processFnName:                processApplicationFnName,
			webhookMappings:              []application.ORDWebhookMapping{ordMappingWithProxy},
			labelSvcFn:                   successfulLabelGetByKey,
			documentValidatorFn:          successfulDocumentValidatorForProxyFn,
		},
		{
			Name: "Success when resources are not in db should Create them for a Static document",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			bundleSvcFn:  successfulBundleCreateForStaticDoc,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixBundles(), fixPackages(), sanitizedStaticDoc.APIResources, mock.Anything).Return(fixAPIs(), fixAPIsFetchRequests(), nil).Once()
				return apiProcessor
			},
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixBundles(), fixPackages(), sanitizedStaticDoc.EventResources, mock.Anything).Return(fixEvents(), fixEventsFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				entityTypeProcessor := &automock.EntityTypeProcessor{}
				entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, sanitizedStaticDoc.EntityTypes, fixPackages(), mock.Anything).Return(fixEntityTypes(), nil).Once()
				return entityTypeProcessor
			},
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixPackages(), sanitizedStaticDoc.IntegrationDependencies, fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductsProcessor := &automock.DataProductProcessor{}
				dataProductsProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixPackages(), sanitizedStaticDoc.DataProducts, fixResourceHashesForDocument(fixORDStaticDocument())).Return(fixDataProducts(), nil).Once()
				return dataProductsProcessor
			},
			capabilityProcessorFn: successfulCapabilityProcessForStaticDoc,
			specSvcFn:             successfulSpecCreateAndUpdateForStaticDoc,
			fetchReqFn:            successfulFetchRequestFetchAndUpdateForStaticDoc,
			packageProcessorFn:    successfulPackageProcessForStaticDoc,
			productProcessorFn:    successfulProductProcessForStaticDoc,
			vendorProcessorFn:     successfulVendorProcessForStaticDoc,
			tombstoneProcessorFn: func() *automock.TombstoneProcessor {
				tombstoneProcessor := &automock.TombstoneProcessor{}
				tombstoneProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, fixORDStaticDocument().Tombstones).Return(fixTombstones(), nil).Once()
				return tombstoneProcessor
			},
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixAPIsFetchRequests()[0:3]
				fetchReq = append(fetchReq, fixEventsFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, fixVendors(), fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionForCreation,
			appTemplateSvcFn:        successAppTemplateGetSvc,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetchForStaticDoc,
			processFnName:           processApplicationTemplateFnName,
			documentValidatorFn:     successfulDocumentValidatorForStaticDocFn,
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
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(5)

				transact.On("Begin").Return(persistTx, nil).Once()
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
			documentValidatorFn: successfulDocumentValidatorForStaticDocFn,
			ExpectedErr:         testErr,
		},
		{
			Name: "Error when fetching the Application Template for the given dynamic doc for a second time",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(5)

				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			webhookSvcFn: successfulStaticWebhookListAppTemplate,
			appTemplateVersionSvcFn: func() *automock.ApplicationTemplateVersionService {
				svc := &automock.ApplicationTemplateVersionService{}
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return([]*model.ApplicationTemplateVersion{}, nil).Once()
				svc.On("Create", txtest.CtxWithDBMatcher(), appTemplateID, fixAppTemplateVersionInput()).Return(appTemplateVersionID, nil)
				svc.On("ListByAppTemplateID", txtest.CtxWithDBMatcher(), appTemplateID).Return(fixAppTemplateVersions(), nil).Once()
				svc.On("GetByAppTemplateIDAndVersion", txtest.CtxWithDBMatcher(), appTemplateID, appTemplateVersionValue).Return(nil, testErr).Once()
				return svc
			},
			appTemplateSvcFn:    successAppTemplateGetSvc,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn:            successfulClientFetchForStaticDoc,
			processFnName:       processApplicationTemplateFnName,
			documentValidatorFn: successfulDocumentValidatorForStaticDocFn,
			ExpectedErr:         testErr,
		},
		{
			Name: "Success when there is ORD webhook on app template",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn: successfulAppSvc,
			tenantSvcFn: func() *automock.TenantService {
				tenantSvc := &automock.TenantService{}
				tenantSvc.On("GetLowestOwnerForResource", txtest.CtxWithDBMatcher(), resource.Application, appID).Return(tenantID, nil).Twice()
				tenantSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(&model.BusinessTenantMapping{ExternalTenant: externalTenantID}, nil).Twice()
				return tenantSvc
			},
			webhookConvFn:                    successfulWebhookConversion,
			webhookSvcFn:                     successfulAppTemplateTenantMappingOnlyCreation,
			bundleSvcFn:                      successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecRecreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			appTemplateSvcFn:                 successAppTemplateGetSvc,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionListForAppTemplateFlow,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				testResources := ord.Resource{
					Type:     resource.Application,
					ID:       testApplication.ID,
					Name:     testApplication.Name,
					ParentID: &appTemplateID,
				}
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResources, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{fixORDDocument()}, []string{}, baseURL, nil).Once()
				return client
			},
			processFnName:       processAppInAppTemplateContextFnName,
			labelSvcFn:          successfulLabelGetByKey,
			documentValidatorFn: successfulDocumentValidatorForApplicationFn,
		},
		{
			Name: "Error when synchronizing global resources from global registry should get them from DB and proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			bundleSvcFn:                      successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecCreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", mock.Anything).Return(map[string]bool{sapVendor: true}, nil).Once()
				return globalRegistrySvcFn
			},
			clientFn:            successfulClientFetch,
			processFnName:       processApplicationFnName,
			documentValidatorFn: successfulDocumentValidatorForApplicationFn,
			labelSvcFn:          successfulLabelGetByKey,
		},
		{
			Name: "Error when synchronizing global resources from global registry and get them from DB should proceed with the rest of the sync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:              successfulAppGet,
			tenantSvcFn:           successfulTenantSvc,
			webhookSvcFn:          successfulTenantMappingOnlyCreation,
			webhookConvFn:         successfulWebhookConversion,
			bundleSvcFn:           successfulBundleCreate,
			apiProcessorFn:        successfulAPIProcess,
			eventProcessorFn:      successfulEventProcess,
			entityTypeProcessorFn: successfulEntityTypeProcess,
			capabilityProcessorFn: successfulCapabilityProcess,

			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecCreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			globalRegistrySvcFn: func() *automock.GlobalRegistryService {
				globalRegistrySvcFn := &automock.GlobalRegistryService{}
				globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				globalRegistrySvcFn.On("ListGlobalResources", mock.Anything).Return(nil, errors.New("error")).Once()
				return globalRegistrySvcFn
			},
			clientFn:      successfulClientFetch,
			processFnName: processApplicationFnName,
			labelSvcFn:    successfulLabelGetByKey,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				docValidator.On("Validate", mock.Anything, []*ord.Document{fixORDDocument()}, baseURL, map[string]bool{}, []string{}).Return(nil, nil)
				return docValidator
			},
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
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(2)

				transact.On("Begin").Return(persistTx, nil).Once()
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
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(nil, []string{}, "", testErr)
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
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", txtest.CtxWithDBMatcher(), appID).Return(fixApplications()[0], nil).Twice()
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalTenantID")}).Return(nil).Once()
				return appSvc
			},
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			bundleSvcFn:                      successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			capabilityProcessorFn:            successfulCapabilityProcess,
			specSvcFn:                        successfulSpecRecreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
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
				appSvc.On("Update", txtest.CtxWithDBMatcher(), appID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr("ordLocalTenantID")}).Return(testErr).Once()
				return appSvc
			},
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulWebhookList,
			bundleRefSvcFn: successfulBundleReferenceFetchingOfBundleIDs,

			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.DescribedSystemInstance.LocalTenantID = str.Ptr("ordLocalTenantID")
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Resync resources for invalid ORD documents when event resource name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:       successfulAppGet,
			tenantSvcFn:    successfulTenantSvc,
			webhookSvcFn:   successfulTenantMappingOnlyCreation,
			webhookConvFn:  successfulWebhookConversion,
			bundleSvcFn:    successfulBundleCreate,
			apiProcessorFn: successfulAPIProcess,
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), []*model.EventDefinitionInput{sanitizedDoc.EventResources[1]}, mock.Anything).Return(fixEvents(), fixOneEventFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: successfulEntityTypeProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(doc)).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(doc)).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			capabilityProcessorFn: successfulCapabilityProcess,
			specSvcFn:             successfulSpecWithOneEventCreateAndUpdate,
			fetchReqFn:            successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:    successfulPackageProcess,
			productProcessorFn:    successfulProductProcess,
			vendorProcessorFn:     successfulVendorProcess,
			tombstoneProcessorFn:  successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixAPIsFetchRequests()[0:3]
				fetchReq = append(fetchReq, fixOneEventFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, fixVendors(), fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.EventResources[0].Name = "" // invalid document
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Run(func(args mock.Arguments) {
					docs := args.Get(1).([]*ord.Document)
					docs[0].EventResources = docs[0].EventResources[1:]
				}).Return(validatingORDDocsErr, nil)
				return docValidator
			},
			ExpectedErr: &ord.ProcessingError{ValidationErrors: validatingORDDocsErr},
		},
		{
			Name: "Resync resources for invalid ORD documents when bundle name is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(9)
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Times(2)
				return bundlesSvc
			},
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, bundles, fixPackages(), sanitizedDoc.APIResources, mock.Anything).Return(fixAPIs(), fixAPIsFetchRequests(), nil).Once()
				return apiProcessor
			},
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, bundles, fixPackages(), sanitizedDoc.EventResources, mock.Anything).Return(fixEvents(), fixEventsFetchRequests(), nil).Once()
				return eventProcessor
			},
			entityTypeProcessorFn: successfulEntityTypeProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = "" // invalid document
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(doc)).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = "" // invalid document
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(doc)).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			capabilityProcessorFn: successfulCapabilityProcess,
			specSvcFn:             successfulSpecCreateAndUpdate,
			fetchReqFn:            successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:    successfulPackageProcess,
			productProcessorFn:    successfulProductProcess,
			vendorProcessorFn: func() *automock.VendorProcessor {
				vendorProcessor := &automock.VendorProcessor{}
				vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Vendors).Return(fixVendors(), nil).Once()
				return vendorProcessor
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixAPIsFetchRequests()[0:3]
				fetchReq = append(fetchReq, fixEventsFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				var nilBundles []*model.Bundle
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, fixVendors(), fixProducts(), fixPackages(), nilBundles, fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = "" // invalid document
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Run(func(args mock.Arguments) {
					docs := args.Get(1).([]*ord.Document)
					docs[0].ConsumptionBundles = docs[0].ConsumptionBundles[1:]
				}).Return(validatingORDDocsErr, nil)
				return docValidator
			},
			ExpectedErr: &ord.ProcessingError{ValidationErrors: validatingORDDocsErr},
		},
		{
			Name: "Resync resources for invalid ORD documents when vendor ordID is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:              successfulAppGet,
			tenantSvcFn:           successfulTenantSvc,
			webhookSvcFn:          successfulTenantMappingOnlyCreation,
			webhookConvFn:         successfulWebhookConversion,
			bundleSvcFn:           successfulBundleCreate,
			apiProcessorFn:        successfulAPIProcess,
			eventProcessorFn:      successfulEventProcess,
			entityTypeProcessorFn: successfulEntityTypeProcess,
			capabilityProcessorFn: successfulCapabilityProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(doc)).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(doc)).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			specSvcFn:          successfulSpecCreateAndUpdate,
			fetchReqFn:         successfulFetchRequestFetchAndUpdate,
			packageProcessorFn: successfulPackageProcess,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn: func() *automock.VendorProcessor {
				vendorProcessor := &automock.VendorProcessor{}
				vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, []*model.VendorInput{sanitizedDoc.Vendors[1]}).Return(fixVendors(), nil).Once()
				return vendorProcessor
			},
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = "" // invalid document
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Run(func(args mock.Arguments) {
					docs := args.Get(1).([]*ord.Document)
					docs[0].Vendors = docs[0].Vendors[1:]
				}).Return(validatingORDDocsErr, nil)
				return docValidator
			},
			ExpectedErr: &ord.ProcessingError{ValidationErrors: validatingORDDocsErr},
		},
		{
			Name: "Resync resources for invalid ORD documents when product title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:              successfulAppGet,
			tenantSvcFn:           successfulTenantSvc,
			webhookSvcFn:          successfulTenantMappingOnlyCreation,
			webhookConvFn:         successfulWebhookConversion,
			bundleSvcFn:           successfulBundleCreate,
			apiProcessorFn:        successfulAPIProcess,
			eventProcessorFn:      successfulEventProcess,
			entityTypeProcessorFn: successfulEntityTypeProcess,
			capabilityProcessorFn: successfulCapabilityProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				doc := fixORDDocument()
				doc.Products[0].Title = "" // invalid document
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(doc)).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				doc := fixORDDocument()
				doc.Products[0].Title = "" // invalid document
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(doc)).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			specSvcFn:          successfulSpecCreateAndUpdate,
			fetchReqFn:         successfulFetchRequestFetchAndUpdate,
			packageProcessorFn: successfulPackageProcess,
			productProcessorFn: func() *automock.ProductProcessor {
				productProcessor := &automock.ProductProcessor{}
				productProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, []*model.ProductInput{}).Return(fixProducts(), nil).Once()
				return productProcessor
			},
			vendorProcessorFn:            successfulVendorProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Products[0].Title = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.Products[0].Title = "" // invalid document
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Run(func(args mock.Arguments) {
					docs := args.Get(1).([]*ord.Document)
					docs[0].Products = docs[0].Products[1:]
				}).Return(validatingORDDocsErr, nil)
				return docValidator
			},
			ExpectedErr: &ord.ProcessingError{ValidationErrors: validatingORDDocsErr},
		},
		{
			Name: "Resync resources for invalid ORD documents when package title is empty",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(6)
			},
			appSvcFn:            successfulAppGet,
			tenantSvcFn:         successfulTenantSvc,
			webhookSvcFn:        successfulWebhookList,
			webhookConvFn:       successfulWebhookConversion,
			fetchReqFn:          successfulFetchRequestFetchAndUpdate,
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Packages[0].Title = "" // invalid document
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.Packages[0].Title = "" // invalid document
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Run(func(args mock.Arguments) {
					docs := args.Get(1).([]*ord.Document)
					docs[0].Packages = docs[0].Packages[1:]
				}).Return(validatingORDDocsErr, nil)
				return docValidator
			},
			ExpectedErr: errors.New("\"validation_errors\":[{\"ordId\":\"ns:eventResource:test:v1\""),
		},
		{
			Name: "Does not resync resources if vendors processing fail",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:     successfulAppGet,
			tenantSvcFn:  successfulTenantSvc,
			webhookSvcFn: successfulWebhookList,
			vendorProcessorFn: func() *automock.VendorProcessor {
				vendorProcessor := &automock.VendorProcessor{}
				vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Vendors).Return(nil, testErr).Once()
				return vendorProcessor
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if products processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:          successfulAppGet,
			tenantSvcFn:       successfulTenantSvc,
			webhookSvcFn:      successfulWebhookList,
			vendorProcessorFn: successfulVendorProcess,
			productProcessorFn: func() *automock.ProductProcessor {
				productProcessor := &automock.ProductProcessor{}
				productProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Products).Return(nil, testErr).Once()
				return productProcessor
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if package processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(6)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(6)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(5)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulWebhookList,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: func() *automock.PackageProcessor {
				packageProcessor := &automock.PackageProcessor{}
				packageProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.Packages, mock.Anything).Return(nil, testErr).Once()
				return packageProcessor
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Fails to list bundles after resync",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(9)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(10)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle update fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(7)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulWebhookList,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(fixBundles(), nil).Once()
				bundlesSvc.On("UpdateBundle", txtest.CtxWithDBMatcher(), resource.Application, bundleID, bundleUpdateInputFromCreateInput(*sanitizedDoc.ConsumptionBundles[0]), bundlePreSanitizedHash).Return(testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle create fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(7)

				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulWebhookList,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", testErr).Once()
				return bundlesSvc
			},
			clientFn:                successfulClientFetch,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if bundle have different tenant mapping configuration",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulWebhookList,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithMultipleSameTypesFormat, credentialExchangeStrategyType, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithMultipleSameTypesFormat, credentialExchangeStrategyType, credentialExchangeStrategyType))

				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: errors.New("There are differences in the Credential Exchange Strategies for Tenant Mappings for application with ID testApp. They should be the same."),
		},
		{
			Name: "Does not resync resources if webhooks could not be enriched",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(8)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(7)
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
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be listed for application",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(8)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(8)
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
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be converted from graphql input to model input",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(8)

				transact.On("Begin").Return(persistTx, nil).Once()
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
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Does not resync resources if webhooks cannot be created",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(8)

				transact.On("Begin").Return(persistTx, nil).Once()
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
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulListTwiceAndCreateBundle,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Does recreate of tenant mapping webhooks when there are differences",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
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
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleCreateWithGenericParam,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			apiProcessorFn:               successfulAPIProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			eventProcessorFn:             successfulEventProcess,
			entityTypeProcessorFn:        successfulEntityTypeProcess,
			capabilityProcessorFn:        successfulCapabilityProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(doc)).Return(fixIntegrationDependencies(), nil).Once()
				return integrationDependencyProcessor
			},
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(doc)).Return(fixDataProducts(), nil).Once()
				return dataProductProcessor
			},
			specSvcFn:               successfulSpecCreateAndUpdate,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			fetchReqFn:              successfulFetchRequestFetchAndUpdate,
			labelSvcFn:              successfulLabelGetByKey,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			processFnName: processApplicationFnName,
		},
		{
			Name: "Does not recreate of tenant mapping webhooks when there are differences but deletion fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(12)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(9)
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
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn: func() *automock.BundleService {
				bundlesSvc := &automock.BundleService{}
				//bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("ListByApplicationIDNoPaging", txtest.CtxWithDBMatcher(), appID).Return(nil, nil).Once()
				bundlesSvc.On("CreateBundle", txtest.CtxWithDBMatcher(), resource.Application, appID, mock.Anything, mock.Anything).Return("", nil).Once()
				return bundlesSvc
			},
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, baseURL, nil)
				return client
			},
			capabilityProcessorFn: successfulCapabilityProcess,
			specSvcFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesWithCustomTypeFormat, credentialExchangeStrategyType))
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Does not resync resources if api process fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleUpdateForApplication,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.APIResources, mock.Anything).Return(nil, nil, testErr).Once()
				return apiProcessor
			},
			clientFn:                successfulClientFetch,
			eventProcessorFn:        successfulEventProcess,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources returns error if api spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleUpdateForApplication,
			bundleRefSvcFn:     successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiProcessorFn: func() *automock.APIProcessor {
				apiProcessor := &automock.APIProcessor{}
				apiProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.APIResources, mock.Anything).Return(fixAPIs(), fixFailedAPIFetchRequests2(), nil).Once()
				return apiProcessor
			},
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
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
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn:                         successfulClientFetch,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			labelSvcFn:                       successfulLabelGetByKey,
			processFnName:                    processApplicationFnName,
			documentValidatorFn:              successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                      errors.New("failed to process 3 specification fetch requests"),
		},
		{
			Name: "Does not resync resources if event process fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleUpdateForApplication,
			bundleRefSvcFn:     successfulBundleReferenceFetchingOfAPIBundleIDs,
			apiProcessorFn:     successfulAPIProcess,
			eventProcessorFn: func() *automock.EventProcessor {
				eventProcessor := &automock.EventProcessor{}
				eventProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixBundles(), fixPackages(), sanitizedDoc.EventResources, mock.Anything).Return(nil, nil, testErr).Once()
				return eventProcessor
			},
			capabilityProcessorFn:   successfulCapabilityProcess,
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Resync resources returns error if event spec refetch fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleUpdateForApplication,
			bundleRefSvcFn:     successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:     successfulAPIProcess,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).
					Times(len(fixAPI1SpecInputs(baseURL)))

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(nil).
					Times(len(fixAPI1SpecInputs(baseURL)))

				return specSvc
			},
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
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
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetch,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionList,
			labelSvcFn:                   successfulLabelGetByKey,
			processFnName:                processApplicationFnName,
			documentValidatorFn:          successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                  errors.New("failed to process 4 specification fetch requests"),
		},
		{
			Name: "Does not resync resources if entity type processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(10)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(10)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(10)
				return persistTx, transact
			},
			appSvcFn:           successfulAppGet,
			tenantSvcFn:        successfulTenantSvc,
			webhookSvcFn:       successfulTenantMappingOnlyCreation,
			webhookConvFn:      successfulWebhookConversion,
			productProcessorFn: successfulProductProcess,
			vendorProcessorFn:  successfulVendorProcess,
			packageProcessorFn: successfulPackageProcess,
			bundleSvcFn:        successfulBundleUpdateForApplication,
			bundleRefSvcFn:     successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:     successfulAPIProcess,
			eventProcessorFn:   successfulEventProcess,
			entityTypeProcessorFn: func() *automock.EntityTypeProcessor {
				entityTypeProcessor := &automock.EntityTypeProcessor{}
				entityTypeProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, sanitizedDoc.EntityTypes, fixPackages(), fixResourceHashesForDocument(fixORDDocument())).Return(nil, testErr).Once()
				return entityTypeProcessor
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if integration dependency processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:              successfulAppGet,
			tenantSvcFn:           successfulTenantSvc,
			webhookSvcFn:          successfulTenantMappingOnlyCreation,
			webhookConvFn:         successfulWebhookConversion,
			productProcessorFn:    successfulProductProcess,
			vendorProcessorFn:     successfulVendorProcess,
			packageProcessorFn:    successfulPackageProcess,
			bundleSvcFn:           successfulBundleUpdateForApplication,
			bundleRefSvcFn:        successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:        successfulAPIProcess,
			eventProcessorFn:      successfulEventProcess,
			entityTypeProcessorFn: successfulEntityTypeProcess,
			capabilityProcessorFn: successfulCapabilityProcess,
			integrationDependencyProcessorFn: func() *automock.IntegrationDependencyProcessor {
				integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
				integrationDependencyProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.IntegrationDependencies, fixResourceHashesForDocument(fixORDDocument())).Return(nil, testErr).Once()
				return integrationDependencyProcessor
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if capability process fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(10)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(10)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(9)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:              successfulAppGet,
			tenantSvcFn:           successfulTenantSvc,
			webhookSvcFn:          successfulTenantMappingOnlyCreation,
			webhookConvFn:         successfulWebhookConversion,
			productProcessorFn:    successfulProductProcess,
			vendorProcessorFn:     successfulVendorProcess,
			packageProcessorFn:    successfulPackageProcess,
			bundleSvcFn:           successfulBundleUpdateForApplication,
			bundleRefSvcFn:        successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:        successfulAPIProcess,
			eventProcessorFn:      successfulEventProcess,
			entityTypeProcessorFn: successfulEntityTypeProcess,
			capabilityProcessorFn: func() *automock.CapabilityProcessor {
				capabilityProcessor := &automock.CapabilityProcessor{}
				capabilityProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.Capabilities, mock.Anything).Return(nil, nil, testErr).Once()
				return capabilityProcessor
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if data product processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			packageProcessorFn:               successfulPackageProcess,
			bundleSvcFn:                      successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn: func() *automock.DataProductProcessor {
				dataProductProcessor := &automock.DataProductProcessor{}
				dataProductProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, fixPackages(), sanitizedDoc.DataProducts, fixResourceHashesForDocument(fixORDDocument())).Return(nil, testErr).Once()
				return dataProductProcessor
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if tombstone processing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			packageProcessorFn:               successfulPackageProcess,
			bundleSvcFn:                      successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
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
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Does not resync resources if resource deletion due to tombstone fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(10)
			},
			appSvcFn:                         successfulAppGet,
			tenantSvcFn:                      successfulTenantSvc,
			webhookSvcFn:                     successfulTenantMappingOnlyCreation,
			webhookConvFn:                    successfulWebhookConversion,
			bundleSvcFn:                      successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, fixVendors(), fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(nil, testErr).Once()
				return tombstonedResourcesDeleterFn
			},
			globalRegistrySvcFn:     successfulGlobalRegistrySvc,
			clientFn:                successfulClientFetch,
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			processFnName:           processApplicationFnName,
			documentValidatorFn:     successfulDocumentValidatorForApplicationFn,
			ExpectedErr:             testErr,
		},
		{
			Name: "Returns error when failing to open final transaction to commit fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(10)
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			fetchReqFn:                       successfulFetchRequestFetch,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn:                successfulVendorProcess,
			tombstoneProcessorFn:             successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn:     successfulTombstoneResDeleter,
			globalRegistrySvcFn:              successfulGlobalRegistrySvc,
			clientFn:                         successfulClientFetch,
			appTemplateVersionSvcFn:          successfulAppTemplateVersionList,
			labelSvcFn:                       successfulLabelGetByKey,
			processFnName:                    processApplicationFnName,
			documentValidatorFn:              successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                      errors.New("error while processing fetch request results"),
		},
		{
			Name: "Returns error when failing to find spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(10)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(nil, testErr).Once()
				return specSvc
			},
			fetchReqFn:                   successfulFetchRequestFetch,
			packageProcessorFn:           successfulPackageProcess,
			productProcessorFn:           successfulProductProcess,
			vendorProcessorFn:            successfulVendorProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetch,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionList,
			labelSvcFn:                   successfulLabelGetByKey,
			processFnName:                processApplicationFnName,
			documentValidatorFn:          successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                  errors.New("error while processing fetch request results"),
		},
		{
			Name: "Returns error when failing to update spec in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(10)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}

				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

				expectedSpecToUpdate := testSpec
				expectedSpecToUpdate.Data = &testSpecData
				specSvc.On("UpdateSpecOnly", txtest.CtxWithDBMatcher(), expectedSpecToUpdate).Return(testErr).Once()

				return specSvc
			},
			fetchReqFn:                   successfulFetchRequestFetch,
			packageProcessorFn:           successfulPackageProcess,
			productProcessorFn:           successfulProductProcess,
			vendorProcessorFn:            successfulVendorProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetch,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionList,
			labelSvcFn:                   successfulLabelGetByKey,
			processFnName:                processApplicationFnName,
			documentValidatorFn:          successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                  errors.New("error while processing fetch request results"),
		},
		{
			Name: "Returns error when failing to update fetch request in final transaction when trying to update and persist fetched specs",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsMultipleTimes(10)
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn: func() *automock.SpecService {
				specSvc := &automock.SpecService{}
				specSvc.On("GetByID", txtest.CtxWithDBMatcher(), mock.Anything, mock.Anything).Return(&testSpec, nil).Once()

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
			packageProcessorFn:           successfulPackageProcess,
			productProcessorFn:           successfulProductProcess,
			vendorProcessorFn:            successfulVendorProcess,
			tombstoneProcessorFn:         successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: successfulTombstoneResDeleter,
			globalRegistrySvcFn:          successfulGlobalRegistrySvc,
			clientFn:                     successfulClientFetch,
			appTemplateVersionSvcFn:      successfulAppTemplateVersionList,
			labelSvcFn:                   successfulLabelGetByKey,
			processFnName:                processApplicationFnName,
			documentValidatorFn:          successfulDocumentValidatorForApplicationFn,
			ExpectedErr:                  errors.New("error while processing fetch request results"),
		},
		{
			Name: "Success when resources are not in db and no SAP Vendor is declared in Documents should Create them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleCreate,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecCreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn: func() *automock.VendorProcessor {
				vendorProcessor := &automock.VendorProcessor{}
				doc := fixORDDocument()
				doc.Vendors = nil
				vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Vendors).Return(nil, nil).Once()
				return vendorProcessor
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixAPIsFetchRequests()[0:3]
				fetchReq = append(fetchReq, fixEventsFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				var nilVendors []*model.Vendor
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, nilVendors, fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.Vendors = nil
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			processFnName: processApplicationFnName,
		},
		{
			Name: "Success when resources are already in db and no SAP Vendor is declared in Documents should Update them as SAP Vendor is coming from the Global Registry",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(11)
			},
			appSvcFn:      successfulAppGet,
			tenantSvcFn:   successfulTenantSvc,
			webhookSvcFn:  successfulTenantMappingOnlyCreation,
			webhookConvFn: successfulWebhookConversion, bundleSvcFn: successfulBundleUpdateForApplication,
			bundleRefSvcFn:                   successfulBundleReferenceFetchingOfBundleIDs,
			apiProcessorFn:                   successfulAPIProcess,
			eventProcessorFn:                 successfulEventProcess,
			entityTypeProcessorFn:            successfulEntityTypeProcess,
			capabilityProcessorFn:            successfulCapabilityProcess,
			integrationDependencyProcessorFn: successfulIntegrationDependencyProcessing,
			dataProductProcessorFn:           successfulDataProductProcessing,
			specSvcFn:                        successfulSpecRecreateAndUpdate,
			fetchReqFn:                       successfulFetchRequestFetchAndUpdate,
			packageProcessorFn:               successfulPackageProcess,
			productProcessorFn:               successfulProductProcess,
			vendorProcessorFn: func() *automock.VendorProcessor {
				vendorProcessor := &automock.VendorProcessor{}
				doc := fixORDDocument()
				doc.Vendors = nil
				vendorProcessor.On("Process", txtest.CtxWithDBMatcher(), resource.Application, appID, doc.Vendors).Return(nil, nil).Once()
				return vendorProcessor
			},
			tombstoneProcessorFn: successfulTombstoneProcessing,
			tombstonedResourcesDeleterFn: func() *automock.TombstonedResourcesDeleter {
				tombstonedResourcesDeleterFn := &automock.TombstonedResourcesDeleter{}
				fetchReq := fixAPIsFetchRequests()[0:3]
				fetchReq = append(fetchReq, fixEventsFetchRequests()...)
				fetchReq = append(fetchReq, fixCapabilityFetchRequests()...)
				var nilVendors []*model.Vendor
				tombstonedResourcesDeleterFn.On("Delete", txtest.CtxWithDBMatcher(), resource.Application, nilVendors, fixProducts(), fixPackages(), fixBundles(), fixAPIs(), fixEvents(), fixEntityTypes(), fixCapabilities(), fixIntegrationDependencies(), fixDataProducts(), fixTombstones(), mock.Anything).Return(fetchReq, nil).Once()
				return tombstonedResourcesDeleterFn
			},
			globalRegistrySvcFn: successfulGlobalRegistrySvc,
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				doc := fixORDDocument()
				doc.Vendors = nil
				client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{doc}, []string{}, *doc.DescribedSystemInstance.BaseURL, nil)
				return client
			},
			appTemplateVersionSvcFn: successfulAppTemplateVersionList,
			labelSvcFn:              successfulLabelGetByKey,
			documentValidatorFn: func() *automock.Validator {
				docValidator := &automock.Validator{}
				doc := fixORDDocument()
				doc.Vendors = nil
				docValidator.On("Validate", mock.Anything, []*ord.Document{doc}, baseURL, map[string]bool{sapVendor: true}, []string{}).Return(nil, nil)
				return docValidator
			},
			processFnName: processApplicationFnName,
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
			apiProcessor := &automock.APIProcessor{}
			if test.apiProcessorFn != nil {
				apiProcessor = test.apiProcessorFn()
			}
			eventProcessor := &automock.EventProcessor{}
			if test.eventProcessorFn != nil {
				eventProcessor = test.eventProcessorFn()
			}
			entityTypeProcessor := &automock.EntityTypeProcessor{}
			if test.entityTypeProcessorFn != nil {
				entityTypeProcessor = test.entityTypeProcessorFn()
			}
			capabilityProcessor := &automock.CapabilityProcessor{}
			if test.capabilityProcessorFn != nil {
				capabilityProcessor = test.capabilityProcessorFn()
			}
			integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
			if test.integrationDependencyProcessorFn != nil {
				integrationDependencyProcessor = test.integrationDependencyProcessorFn()
			}
			dataProductProcessor := &automock.DataProductProcessor{}
			if test.dataProductProcessorFn != nil {
				dataProductProcessor = test.dataProductProcessorFn()
			}
			specSvc := &automock.SpecService{}
			if test.specSvcFn != nil {
				specSvc = test.specSvcFn()
			}
			fetchReqSvc := &automock.FetchRequestService{}
			if test.fetchReqFn != nil {
				fetchReqSvc = test.fetchReqFn()
			}
			packageProcessor := &automock.PackageProcessor{}
			if test.packageProcessorFn != nil {
				packageProcessor = test.packageProcessorFn()
			}
			productProcessor := &automock.ProductProcessor{}
			if test.productProcessorFn != nil {
				productProcessor = test.productProcessorFn()
			}
			vendorProcessor := &automock.VendorProcessor{}
			if test.vendorProcessorFn != nil {
				vendorProcessor = test.vendorProcessorFn()
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
			tombstonedResourcesDeleter := &automock.TombstonedResourcesDeleter{}
			if test.tombstonedResourcesDeleterFn != nil {
				tombstonedResourcesDeleter = test.tombstonedResourcesDeleterFn()
			}
			labelSvc := &automock.LabelService{}
			if test.labelSvcFn != nil {
				labelSvc = test.labelSvcFn()
			}
			var ordWebhookMappings []application.ORDWebhookMapping
			if test.webhookMappings != nil {
				ordWebhookMappings = test.webhookMappings
			}

			documentValidator := &automock.Validator{}
			if test.documentValidatorFn != nil {
				documentValidator = test.documentValidatorFn()
			}

			metrixCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiProcessor, eventProcessor, entityTypeProcessor, capabilityProcessor, integrationDependencyProcessor, dataProductProcessor, specSvc, fetchReqSvc, packageProcessor, productProcessor, vendorProcessor, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, tombstonedResourcesDeleter, labelSvc, ordWebhookMappings, nil, documentValidator)

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

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, entityTypeProcessor, integrationDependencyProcessor, dataProductProcessor, specSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc, tombstonedResourcesDeleter)
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
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResource, testWebhookForApplication, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, []string{}, baseURL, nil)
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
			apiProcessor := &automock.APIProcessor{}
			eventProcessor := &automock.EventProcessor{}
			capabilityProcessor := &automock.CapabilityProcessor{}
			integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
			dataProductProcessor := &automock.DataProductProcessor{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageProcessor := &automock.PackageProcessor{}
			productProcessor := &automock.ProductProcessor{}
			vendorProcessor := &automock.VendorProcessor{}
			tombstoneProcessor := &automock.TombstoneProcessor{}
			appTemplateVersionSvc := &automock.ApplicationTemplateVersionService{}
			appTemplateSvc := &automock.ApplicationTemplateService{}
			tombstonedResourcesDeleter := &automock.TombstonedResourcesDeleter{}

			documentValidator := &automock.Validator{}

			metrixCfg := ord.MetricsConfig{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiProcessor, eventProcessor, nil, capabilityProcessor, integrationDependencyProcessor, dataProductProcessor, specSvc, fetchReqSvc, packageProcessor, productProcessor, vendorProcessor, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, tombstonedResourcesDeleter, labelSvc, []application.ORDWebhookMapping{}, nil, documentValidator)
			err := svc.ProcessApplication(context.TODO(), test.appID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, integrationDependencyProcessor, specSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
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
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceAppTemplate, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, []string{}, baseURL, nil).Once()
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
			apiProcessor := &automock.APIProcessor{}
			eventProcessor := &automock.EventProcessor{}
			capabilityProcessor := &automock.CapabilityProcessor{}
			integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
			dataProductProcessor := &automock.DataProductProcessor{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageProcessor := &automock.PackageProcessor{}
			productProcessor := &automock.ProductProcessor{}
			vendorProcessor := &automock.VendorProcessor{}
			tombstoneProcessor := &automock.TombstoneProcessor{}
			tombstonedResourcesDeleter := &automock.TombstonedResourcesDeleter{}

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

			documentValidator := &automock.Validator{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metricsCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiProcessor, eventProcessor, nil, capabilityProcessor, integrationDependencyProcessor, dataProductProcessor, specSvc, fetchReqSvc, packageProcessor, productProcessor, vendorProcessor, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, tombstonedResourcesDeleter, labelSvc, []application.ORDWebhookMapping{}, nil, documentValidator)
			err := svc.ProcessApplicationTemplate(context.TODO(), test.appTemplateID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, integrationDependencyProcessor, specSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
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
		client.On("FetchOpenResourceDiscoveryDocuments", txtest.CtxWithDBMatcher(), testResourceApp, testWebhookForAppTemplate, emptyORDMapping, ordRequestObject).Return(ord.Documents{}, []string{}, baseURL, nil).Once()
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
			apiProcessor := &automock.APIProcessor{}
			eventProcessor := &automock.EventProcessor{}
			capabilityProcessor := &automock.CapabilityProcessor{}
			integrationDependencyProcessor := &automock.IntegrationDependencyProcessor{}
			dataProductProcessor := &automock.DataProductProcessor{}
			specSvc := &automock.SpecService{}
			fetchReqSvc := &automock.FetchRequestService{}
			packageProcessor := &automock.PackageProcessor{}
			productProcessor := &automock.ProductProcessor{}
			vendorProcessor := &automock.VendorProcessor{}
			tombstoneProcessor := &automock.TombstoneProcessor{}
			tombstonedResourcesDeleter := &automock.TombstonedResourcesDeleter{}

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

			documentValidator := &automock.Validator{}

			ordCfg := ord.NewServiceConfig(100, credentialExchangeStrategyTenantMappings)
			svc := ord.NewAggregatorService(ordCfg, metrixCfg, tx, appSvc, whSvc, bndlSvc, bndlRefSvc, apiProcessor, eventProcessor, nil, capabilityProcessor, integrationDependencyProcessor, dataProductProcessor, specSvc, fetchReqSvc, packageProcessor, productProcessor, vendorProcessor, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, whConverter, appTemplateVersionSvc, appTemplateSvc, tombstonedResourcesDeleter, labelSvc, []application.ORDWebhookMapping{}, nil, documentValidator)
			err := svc.ProcessAppInAppTemplateContext(context.TODO(), test.appTemplateID, test.appID)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, appSvc, whSvc, bndlSvc, integrationDependencyProcessor, specSvc, tombstoneProcessor, tenantSvc, globalRegistrySvcFn, client, labelSvc)
		})
	}
}

func successfulGlobalRegistrySvc() *automock.GlobalRegistryService {
	globalRegistrySvcFn := &automock.GlobalRegistryService{}
	globalRegistrySvcFn.On("SyncGlobalResources", mock.Anything).Return(map[string]bool{sapVendor: true}, nil).Once()
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

func fixAPIsFetchRequests() []*processor.OrdFetchRequest {
	api1SpecInput1 := fixAPI1SpecInputs(baseURL)[0]
	api1SpecInput2 := fixAPI1SpecInputs(baseURL)[1]
	api1SpecInput3 := fixAPI1SpecInputs(baseURL)[2]

	api2SpecInput1 := fixAPI2SpecInputs(baseURL)[0]
	api2SpecInput2 := fixAPI2SpecInputs(baseURL)[1]

	fr1 := fixFetchRequestFromFetchRequestInput(api1SpecInput1.FetchRequest, model.APISpecFetchRequestReference, "")
	fr2 := fixFetchRequestFromFetchRequestInput(api1SpecInput2.FetchRequest, model.APISpecFetchRequestReference, "")
	fr3 := fixFetchRequestFromFetchRequestInput(api1SpecInput3.FetchRequest, model.APISpecFetchRequestReference, "")
	fr4 := fixFetchRequestFromFetchRequestInput(api2SpecInput1.FetchRequest, model.APISpecFetchRequestReference, "")
	fr5 := fixFetchRequestFromFetchRequestInput(api2SpecInput2.FetchRequest, model.APISpecFetchRequestReference, "")

	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr3,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr4,
			RefObjectOrdID: api2ORDID,
		},
		{
			FetchRequest:   fr5,
			RefObjectOrdID: api2ORDID,
		},
	}
}

func fixEventsFetchRequests() []*processor.OrdFetchRequest {
	event1Spec := fixEvent1SpecInputs()[0]
	event2Spec := fixEvent2SpecInputs(baseURL)[0]

	fr1 := fixFetchRequestFromFetchRequestInput(event1Spec.FetchRequest, model.EventSpecFetchRequestReference, "")
	fr2 := fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, "")

	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: event1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: event2ORDID,
		},
	}
}

func fixFailedEventsFetchRequests() []*processor.OrdFetchRequest {
	fr1 := fixFailedFetchRequest()
	fr2 := fixFailedFetchRequest()

	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: event1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: event2ORDID,
		},
	}
}

func fixFailedAPIFetchRequests() []*processor.OrdFetchRequest {
	fr1 := fixFailedFetchRequest()
	fr2 := fixFailedFetchRequest()

	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: api2ORDID,
		},
	}
}

func fixFailedAPIFetchRequests2() []*processor.OrdFetchRequest {
	fr1 := fixFailedFetchRequest()
	fr2 := fixFailedFetchRequest()
	fr3 := fixFailedFetchRequest()
	fr4 := fixFailedFetchRequest()
	fr5 := fixFailedFetchRequest()

	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr3,
			RefObjectOrdID: api1ORDID,
		},
		{
			FetchRequest:   fr4,
			RefObjectOrdID: api2ORDID,
		},
		{
			FetchRequest:   fr5,
			RefObjectOrdID: api2ORDID,
		},
	}
}

func fixOneEventFetchRequests() []*processor.OrdFetchRequest {
	event2Spec := fixEvent2SpecInputs(baseURL)[0]
	fr := fixFetchRequestFromFetchRequestInput(event2Spec.FetchRequest, model.EventSpecFetchRequestReference, "")
	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr,
			RefObjectOrdID: event2ORDID,
		},
	}
}

func fixCapabilityFetchRequests() []*processor.OrdFetchRequest {
	capabilitySpec := fixCapabilitySpecInputs()[0]
	fr1 := fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, "")
	fr2 := fixFetchRequestFromFetchRequestInput(capabilitySpec.FetchRequest, model.EventSpecFetchRequestReference, "")
	return []*processor.OrdFetchRequest{
		{
			FetchRequest:   fr1,
			RefObjectOrdID: capability1ORDID,
		},
		{
			FetchRequest:   fr2,
			RefObjectOrdID: capability2ORDID,
		},
	}
}
