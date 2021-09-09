package inputvalidation

import (
	"errors"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/asaskevich/govalidator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type urlValidator struct{}

// IsURL missing godoc
var IsURL = &urlValidator{}

const protocolRegex = govalidator.URLSchema

const errMsg = "must be a valid URL"

// Validate missing godoc
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
		return apperrors.InternalErrorFrom(err, "error during checking URL: %s", value)
	}

	if !matched {
		return errors.New(errMsg)
	}

	return validation.Validate(value, is.URL)
}
