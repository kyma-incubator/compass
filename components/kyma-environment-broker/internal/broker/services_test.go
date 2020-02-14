package broker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServices_Services(t *testing.T) {
	// given
	optComponentsProviderMock := &automock.OptionalComponentNamesProvider{}
	defer optComponentsProviderMock.AssertExpectations(t)

	optComponentsNames := []string{"monitoring", "kiali", "loki", "jaeger"}
	optComponentsProviderMock.On("GetAllOptionalComponentsNames").Return(optComponentsNames)

	servicesEndpoint := broker.NewServices(
		broker.Config{EnablePlans: []string{"gcp", "azure"}},
		optComponentsProviderMock,
		&broker.DumyDumper{},
	)

	// when
	services, err := servicesEndpoint.Services(context.TODO())

	// then
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Len(t, services[0].Plans, 2)

	// assert provisioning schema
	componentItem := services[0].Plans[0].Schemas.Instance.Create.Parameters["properties"].(map[string]interface{})["components"]
	componentJSON, err := json.Marshal(componentItem)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`
		{
		  "type": "array",
		  "items": {
			  "type": "string",
			  "enum": %s
		  }
		}`, toJSONList(optComponentsNames)), string(componentJSON))
}

func toJSONList(in []string) string {
	return fmt.Sprintf(`["%s"]`, strings.Join(in, `", "`))
}
