package scope

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/oliveagle/jsonpath"
	"github.com/pkg/errors"
	"io/ioutil"
)

var NotDefinedScopesError = errors.New("scopes are not defined")

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
	jsonRepresentation, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		return errors.Wrap(err, "while converting YAML to JSON")
	}

	out := map[string]interface{}{}
	err = json.Unmarshal(([]byte)(jsonRepresentation), &out)
	if err != nil {
		return errors.Wrap(err, "While unmarshalling JSON")
	}

	p.cachedConfig = out

	return nil

}

func (p *provider) GetRequiredScopes(scopesDefinition string) ([]string, error) {
	if p.cachedConfig == nil {
		return nil, errors.New("missing definition of required scopes")
	}
	path := fmt.Sprintf("$.%s", scopesDefinition)
	res, err := jsonpath.JsonPathLookup(p.cachedConfig, path)
	if err != nil {
		return nil, errors.Wrapf(err, "while searching configuration using path %s", path)
	}

	if res == nil {
		return nil, NotDefinedScopesError
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
			return nil, fmt.Errorf("unexpected scope value ina list, should be string but was %T", strVal)
		}
		scopes = append(scopes, strVal)

	}
	return scopes, nil
}
