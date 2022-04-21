package scalar

import (
	"encoding/json"
	"io"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"
)

// WriteMarshalled missing godoc
func WriteMarshalled(in interface{}, w io.Writer) error {
	bytes, err := json.Marshal(in)
	if err != nil {
		return errors.Errorf("error with marshalling %T", in)
	}

	_, err = w.Write(bytes)
	if err != nil {
		return errors.Errorf("error with writing %T", in)
	}
	return nil
}

// ConvertToString missing godoc
func ConvertToString(in interface{}) (string, error) {
	if in == nil {
		return "", apperrors.NewInvalidDataError("input should not be nil")
	}

	value, ok := in.(string)
	if !ok {
		return "", errors.Errorf("unexpected input type: %T, should be string", in)
	}

	return value, nil
}

// ConvertToMapStringStringArray missing godoc
func ConvertToMapStringStringArray(in interface{}) (map[string][]string, error) {
	if in == nil {
		return nil, apperrors.NewInvalidDataError("input should not be nil")
	}

	result := make(map[string][]string)

	value, ok := in.(map[string][]string)
	if !ok {
		return nil, errors.Errorf("unexpected input type: %T, should be map[string][]string", in)
	}

	for k, v := range value {
		var strValues []string
		for _, item := range v {
			str, ok := interface{}(item).(string)
			if !ok {
				return nil, errors.Errorf("value `%+v` must be a string, not %T", item, item)
			}

			strValues = append(strValues, str)
		}

		result[k] = strValues
	}
	return result, nil
}
