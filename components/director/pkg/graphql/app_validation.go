package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"os"
	"strconv"
)

const appNormalizationEnvVar = "APP_ENABLE_APP_NAME_NORMALIZATION"

var appNameNormalizationEnabled bool

func init() {
	appNameNormalizationEnabled = false

	value, isSet := os.LookupEnv(appNormalizationEnvVar)
	if isSet {
		parsedValue, err := strconv.ParseBool(value)
		if err == nil {
			appNameNormalizationEnabled = parsedValue
		}
	}
}

func (i ApplicationRegisterInput) Validate() error {
	fieldRules := make([]*validation.FieldRules, 0)
	if !appNameNormalizationEnabled {
		fieldRules = append(fieldRules, validation.Field(&i.Name, validation.Required, inputvalidation.DNSName))
	}

	fieldRules = append(fieldRules, validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)))
	fieldRules = append(fieldRules, validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)))
	fieldRules = append(fieldRules, validation.Field(&i.Labels, inputvalidation.EachKey(validation.Required, validation.Match(alphanumericUnderscoreRegexp))))
	fieldRules = append(fieldRules, validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)))
	fieldRules = append(fieldRules, validation.Field(&i.Webhooks, validation.Each(validation.Required)))

	return validation.ValidateStruct(&i, fieldRules...)
}

func (i ApplicationUpdateInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.ProviderName, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.HealthCheckURL, inputvalidation.IsURL, validation.RuneLength(0, longStringLengthLimit)),
	)
}
