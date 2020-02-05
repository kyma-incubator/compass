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
	_, err := testkit.ReadConfiguration()
	require.NoError(t, err)

	client := director.NewClient("3e64ebae-38b5-46a0-b1ed-9ccee153a0ae", []string{"application:read", "application:write", "runtime:write", "runtime:read", "eventing:manage"})

	appInput := directorSchema.ApplicationRegisterInput{
		Name:           "myapp1",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	descr := "test"
	runtimeInput := directorSchema.RuntimeInput{
		Name:        "myrunt1",
		Description: &descr,
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	runtime, err := client.CreateRuntime(runtimeInput)
	require.NotEmpty(t, runtime)

	app, err := client.CreateApplication(appInput)
	require.NotEmpty(t, app.ID)

	err = client.SetDefaultEventing(runtime, app.ID, "www.events.com")
	require.NoError(t, err)

	tokenURL, err := client.GetOneTimeTokenUrl(app.ID)
	require.NotEmpty(t, tokenURL)
}
