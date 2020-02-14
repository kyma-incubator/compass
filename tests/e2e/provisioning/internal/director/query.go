package director

import "fmt"

const instanceIDLabelKey = "broker_instance_id"

type queryProvider struct{}

func (qp queryProvider) Runtime(instanceID string) string {
	return fmt.Sprintf(`query {
	result: runtimes(filter: { key: "%s" query: "%s" }, first: 1, after: "") {
    data {
      id
	}
}`, instanceIDLabelKey, instanceID)
}
