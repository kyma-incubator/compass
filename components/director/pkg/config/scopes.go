package config

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

func (p *Provider) GetRequiredScopes(path string) ([]string, error) {
	val, err := p.getValueForJSONPath(path)
	if err != nil {
		if err == ValueNotFoundError {
			return nil, scope.RequiredScopesNotDefinedError
		}
		return nil, err
	}

	singleVal, ok := val.(string)
	if ok {
		return []string{singleVal}, nil
	}
	manyVals, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected scopes definition, should be string or list of strings, but was %T", val)
	}

	var scopes []string
	for _, val := range manyVals {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected scope value in a list, should be string but was %T", val)
		}
		scopes = append(scopes, strVal)

	}
	return scopes, nil
}
