package provisioner

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

func (qp queryProvider) deprovisionRuntime(runtimeID, credentialsInput string) string {
	return fmt.Sprintf(`mutation {
	result: deprovisionRuntime(id: "%s", credentials: "%s")
}`, runtimeID, credentialsInput)
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
}`, operationID, runtimeStatusData())
}

func (qp queryProvider) runtimeOperationStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeOperationStatus(id: "%s") {
	%s
	}
}`, operationID, operationStatusData())
}

func runtimeStatusData() string {
	return fmt.Sprintf(`lastOperationStatus { operation state message }
			runtimeConnectionStatus { status }
			runtimeConfiguration { 
				kubeconfig
				clusterConfig { 
					%s
				} 
				kymaConfig { version modules } 
			}`, clusterConfig())
}

func clusterConfig() string {
	return fmt.Sprintf(`
		... on GardenerConfig {
			name 
			kubernetesVersion
			projectName
			nodeCount 
			volumeSize
			diskType
			machineType
			region
		  	targetProvider
			targetSecret
			zone
			cidr
			autoScalerMin
			autoScalerMax
			maxSurge
			maxUnavailable
		}
		...  on GCPConfig {
			name 
			kubernetesVersion
			projectName
			numberOfNodes 
			bootDiskSize
			machineType
			region
			zone
		}
`)
}

func operationStatusData() string {
	return `id
			operation 
			state
			message
			runtimeID`
}
