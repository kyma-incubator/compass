package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateCertificateSubjectMapping(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.CertificateSubjectMappingInput) graphql.CertificateSubjectMapping {
	csmGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(in)
	require.NoError(t, err)

	certSubjectMappingReq := FixCreateCertificateSubjectMappingRequest(csmGQLInput)
	csm := graphql.CertificateSubjectMapping{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, certSubjectMappingReq, &csm)
	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)

	return csm
}

func FixCertificateSubjectMappingInput(subject, consumerType string, internalConsumerID *string, tenantAccessLevels []string) graphql.CertificateSubjectMappingInput {
	return graphql.CertificateSubjectMappingInput{
		Subject:            subject,
		ConsumerType:       consumerType,
		InternalConsumerID: internalConsumerID,
		TenantAccessLevels: tenantAccessLevels,
	}
}

func CleanupCertificateSubjectMapping(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, csm *graphql.CertificateSubjectMapping) *graphql.CertificateSubjectMapping {
	deleteRequest := FixDeleteCertificateSubjectMappingRequest(csm.ID)

	result := graphql.CertificateSubjectMapping{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &result)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &result
}

func FindCertSubjectMappingForApplicationTemplate(t *testing.T, ctx context.Context, gqlClient *gcli.Client, appTemplateID string) *graphql.CertificateSubjectMapping {
	after := ""
	for {
		queryCertSubjectMappingReq := FixQueryCertificateSubjectMappingsRequestWithPagination(300, after)
		currentCertSubjectMappings := graphql.CertificateSubjectMappingPage{}

		t.Log("Getting current certificate subject mappings...")
		err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryCertSubjectMappingReq, &currentCertSubjectMappings)
		require.NoError(t, err)

		after = string(currentCertSubjectMappings.PageInfo.EndCursor)

		for _, csm := range currentCertSubjectMappings.Data {
			if str.PtrStrToStr(csm.InternalConsumerID) == appTemplateID {
				return csm
			}
		}

		if !currentCertSubjectMappings.PageInfo.HasNextPage {
			return nil
		}
	}
}
