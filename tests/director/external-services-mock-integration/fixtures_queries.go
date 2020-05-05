package external_services_mock_integration

import (
	"context"
	"testing"

	gcli "github.com/machinebox/graphql"

	"github.com/stretchr/testify/require"
)

//Application
func unregisterApplicationInTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string, tenant string) {
	req := fixDeleteApplicationRequest(t, id)
	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}
