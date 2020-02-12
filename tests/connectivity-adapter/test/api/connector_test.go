package api

import (
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"testing"

	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit/director"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/connectivity-adapter/test/testkit"
	"github.com/stretchr/testify/require"
)

func TestConnector(t *testing.T) {
	appInput := directorSchema.ApplicationRegisterInput{
		Name:           "mytestapp2",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	descr := "test"
	runtimeInput := directorSchema.RuntimeInput{
		Name:        "myrunt2",
		Description: &descr,
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	config, err := testkit.ReadConfiguration()
	require.NoError(t, err)

	// TODO: what tenant to use
	client, err := director.NewClient(config.DirectorUrl, "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae", []string{"application:read", "application:write", "runtime:write", "runtime:read", "eventing:manage"})
	require.NoError(t, err)

	appID, err := client.CreateApplication(appInput)
	require.NoError(t, err)

	defer func() {
		err = client.DeleteApplication(appID)
		require.NoError(t, err)
	}()

	runtimeID, err := client.CreateRuntime(runtimeInput)
	require.NoError(t, err)

	defer func() {
		err = client.DeleteRuntime(runtimeID)
		require.NoError(t, err)
	}()

	err = client.SetDefaultEventing(runtimeID, appID, "www.events.com")
	require.NoError(t, err)

	tokenURL, err := client.GetOneTimeTokenUrl(appID)
	require.NotEmpty(t, tokenURL)
}
