package scalars

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

func writeResponse(in interface{}, w io.Writer) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return errors.Errorf("error with marshalling %T", in)
	}

	_, err = w.Write(bytes)
	if err != nil {
		return errors.Errorf("Error with writing %T", in)
	}
	return nil
}

func convertToString(in interface{}) (string, error) {
	if in == nil {
		return "", errors.New("input should not be nil")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.Errorf("unexpected input type: %T, should be map[string][]string", in)
	}

	return value, nil
}

func convertToMapStringStringArray(in interface{}) (map[string][]string, error) {
	result := make(map[string][]string)

	value, ok := in.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("unexpected input type: %T, should be map[string][]string", in)
	}

	for k, v := range value {
		val, ok := v.([]string)
		if !ok {
			return nil, errors.Errorf("given value `%T` must be a string array", v)
		}
		result[k] = val
	}
	return result, nil
}
