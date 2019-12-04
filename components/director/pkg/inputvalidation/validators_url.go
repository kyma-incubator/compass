package inputvalidation

import (
	"regexp"

	"github.com/asaskevich/govalidator"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/pkg/errors"
)

type urlValidator struct{}

var IsURL = &urlValidator{}

const protocolRegex = govalidator.URLSchema

const errMsg = "is not valid URL"

func (v *urlValidator) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	matched, err := regexp.Match(protocolRegex, []byte(s))
	if err != nil {
		return errors.Wrapf(err, "error during checking URL: %s", value)
	}

	if !matched {
		return errors.New(errMsg)
	}

	return validation.Validate(value, is.URL)
}
