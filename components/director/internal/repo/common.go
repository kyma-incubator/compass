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

func appendEnumeratedConditions(query string, startIdx int, conditions Conditions) string {
	out := query
	for idx, idAndVal := range conditions {
		out = fmt.Sprintf("%s AND %s = $%d", out, idAndVal.Field, idx+startIdx)
	}
	return out
}
