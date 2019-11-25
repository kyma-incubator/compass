package inputvalidation

import (
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

type eachKeyRule []validation.Rule

// EachKey returns a validation rule that loops through a map and validates each key inside with the provided rules.
// An empty iterable is considered valid. Use the Required rule to make sure the iterable is not empty.
func EachKey(rules ...validation.Rule) *eachKeyRule {
	mr := eachKeyRule(rules)
	return &mr
}

func (v eachKeyRule) Validate(value interface{}) error {
	errs := validation.Errors{}

	t := reflect.ValueOf(value)
	switch t.Kind() {
	case reflect.Map:
		for _, k := range t.MapKeys() {
			val := getInterface(k)
			if err := validation.Validate(val, v...); err != nil {
				errs[getString(k)] = err
			}
		}
	default:
		return errors.New("must be a map")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func getInterface(value reflect.Value) interface{} {
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return value.Elem().Interface()
	default:
		return value.Interface()
	}
}

func getString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return ""
		}
		return value.Elem().String()
	default:
		return value.String()
	}
}
