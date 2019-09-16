package testkit

import "fmt"

type queryProvider struct{}

func (qp queryProvider) provisionRuntime(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: provisionRuntime(id: %s, config: %s) {
	id
	}
}`, runtimeID, config)
}

func (qp queryProvider) upgradeRuntime(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: upgradeRuntime(id: %s, config: %s) {
	id
	}
}`, runtimeID, config)
}

func (qp queryProvider) deprovisionRuntime(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: deprovisionRuntime(id: %s) {
	id
	}
}`, runtimeID)
}

func (qp queryProvider) reconnectRuntimeAgent(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: reconnectRuntimeAgent(id: %s) {
	id
	}
}`, runtimeID)
}

func (qp queryProvider) runtimeStatus(operationID string) string {
	return fmt.Sprintf(`mutation {
	result: reconnectRuntimeAgent(id: %s) {
	%s
	}
}`, operationID)
}

func (qp queryProvider) runtimeOperationStatus(operationID string) string {
	return fmt.Sprintf(`mutation {
	result: reconnectRuntimeAgent(id: %s) {
	%s
	}
}`, operationID, operationStatusRestult())
}

func runtimeStatusResult() string {
	return ``
}

func operationStatusRestult() string {
	return ``
}
