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

	client := director.NewClient("3e64ebae-38b5-46a0-b1ed-9ccee153a0ae", []string{"application:read", "application:write", "runtime:write", "runtime:read"})

	appInput := directorSchema.ApplicationRegisterInput{
		Name:           "myapp22",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &directorSchema.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	descr := "test"
	runtimeInput := directorSchema.RuntimeInput{
		Name:        "myrunt",
		Description: &descr,
		Labels: &directorSchema.Labels{
			"scenarios":                 []interface{}{"DEFAULT"},
			"runtime/event_service_url": []interface{}{"http://eventing.runtime2"},
		},
	}

	runtime, err := client.CreateRuntime(runtimeInput)
	require.Equal(t, runtime, "myrunt")

	app, err := client.CreateApplication(appInput)
	require.Equal(t, "myapp", app.Name)
}

// bdf4e4b0-680f-463f-8fad-60bd03802068
