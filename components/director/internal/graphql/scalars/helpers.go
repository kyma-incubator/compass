package scalars

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

func writeResponse(in interface{}, w io.Writer) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return errors.Errorf("Error with marshalling %v", reflect.TypeOf(in))
	}

	_, err = w.Write(bytes)
	if err != nil {
		return errors.Errorf("Error with writing %v", reflect.TypeOf(in))
	}
	return nil
}

func convertToString(in interface{}) (string, error) {
	if in == nil {
		return "", errors.New("error")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.New("Error with unmarshalling Tenant")
	}

	return value, nil
}

func convertToMapStringArrString(in interface{}) (map[string][]string, error) {
	if in == nil {
		return nil, errors.New("Error: input is nil")
	}

	val, ok := in.(string)
	if !ok {
		return nil, errors.New("Error: input is not a string")
	}

	var result map[string][]string

	err := json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, errors.Errorf("Error with converting string to map[string][]string: %v", err)
	}

	return result, nil
}
