package director

import "fmt"

const consoleURLLabelKey = "runtime/console_url"

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
