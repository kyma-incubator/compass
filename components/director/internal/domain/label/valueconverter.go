package label

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func ValueToStringsSlice(value interface{}) ([]string, error) {
	_value, ok := value.([]interface{})
	if !ok {
		return nil, apperrors.NewInternalError("cannot convert label value to slice of strings")
	}

	var values = make([]string, len(_value))
	for idx, v := range _value {
		_v, ok := v.(string)
		if !ok {
			return nil, apperrors.NewInternalError("cannot cast label value as a string")
		}
		values[idx] = _v
	}

	return values, nil
}
