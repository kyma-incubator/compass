package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) createRuntimeMutation(runtimeInput string) string {
	return fmt.Sprintf(`mutation {
	result: registerRuntime(in: %s) { id } }`, runtimeInput)
}

func (qp queryProvider) getRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`query {
    result: runtime(id: "%s") {
         id name description labels
}}`, runtimeID)
}

func (qp queryProvider) updateRuntimeMutation(runtimeID, runtimeInput string) string {
	return fmt.Sprintf(`mutation {
    result: updateRuntime(id: "%s" in: %s) {
		id
}}`, runtimeID, runtimeInput)
}

func (qp queryProvider) deleteRuntimeMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: unregisterRuntime(id: "%s") {
		id
}}`, runtimeID)
}

func (qp queryProvider) requestOneTimeTokeneMutation(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: requestOneTimeTokenForRuntime(id: "%s") {
		token connectorURL
}}`, runtimeID)
}
