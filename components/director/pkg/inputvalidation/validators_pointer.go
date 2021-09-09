package inputvalidation

import (
	"reflect"

	"github.com/pkg/errors"
)

// ValidateExactlyOneNotNil validates if exactly one of passed pointers is not a nil.
func ValidateExactlyOneNotNil(errorMessage string, ptr interface{}, ptrs ...interface{}) error {
	ptrs = append(ptrs, ptr)

	ok, err := exactlyOneNotNil(ptrs)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New(errorMessage)
	}

	return nil
}

func exactlyOneNotNil(ptrs []interface{}) (bool, error) {
	notNilFound := 0
	for _, v := range ptrs {
		if v == nil {
			continue
		}
		if reflect.ValueOf(v).Kind() != reflect.Ptr {
			return false, errors.New("field is not a pointer")
		}
		if !reflect.ValueOf(v).IsNil() {
			notNilFound++
		}
	}
	return notNilFound == 1, nil
}
