package scope

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/oliveagle/jsonpath"
	"github.com/pkg/errors"
)

func NewProvider(fileName string) *provider {
	return &provider{
		fileName: fileName,
	}
}

type provider struct {
	fileName     string
	cachedConfig map[string]interface{}
}

func (p *provider) Load() error {
	b, err := ioutil.ReadFile(p.fileName)
	if err != nil {
		return errors.Wrapf(err, "while reading file %s", p.fileName)
	}
	out := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(b),&out); err != nil {
		return errors.Wrap(err, "while unmarshalling YAML")
	}
	p.cachedConfig = out

	return nil

}

func (p *provider) GetRequiredScopes(path string) ([]string, error) {
	if p.cachedConfig == nil {
		return nil, errors.New("required scopes configuration not loaded")
	}
	jPath := fmt.Sprintf("$.%s", path)
	res, err := jsonpath.JsonPathLookup(p.cachedConfig, jPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while searching configuration using path %s", jPath)
	}

	if res == nil {
		return nil, RequiredScopesNotDefinedError
	}
	singleVal, ok := res.(string)
	if ok {
		return []string{singleVal}, nil
	}
	manyVals, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected scopes definition, should be string or list of strings, but was %T", res)
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


func (p *provider) GetAllScopes() ([]string, error) {
	var scopes []string
	for key, _ := range p.cachedConfig {
		operations, ok := p.cachedConfig[key].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type %T for %s; expected map[string]interface{}", p.cachedConfig[key], key)
		}
		for opKey, opScopes := range operations {
			if opScopes == nil {
				continue
			}

			values, ok := opScopes.([]interface{})
			if !ok {
				// try with string

				value, ok := opScopes.(string)
				if !ok {
					return nil, fmt.Errorf("invalid type %T for %s values; expected map[string]interface{} or string", opScopes, opKey)
				}

				scopes = append(scopes, value)
				continue
			}

			for _, val := range values {
				strVal, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf("invalid type %T for %s value; expected map[string]interface{}", val, opKey)
				}
				scopes = append(scopes, strVal)
			}
		}
	}

	uniqueScopes := strings.Unique(scopes)

	return uniqueScopes, nil
}