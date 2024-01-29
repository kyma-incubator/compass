package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

type GenerateOnetimeTokenForApplicationOperation struct {
	applicationID  string
	tenantID       string
	scenarioGroups string
	asserters      []asserters.Asserter
}

func NewGenerateOnetimeTokenForApplicationOperation(applicationID string, tenantID string) *GenerateOnetimeTokenForApplicationOperation {
	return &GenerateOnetimeTokenForApplicationOperation{applicationID: applicationID, tenantID: tenantID}
}

func (o *GenerateOnetimeTokenForApplicationOperation) WithScenarioGroups(scenarioGroups string) *GenerateOnetimeTokenForApplicationOperation {
	o.scenarioGroups = scenarioGroups
	return o
}

func (o *GenerateOnetimeTokenForApplicationOperation) WithAsserters(asserters ...asserters.Asserter) *GenerateOnetimeTokenForApplicationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *GenerateOnetimeTokenForApplicationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	headers := make(map[string]string)
	if o.scenarioGroups != "" {
		headers["scenario_groups"] = o.scenarioGroups
	}

	fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, gqlClient, o.tenantID, o.applicationID, headers)

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *GenerateOnetimeTokenForApplicationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
}

func (o *GenerateOnetimeTokenForApplicationOperation) Operation() Operation {
	return o
}
