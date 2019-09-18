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
	return fmt.Sprintf(`query {
	result: reconnectRuntimeAgent(id: %s) {
	%s
	}
}`, operationID, runtimeStatusResult)
}

func (qp queryProvider) runtimeOperationStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: reconnectRuntimeAgent(id: %s) {
	%s
	}
}`, operationID, operationStatusResult())
}

func runtimeStatusResult() string {
	return `lastOperationStatus { operation state message errors }
			runtimeConnectionStatus { status errors }
			runtimeConnectionConfig { kubeconfig }
			runtimeConfiguration { 
				clusterConfig { name size memory computeZone version infrastructureProvider } 
				kymaConfig { version modules } 
			}`
}

func operationStatusResult() string {
	return `operation 
			state
			message
			errors`
}
