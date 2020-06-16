package director

import "fmt"

const (
	consoleURLLabelKey = "runtime_consoleUrl"
	instanceIDLabelKey = "broker_instance_id"
)

type queryProvider struct{}

func (qp queryProvider) Runtime(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: runtime(id: "%s") {
	%s
	}
}`, runtimeID, runtimeStatusData())
}

func (qp queryProvider) SetRuntimeLabel(runtimeId, key, value string) string {
	return fmt.Sprintf(`mutation {
		result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: "%s") {
			%s
		}
	}`, runtimeId, key, value, labelData())
}

func (qp queryProvider) RuntimeLabels(runtimeID string) string {
	return fmt.Sprintf(`query {
	result: runtime(id: "%s") {
	labels
	}
}`, runtimeID)
}

func runtimeStatusData() string {
	return fmt.Sprintf(`id
			labels(key: "%s") 
			status{
				condition
			}`, consoleURLLabelKey)
}

func labelData() string {
	return `key
			value`
}
