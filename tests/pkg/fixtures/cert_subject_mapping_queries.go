package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateCertificateSubjectMapping(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.CertificateSubjectMappingInput) *graphql.CertificateSubjectMapping {
	csmGQLInput, err := testctx.Tc.Graphqlizer.CertificateSubjectMappingInputToGQL(in)
	require.NoError(t, err)

	certSubjectMappingReq := FixCreateCertificateSubjectMappingRequest(csmGQLInput)
	csm := graphql.CertificateSubjectMapping{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, certSubjectMappingReq, &csm)
	require.NoError(t, err)
	require.NotEmpty(t, csm.ID)

	return &csm
}

func FixCertificateSubjectMappingInput(subject, consumerType string, internalConsumerID *string, tenantAccessLevels []string) graphql.CertificateSubjectMappingInput {
	return graphql.CertificateSubjectMappingInput{
		Subject:            subject,
		ConsumerType:       consumerType,
		InternalConsumerID: internalConsumerID,
		TenantAccessLevels: tenantAccessLevels,
	}
}

func CleanupCertificateSubjectMapping(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.CertificateSubjectMapping {
	deleteRequest := FixDeleteCertificateSubjectMappingRequest(id)

	csm := graphql.CertificateSubjectMapping{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &csm)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &csm
}
