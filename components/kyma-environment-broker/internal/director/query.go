package director

import "fmt"

const consoleURLLabelKey = "runtime_consoleUrl"

type queryProvider struct{}

func (qp queryProvider) Runtime(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: runtime(id: "%s") {
	%s
	}
}`, runtimeID, runtimeStatusData())
}

func runtimeStatusData() string {
	return fmt.Sprintf(`id
			labels(key: "%s") 
			status{
				condition
			}`, consoleURLLabelKey)
}
