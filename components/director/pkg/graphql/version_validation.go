package graphql

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/common"
)

// Validate missing godoc
func (i VersionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Value, validation.Required, validation.RuneLength(1, longStringLengthLimit), validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&i.Deprecated, validation.NotNil),
		validation.Field(&i.DeprecatedSince, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.ForRemoval, validation.NotNil),
	)
}
