package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

func TestSelfRegisterRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	distinguishLblValue := "test-distinguish-value"

	selfRegisterSubaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestSelfRegisterSubaccount)

	// Build graphql director client configured with certificate
	clientKey, rawCertChain := certs.IssueExternalIssuerCertificate(t, conf.CA.Certificate, conf.CA.Key, selfRegisterSubaccountID)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, clientKey, rawCertChain)

	input := graphql.RuntimeInput{
		Name:   "runtime-self-register",
		Labels: graphql.Labels{conf.SelfRegisterDistinguishLabelKey: distinguishLblValue},
	}

	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(input)
	require.NoError(t, err)

	// WHEN
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)

	actualRtm := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, registerReq, &actualRtm)
	defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, tenant.TestSelfRegisterSubaccount, &actualRtm)

	require.NoError(t, err)
	strLbl, ok := actualRtm.Labels[conf.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, strLbl, distinguishLblValue)
}
