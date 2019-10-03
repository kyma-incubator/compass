package testkit

import "fmt"

type queryProvider struct{}

func (qp queryProvider) provisionRuntime(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: provisionRuntime(id: "%s", config: %s)
}`, runtimeID, config)
}

func (qp queryProvider) upgradeRuntime(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: upgradeRuntime(id: "%s", config: %s)
}`, runtimeID, config)
}

func (qp queryProvider) deprovisionRuntime(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: deprovisionRuntime(id: "%s")
}`, runtimeID)
}

func (qp queryProvider) reconnectRuntimeAgent(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: reconnectRuntimeAgent(id: "%s")
}`, runtimeID)
}

func (qp queryProvider) runtimeStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeStatus(id: "%s") {
	%s
	}
}`, operationID, runtimeStatusResult())
}

func (qp queryProvider) runtimeOperationStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeOperationStatus(id: "%s") {
	%s
	}
}`, operationID, operationStatusResult())
}

func runtimeStatusResult() string {
	return fmt.Sprintf(`lastOperationStatus { operation state message }
			runtimeConnectionStatus { status }
			runtimeConnectionConfig { kubeconfig }
			runtimeConfiguration { 
				clusterConfig { 
					%s
				} 
				kymaConfig { version modules } 
			}`, clusterConfig())
}

func clusterConfig() string {
	return fmt.Sprintf(`
		name 
		nodeCount 
		diskSize
		machineType
		computeZone 
		version 
		providerConfig {
			... on GardenerProviderConfig {
			  	targetProvider
				targetSecret
				autoScalerMin
				autoScalerMax
				maxSurge
				maxUnavailable
				additionalProperties
			}
			...  on GCPProviderConfig {
				additionalProperties
			}
			...  on AKSProviderConfig {
				additionalProperties
			}
		}
`)
}

func operationStatusResult() string {
	return `operation 
			state
			message`
}
