package repo

import "fmt"

type Conditions []Condition
type Condition struct {
	Field string
	Val   string
}

func getAllArgs(tenant string, conditions Conditions) []interface{} {
	allArgs := []interface{}{tenant}
	for _, idAndVal := range conditions {
		allArgs = append(allArgs, idAndVal.Val)
	}
	return allArgs
}

func appendConditions(query string, conditions Conditions) string {
	var out string
	for idx, idAndVal := range conditions {
		out = fmt.Sprintf("%s AND %s = $%d", query, idAndVal.Field, idx+2)
	}
	return out
}
