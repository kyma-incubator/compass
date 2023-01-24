package tests

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	ctx = context.Background()

	subject = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	consumerType = "Integration System"
	internalConsumerID = "e01a1918-5ee9-40c4-8ec7-e407264d43d2"
	tenantAccessLevels = []string{"global"}

	updatedSubject = "C=DE, L=E2E-test-updated, O=E2E-Org, OU=TestRegion-updated, OU=E2E-Org-Unit-updated, OU=8e255922-6a2e-4677-a1a4-246ffcb391df, CN=E2E-test-cmp-updated"
	updatedConsumerType = "Runtime"
	updatedInternalConsumerID = "d5644469-7605-48a7-9f18-f5dee8805904"
	updatedTntAccessLevels = []string{"customer"}
)

func TestCreateCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	csmGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(csmInput)
	require.NoError(t, err)

	certSubjectMappingReq := fixtures.FixCreateCertificateSubjectMappingRequest(csmGQLInput)
	saveExample(t, certSubjectMappingReq.Query(), "create certificate subject mapping")
	csm := graphql.CertificateSubjectMapping{}

	t.Logf("Creating certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, certSubjectMappingReq, &csm)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csm.ID)

	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)
	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csm)
}

func TestDeleteCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)

	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	csmCreate := fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreate.ID)

	deleteCertSubjectMappingReq := fixtures.FixDeleteCertificateSubjectMappingRequest(csmCreate.ID)
	saveExample(t, deleteCertSubjectMappingReq.Query(), "delete certificate subject mapping")
	csmDelete := graphql.CertificateSubjectMapping{}

	t.Logf("Deleting certificate subject mapping with ID: %s, subject: %s, consumer type: %s and tenant access levels: %s", csmCreate.ID, subject, consumerType, tenantAccessLevels)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteCertSubjectMappingReq, &csmDelete)
	require.NoError(t, err)
	require.NotEmpty(t, csmDelete.ID)

	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csmDelete)
}

func TestUpdateCertSubjectMapping(t *testing.T) {
	csmCreateInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)

	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	csmCreate := fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreateInput)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreate.ID)

	updatedCertSubjectMappingInput := fixtures.FixCertificateSubjectMappingInput(updatedSubject, updatedConsumerType, &updatedInternalConsumerID, updatedTntAccessLevels)
	csmUpdatedGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(updatedCertSubjectMappingInput)
	require.NoError(t, err)

	updateCertSubjectMappingReq := fixtures.FixUpdateCertificateSubjectMappingRequest(csmCreate.ID, csmUpdatedGQLInput)
	saveExample(t, updateCertSubjectMappingReq.Query(), "update certificate subject mapping")
	csmUpdated := graphql.CertificateSubjectMapping{}

	t.Logf("Updating certificate subject mapping with ID: %s with new subject: %s and new consumer type: %s", csmCreate.ID, updatedSubject, updatedConsumerType)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateCertSubjectMappingReq, &csmUpdated)
	require.NoError(t, err)
	require.NotEmpty(t, csmUpdated.ID)

	assertions.AssertCertificateSubjectMapping(t, &updatedCertSubjectMappingInput, &csmUpdated)
}

func TestQueryCertSubjectMapping(t *testing.T) {
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)

	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	csmCreate := fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreate.ID)

	queryCertSubjectMappingReq := fixtures.FixQueryCertificateSubjectMappingRequest(csmCreate.ID)
	saveExample(t, queryCertSubjectMappingReq.Query(), "query certificate subject mapping")
	csm := graphql.CertificateSubjectMapping{}

	t.Logf("Query certificate subject mapping by ID: %s", csmCreate.ID)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingReq, &csm)
	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)

	assertions.AssertCertificateSubjectMapping(t, &csmInput, &csm)
}

func TestQueryCertSubjectMappings(t *testing.T) {
	first := 100
	queryCertSubjectMappingsWithPaginationReq := fixtures.FixQueryCertificateSubjectMappingsRequestWithPageSize(first)
	saveExample(t, queryCertSubjectMappingsWithPaginationReq.Query(), "query certificate subject mappings")
	certSubjectMappings := graphql.CertificateSubjectMappingPage{}

	t.Log("Query certificate subject mappings")
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingsWithPaginationReq, &certSubjectMappings)
	require.NoError(t, err)
	require.Equal(t, 0, certSubjectMappings.TotalCount)

	// Create first certificate subject mapping
	csmInput := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", subject, consumerType, tenantAccessLevels)
	csmCreate := fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreate.ID)

	// Create second certificate subject mapping
	csmInput2 := fixtures.FixCertificateSubjectMappingInput(updatedSubject, updatedConsumerType, &updatedInternalConsumerID, updatedTntAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", updatedSubject, updatedConsumerType, updatedTntAccessLevels)
	csmCreate2 := fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput2)
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmCreate2.ID)

	t.Log("Query certificate subject mappings")
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingsWithPaginationReq, &certSubjectMappings)
	require.NoError(t, err)
	require.NotEmpty(t, certSubjectMappings)
	require.Equal(t, 2, certSubjectMappings.TotalCount)
	require.Subset(t, certSubjectMappings.Data, []*graphql.CertificateSubjectMapping{
		csmCreate,
		csmCreate2,
	})
}
