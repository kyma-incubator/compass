package inputvalidation

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	k8svalidation "k8s.io/apimachinery/pkg/api/validation"
)

var (
	DNSName     = &dnsNameRule{}
	RuntimeName = &runtimeNameRule{}

	runtimeNameRgx = regexp.MustCompile(`^[a-zA-Z0-9-._]+$`).MatchString
)

type dnsNameRule struct{}
type runtimeNameRule struct{}

func (v *dnsNameRule) Validate(value interface{}) error {
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

func (v *runtimeNameRule) Validate(value interface{}) error {
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

	if !runtimeNameRgx(fmt.Sprintf("%v", value)) {
		return errors.New("must contain only alpha-numeric characters, \".\", \"_\" or \"-\"")
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
