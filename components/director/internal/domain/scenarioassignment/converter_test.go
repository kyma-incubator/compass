package scenarioassignment_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestFromInputGraphql(t *testing.T) {
	sut := scenarioassignment.NewConverter()
	t.Run("happy path", func(t *testing.T) {
		// WHEN
		actual := sut.FromInputGraphQL(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: scenarioName,
			Selector: &graphql.LabelSelectorInput{
				Key:   scenarioassignment.SubaccountIDKey,
				Value: externalTargetTenantID,
			},
		}, targetTenantID)
		// THEN
		assert.Equal(t, model.AutomaticScenarioAssignment{
			ScenarioName:   scenarioName,
			TargetTenantID: targetTenantID,
		}, actual)
	})
}

func TestToGraphQL(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.ToGraphQL(model.AutomaticScenarioAssignment{
		ScenarioName:   scenarioName,
		Tenant:         tenantID,
		TargetTenantID: targetTenantID,
	}, externalTargetTenantID)
	// THEN
	assert.Equal(t, graphql.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Selector: &graphql.Label{
			Key:   scenarioassignment.SubaccountIDKey,
			Value: externalTargetTenantID,
		},
	}, actual)
}

func TestToEntity(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.ToEntity(model.AutomaticScenarioAssignment{
		ScenarioName:   scenarioName,
		Tenant:         tenantID,
		TargetTenantID: targetTenantID,
	})

	// THEN
	assert.Equal(t, scenarioassignment.Entity{
		Scenario:       scenarioName,
		TenantID:       tenantID,
		TargetTenantID: targetTenantID,
	}, actual)
}

func TestFromEntity(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.FromEntity(scenarioassignment.Entity{
		Scenario:       scenarioName,
		TenantID:       tenantID,
		TargetTenantID: targetTenantID,
	})

	// THEN
	assert.Equal(t, model.AutomaticScenarioAssignment{
		ScenarioName:   scenarioName,
		Tenant:         tenantID,
		TargetTenantID: targetTenantID,
	}, actual)
}
