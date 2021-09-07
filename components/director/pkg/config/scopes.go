package config

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func (p *Provider) GetRequiredScopes(path string) ([]string, error) {
	return p.getValues("scopes", path, true)
}

func (p *Provider) GetRequiredGrantTypes(path string) ([]string, error) {
	return p.getValues("grant_types", path, false)
}

func (p *Provider) getValues(valueType, path string, singeValueExpected bool) ([]string, error) {
	val, err := p.getValueForJSONPath(path)
	if err != nil {
		if apperrors.IsValueNotFoundInConfiguration(err) {
			return nil, apperrors.NewRequiredScopesNotDefinedError()
		}
		return nil, err
	}

	singleVal, ok := val.(string)
	if ok && singeValueExpected {
		return []string{singleVal}, nil
	}
	manyVals, ok := val.([]interface{})
	if !ok {
		errorMessageFormat := "unexpected %s definition, should be a list of strings, but was %T"
		if singeValueExpected {
			errorMessageFormat = "unexpected %s definition, should be string or list of strings, but was %T"
		}
		return nil, fmt.Errorf(errorMessageFormat, valueType, val)
	}

	scopes := make([]string, 0, len(manyVals))
	for _, val := range manyVals {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected %T value in a string list", val)
		}
		scopes = append(scopes, strVal)
	}
	return scopes, nil
}
