package inputvalidation

import (
	"reflect"
	"unicode"

	k8svalidation "k8s.io/apimachinery/pkg/api/validation"

	"github.com/pkg/errors"
)

var (
	Name                    = &nameRule{}
	Printable               = &printableRule{}
	PrintableWithWhitespace = &printableWithWhitespaceRule{}
)

type nameRule struct{}

func (v *nameRule) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
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
	if errorMsg := k8svalidation.NameIsDNSSubdomain(s, false); errorMsg != nil {
		return errors.Errorf("%v", errorMsg)
	}
	return nil
}

type printableRule struct{}

func (v printableRule) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
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

type printableWithWhitespaceRule struct{}

func (v printableWithWhitespaceRule) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
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

func ensureIsString(in interface{}) (val string, isNil bool, err error) {
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
