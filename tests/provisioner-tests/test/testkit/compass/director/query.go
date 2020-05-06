package director

import "fmt"

type queryProvider struct{}

func (qp queryProvider) getRuntimeQuery(runtimeID string) string {
	return fmt.Sprintf(`
		query {
			result: runtime(id: "%s") {
				id name description labels status {condition}
			}
		}`, runtimeID)
}
