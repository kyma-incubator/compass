package director

type queryProvider struct{}

func (qp queryProvider) createRuntimeMutation () string {

	return "Query to create Runtime"
}

func (qp queryProvider) updateRuntimeMutation () string {

	return "Query to update Runtime"
}

func (qp queryProvider) deleteRuntimeMutation () string {

	return "Query to delete Runtime"
}

//func (qp queryProvider) setRuntimeMutation(runtimeId, key, value string) string {
//	return fmt.Sprintf(`mutation {
//		result: setRuntime(key: "%s", value: "%s") {
//			%s
//		}
//	}`, runtimeId, key, value, labelData())
//}
//
//func labelData() string {
//	return `key
//			value`
//}
