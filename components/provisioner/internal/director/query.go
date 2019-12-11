package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) createRuntimeMutation (runtimeInput string) string {
	return fmt.Sprintf(`mutation {
	result: createRuntime(in: %s)`, runtimeInput)
}

func (qp queryProvider) updateRuntimeMutation (runtimeID, runtimeInput string) string {
	return fmt.Sprintf(`mutation {
	result: updateRuntime(id: %s, in: %s)`, runtimeID, runtimeInput)
}

func (qp queryProvider) deleteRuntimeMutation (runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: deleteRuntime(id: %s)`, runtimeID)
}
