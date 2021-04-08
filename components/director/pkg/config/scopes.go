package config

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func (p *Provider) GetRequiredScopes(path string) ([]string, error) {
	return p.getValues("scopes", path)
}

func (p *Provider) GetRequiredGrantTypes(path string) ([]string, error) {
	return p.getValues("grant_types", path)
}

func (p *Provider) getValues(valueType, path string) ([]string, error) {
	val, err := p.getValueForJSONPath(path)
	if err != nil {
		if apperrors.IsValueNotFoundInConfiguration(err) {
			return nil, apperrors.NewRequiredScopesNotDefinedError()
		}
		return nil, err
	}

	singleVal, ok := val.(string)
	if ok {
		return []string{singleVal}, nil
	}
	manyVals, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected %s definition, should be string or list of strings, but was %T", valueType, val)
	}

	var scopes []string
	for _, val := range manyVals {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected %T value in a string list", val)
		}
		scopes = append(scopes, strVal)

	}
	return scopes, nil
}
