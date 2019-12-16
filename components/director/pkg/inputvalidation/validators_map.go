/*
Part of this file is copied from https://github.com/go-ozzo/ozzo-validation/blob/master/each.go.
Below you can find its licence.

The MIT License (MIT)
Copyright (c) 2016, Qiang Xue

Permission is hereby granted, free of charge, to any person obtaining a copy of this software
and associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software
is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

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
	if t.Kind() == reflect.Ptr {
		if t.IsNil() {
			return nil
		}
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Map:
		for _, k := range t.MapKeys() {
			val := getInterface(k)
			if err := validation.Validate(val, v...); err != nil {
				errs[getString(k)] = err
			}
		}
	default:
		return errors.New("the value must be a map")
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
