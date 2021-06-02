package inputvalidation

import (
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func Validate(validatable Validatable) error {

	var typeName string
	split := strings.Split(fmt.Sprintf("%T", validatable), ".")
	if len(split) > 1 {
		typeName = split[1]
	} else {
		typeName = split[0]
	}

	err := validatable.Validate()
	if err != nil {
		switch value := err.(type) {
		case validation.Errors:
			return apperrors.NewInvalidDataErrorWithFields(value, typeName)
		case validation.InternalError:
			return apperrors.InternalErrorFrom(value, "while validating")
		case apperrors.Error:
			return err
		default:
			return apperrors.NewInternalError(fmt.Sprintf("%+v error is not handled", value))
		}
	}
	return nil
}
