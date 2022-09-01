package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ReadConfigFile reads config from file
func ReadConfigFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("config path cannot be empty")
	}
	config, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read config file")
	}

	return string(config), nil
}

// ParseConfigToJSONMap parses configuration into json object
func ParseConfigToJSONMap(configData string) (map[string]gjson.Result, error) {
	if ok := gjson.Valid(configData); !ok {
		return nil, errors.New("failed to validate config data")
	}

	result := gjson.Parse(configData)
	return result.Map(), nil
}
