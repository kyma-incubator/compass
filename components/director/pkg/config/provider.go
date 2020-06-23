package config

import (
	"fmt"
	"io/ioutil"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/ghodss/yaml"
	"github.com/oliveagle/jsonpath"
	"github.com/pkg/errors"
)

func NewProvider(fileName string) *Provider {
	return &Provider{
		fileName: fileName,
	}
}

type Provider struct {
	fileName     string
	cachedConfig map[string]interface{}
}

func (p *Provider) Load() error {
	b, err := ioutil.ReadFile(p.fileName)
	if err != nil {
		return errors.Wrapf(err, "while reading file %s", p.fileName)
	}
	out := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(b), &out); err != nil {
		return errors.Wrap(err, "while unmarshalling YAML")
	}
	p.cachedConfig = out

	return nil

}

func (p *Provider) getValueForJSONPath(path string) (interface{}, error) {
	if p.cachedConfig == nil {
		return nil, apperrors.NewInternalError("required configuration not loaded")
	}
	jPath := fmt.Sprintf("$.%s", path)
	res, err := jsonpath.JsonPathLookup(p.cachedConfig, jPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while searching configuration using path %s", jPath)
	}

	if res == nil {
		return nil, apperrors.NewValueNotFoundInConfigurationError()
	}

	return res, nil
}
