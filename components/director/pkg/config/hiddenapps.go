package config

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const applicationHideSelectorsPath = "applicationHideSelectors"

// GetApplicationHideSelectors missing godoc
func (p *Provider) GetApplicationHideSelectors() (map[string][]string, error) {
	val, err := p.getValueForJSONPath(applicationHideSelectorsPath)
	if err != nil {
		if apperrors.IsValueNotFoundInConfiguration(err) {
			return nil, nil
		}
		return nil, err
	}

	selectorsMap, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected application hide selectors definition, should be a map, but was %T", val)
	}

	selectors := make(map[string][]string)
	for key, values := range selectorsMap {
		valuesList, ok := values.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected application hide selector values definition for key %s, should be a list, but was %T", key, values)
		}
		var v []string
		for _, value := range valuesList {
			valueString, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected application hide selector value definition for key %s, should be a string, but was %T", key, value)
			}
			v = append(v, valueString)
		}
		selectors[key] = v
	}

	return selectors, nil
}
