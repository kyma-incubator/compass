package scenariogroups

import (
	"context"
)

// Key missing go doc
type Key string

// ScenarioGroupsContextKey missing godoc
const ScenarioGroupsContextKey Key = "scenarioGroups"

// LoadFromContext missing godoc
func LoadFromContext(ctx context.Context) []string {
	scenarioGroups := ctx.Value(ScenarioGroupsContextKey)
	if scenarioGroups == nil {
		return nil
	}

	return scenarioGroups.([]string)
}

// SaveToContext missing godoc
func SaveToContext(ctx context.Context, clientID []string) context.Context {
	return context.WithValue(ctx, ScenarioGroupsContextKey, clientID)
}
