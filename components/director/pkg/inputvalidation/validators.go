package inputvalidation

import (
	"reflect"
	"unicode"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/validation"
)

func ValidateName(value interface{}) error {
	s, isNil, err := ensureInputIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	if len(s) > 36 {
		return errors.New("must be no more than 36 characters")
	}
	if s[0] >= '0' && s[0] <= '9' {
		return errors.New("cannot start with digit")
	}
	if errorMsg := validation.NameIsDNSSubdomain(s, false); errorMsg != nil {
		return errors.Errorf("%v", errorMsg)
	}
	return nil
}

func ValidatePrintable(value interface{}) error {
	s, isNil, err := ensureInputIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	for _, r := range s {
		if !unicode.IsPrint(r) {
			return errors.New("cannot contain not printable characters")
		}
	}
	return nil
}

func ValidatePrintableWithWhitespace(value interface{}) error {
	s, isNil, err := ensureInputIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	for _, r := range s {
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return errors.New("cannot contain not printable or whitespace characters")
		}
	}
	return nil
}

func ensureInputIsString(in interface{}) (val string, isNil bool, err error) {
	t := reflect.ValueOf(in)
	if t.Kind() == reflect.Ptr {
		if t.IsNil() {
			return "", true, nil
		}
		t = t.Elem()
	}

	if t.Kind() != reflect.String {
		return "", false, errors.New("type has to be a string")
	}

	return t.String(), false, nil
}
