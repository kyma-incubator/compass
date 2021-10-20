package fixtures

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FixCreateFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  createFormation(formation: {name: "%s"}){
					name
				  }
				}`, formationName))
}

func FixDeleteFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  deleteFormation(formation: {name: "%s"}){
					name
				  }
				}`, formationName))
}

func CleanupFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, formation string) {
	deleteRequest := FixDeleteFormationRequest(formation)

	var f graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &f)
	assert.NoError(t, err)
}
