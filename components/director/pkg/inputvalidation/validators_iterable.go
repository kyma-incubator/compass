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
	"errors"
	"reflect"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation"
)

type eachRule []validation.Rule

// Each returns a validation rule that loops through an iterable (map, slice or array)
// and validates each value inside with the provided rules.
// An empty iterable is considered valid. Use the Required rule to make sure the iterable is not empty.
func Each(rules ...validation.Rule) *eachRule {
	r := eachRule(rules)
	return &r
}

// Loops through the given iterable and calls the Ozzo Validate() method for each value.
func (r eachRule) Validate(value interface{}) error {
	errs := validation.Errors{}

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		for _, k := range v.MapKeys() {
			val := r.getInterface(v.MapIndex(k))
			if err := validation.Validate(val, r...); err != nil {
				errs[r.getString(k)] = err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			val := r.getInterface(v.Index(i))
			if err := validation.Validate(val, r...); err != nil {
				errs[strconv.Itoa(i)] = err
			}
		}
	default:
		return errors.New("must be an iterable (map, slice or array) or a pointer to iterable")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (r *eachRule) getInterface(value reflect.Value) interface{} {
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

func (r *eachRule) getString(value reflect.Value) string {
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
