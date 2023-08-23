package scenariogroups

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// Key is used for type of the key for scenarioGroups in context
type Key string

// ScenarioGroupsContextKey is the key of the value for scenarioGroups coming from header
const ScenarioGroupsContextKey Key = "scenarioGroups"

// LoadFromContext retrieves the value of the scenario groups in the context
func LoadFromContext(ctx context.Context) []model.ScenarioGroup {
	scenarioGroups := ctx.Value(ScenarioGroupsContextKey)
	if scenarioGroups == nil {
		return []model.ScenarioGroup{}
	}

	return scenarioGroups.([]model.ScenarioGroup)
}

// SaveToContext adds the value of scenario groups to the context
func SaveToContext(ctx context.Context, scenarioGroups []model.ScenarioGroup) context.Context {
	return context.WithValue(ctx, ScenarioGroupsContextKey, scenarioGroups)
}
