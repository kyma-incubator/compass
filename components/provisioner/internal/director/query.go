package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) createRuntimeMutation(runtimeInput string) string {
	return fmt.Sprintf(`mutation {
	result: registerRuntime(in: %s) { id } }`, runtimeInput)
}

func (qp queryProvider) deleteRuntimeMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: deleteRuntime(id: %s)}`, runtimeID)
}
