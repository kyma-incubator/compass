package label

import (
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

func ValueToStringsSlice(labelValue interface{}) ([]string, error) {
	var result []string

	switch value := labelValue.(type) {
	case []string:
		result = value
	case []interface{}:
		convertedScenarios, err := str.InterfaceSliceToStringSlice(value)
		if err != nil {
			return nil, errors.Wrap(err, "while casting label value (slice of interfaces to array of slice)")
		}
		result = convertedScenarios
	default:
		return nil, errors.Errorf("value is invalid type: %T", labelValue)
	}

	return result, nil
}

func UniqueScenarios(scenarios, newScenarios []string) []string {
	scenarios = append(scenarios, newScenarios...)
	return str.Unique(scenarios)
}
