package caller

import "fmt"

type queryProvider struct{}

func (qp queryProvider) runtimeStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeStatus(id: "%s") {
		%s
	}
}`, operationID, runtimeStatusData())
}

func runtimeStatusData() string {
	return `runtimeConfiguration { 
				kubeconfig
			}`
}
