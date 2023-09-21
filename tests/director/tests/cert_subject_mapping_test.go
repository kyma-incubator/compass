package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()

	subject            = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	subjectTwo         = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=3c0fe289-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	sortedSubject      = certs.SortSubject(subject)
	consumerType       = "Integration System"                   // should be a valid consumer type
	internalConsumerID = "e01a1918-5ee9-40c4-8ec7-e407264d43d2" // randomly chosen
	tenantAccessLevels = []string{"account", "global"}          // should be a valid tenant access level

	updatedSubject            = "C=DE, L=E2E-test-updated, O=E2E-Org, OU=TestRegion-updated, OU=E2E-Org-Unit-updated, OU=8e255922-6a2e-4677-a1a4-246ffcb391df, CN=E2E-test-cmp-updated"
	sortedUpdatedSubject      = certs.SortSubject(updatedSubject)
	updatedConsumerType       = "Runtime"                              // should be a valid consumer type
	updatedInternalConsumerID = "d5644469-7605-48a7-9f18-f5dee8805904" // randomly chosen
	updatedTntAccessLevels    = []string{"customer"}                   // should be a valid tenant access level
)

func TestCreateCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	csmGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(csmInput)
	require.NoError(t, err)

	certSubjectMappingReq := fixtures.FixCreateCertificateSubjectMappingRequest(csmGQLInput)
	SaveExample(t, certSubjectMappingReq.Query(), "create certificate subject mapping")
	csm := graphql.CertificateSubjectMapping{}

	t.Logf("Creating certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, certSubjectMappingReq, &csm)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csm)

	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)

	csmInput.Subject = sortedSubject // Assert that the subject is sorted
	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csm)
}

func TestDeleteCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

	deleteCertSubjectMappingReq := fixtures.FixDeleteCertificateSubjectMappingRequest(csmCreate.ID)
	SaveExample(t, deleteCertSubjectMappingReq.Query(), "delete certificate subject mapping")
	csmDelete := graphql.CertificateSubjectMapping{}

	t.Logf("Deleting certificate subject mapping with ID: %s, subject: %s, consumer type: %s and tenant access levels: %s", csmCreate.ID, subject, consumerType, tenantAccessLevels)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteCertSubjectMappingReq, &csmDelete)
	require.NoError(t, err)
	require.NotEmpty(t, csmDelete.ID)

	csmInput.Subject = sortedSubject // Assert that the subject is sorted
	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csmDelete)
}

func TestUpdateCertSubjectMapping(t *testing.T) {
	csmCreateInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreateInput)

	updatedCertSubjectMappingInput := fixtures.FixCertificateSubjectMappingInput(updatedSubject, updatedConsumerType, &updatedInternalConsumerID, updatedTntAccessLevels)
	csmUpdatedGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(updatedCertSubjectMappingInput)
	require.NoError(t, err)

	updateCertSubjectMappingReq := fixtures.FixUpdateCertificateSubjectMappingRequest(csmCreate.ID, csmUpdatedGQLInput)
	SaveExample(t, updateCertSubjectMappingReq.Query(), "update certificate subject mapping")
	csmUpdated := graphql.CertificateSubjectMapping{}

	t.Logf("Updating certificate subject mapping with ID: %s with new subject: %s and new consumer type: %s", csmCreate.ID, updatedSubject, updatedConsumerType)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateCertSubjectMappingReq, &csmUpdated)
	require.NoError(t, err)
	require.NotEmpty(t, csmUpdated.ID)

	updatedCertSubjectMappingInput.Subject = sortedUpdatedSubject // Assert that the subject is sorted
	assertions.AssertCertificateSubjectMapping(t, &updatedCertSubjectMappingInput, &csmUpdated)
}

func TestQuerySingleCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

	queryCertSubjectMappingReq := fixtures.FixQueryCertificateSubjectMappingRequest(csmCreate.ID)
	SaveExample(t, queryCertSubjectMappingReq.Query(), "query certificate subject mapping")
	csm := graphql.CertificateSubjectMapping{}

	t.Logf("Query certificate subject mapping by ID: %s", csmCreate.ID)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingReq, &csm)
	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)

	csmInput.Subject = sortedSubject // Assert that the subject is sorted
	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csm)
}

func TestQueryCertSubjectMappings(t *testing.T) {
	first := 100
	queryCertSubjectMappingsWithPaginationReq := fixtures.FixQueryCertificateSubjectMappingsRequestWithPageSize(first)
	SaveExample(t, queryCertSubjectMappingsWithPaginationReq.Query(), "query certificate subject mappings")
	currentCertSubjectMappings := graphql.CertificateSubjectMappingPage{}

	t.Log("Getting current certificate subject mappings...")
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingsWithPaginationReq, &currentCertSubjectMappings)
	require.NoError(t, err)
	t.Logf("Current number of certificate subject mappings is: %d", currentCertSubjectMappings.TotalCount)

	// Create first certificate subject mapping
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

	// Create second certificate subject mapping
	csmInput2 := fixtures.FixCertificateSubjectMappingInput(updatedSubject, updatedConsumerType, &updatedInternalConsumerID, updatedTntAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", updatedSubject, updatedConsumerType, updatedTntAccessLevels)

	var csmCreate2 graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate2)
	csmCreate2 = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput2)

	certSubjectMappings := graphql.CertificateSubjectMappingPage{}
	t.Log("Query certificate subject mappings...")
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingsWithPaginationReq, &certSubjectMappings)
	require.NoError(t, err)
	require.NotEmpty(t, certSubjectMappings)
	require.Equal(t, currentCertSubjectMappings.TotalCount+2, certSubjectMappings.TotalCount)
}

func TestQueryCertSubjectMappingWithNewlyCreatedSubjectMapping(t *testing.T) {
	certSubjcetMappingCN := "cert-subject-mapping-cn"
	certSubjectMappingCustomSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjcetMappingCN, -1)

	// We need an externally issued cert with a custom subject that will be used to create a certificate subject mapping through the GraphQL API,
	// which later will be loaded in-memory from the hydrator component
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               certSubjectMappingCustomSubject,
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	directorCertSecuredClientWithCustomCert := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	formationName := "cert-subject-mapping-formation"
	t.Logf("Creating formation with name: %q should fail due to missing scopes", formationName)
	var formation graphql.Formation
	createFormationReq := fixtures.FixCreateFormationRequest(formationName)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClientWithCustomCert, createFormationReq, &formation)
	defer fixtures.DeleteFormation(t, ctx, directorCertSecuredClientWithCustomCert, formationName)
	require.Error(t, err)
	require.Empty(t, formation)
	require.Contains(t, err.Error(), "insufficient scopes provided") // we expect that error because the consumer from the certificate doesn't have formation permissions

	// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
	certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubjectMappingCustomSubject, "/"), "/", ",")
	csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

	t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
	time.Sleep(conf.CertSubjectMappingResyncInterval)

	t.Logf("Creating formation with name: %q after certificate subject mapping is created and it should be successful", formationName)
	err = testctx.Tc.RunOperation(ctx, directorCertSecuredClientWithCustomCert, createFormationReq, &formation)
	defer fixtures.DeleteFormation(t, ctx, directorCertSecuredClientWithCustomCert, formationName)
	require.NoError(t, err)
	require.NotEmpty(t, formation)
	require.Equal(t, formationName, formation.Name)
	t.Log("The formation was successfully created.")
}
