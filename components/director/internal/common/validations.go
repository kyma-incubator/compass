package common

import (
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"regexp"
	"strings"
)

const (
	// MinTitleLength represents the minimal accepted length of the Title field
	MinTitleLength = 1
	// MaxTitleLength represents the maximal accepted length of the Title field
	MaxTitleLength = 255
	// MinDescriptionLength represents the minimal accepted length of the Description field
	MinDescriptionLength = 1
	// MaxDescriptionLength represents the maximal accepted length of the Description field
	MaxDescriptionLength = 5000
	// MinOrdIDLength represents the minimal accepted length of the OrdID field
	MinOrdIDLength = 1
	// MaxOrdIDLength represents the maximal accepted length of the OrdID field
	MaxOrdIDLength = 255

	// AspectApiResourceRegex represents the valid structure of the apiResource items in Integration Dependency Aspect
	AspectApiResourceRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(apiResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// AspectEventResourceRegex represents the valid structure of the eventResource items in Integration Dependency Aspect
	AspectEventResourceRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// EventResourceEventTypeRegex represents the valid structure of the event type items in event resource subset
	EventResourceEventTypeRegex = "^([a-z0-9A-Z]+(?:[.][a-z0-9A-Z]+)(?:[.][a-z0-9A-Z]+)+)\\.(v0|v[1-9][0-9]*)$"
	//AspectResourcesMinVersionRegex represents the valid structure of the minVersion for apiResources and eventResources in Aspect
	AspectResourcesMinVersionRegex = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	// AspectMsg represents the resource name for Aspect used in error message
	AspectMsg = "aspect"
)

// ValidateAspectApiResources validates the JSONB field `apiResources` in Aspect
func ValidateAspectApiResources(value interface{}) error {
	return ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"ordId": {
			validation.Required,
			validation.Length(MinOrdIDLength, MaxOrdIDLength),
			validation.Match(regexp.MustCompile(AspectApiResourceRegex)),
		},
		"minVersion": {
			validation.Match(regexp.MustCompile(AspectResourcesMinVersionRegex)),
		},
	})
}

// ValidateAspectEventResources validates the JSONB field `eventResources` in Aspect
func ValidateAspectEventResources(value interface{}) error {
	return ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"ordId": {
			validation.Required,
			validation.Length(MinOrdIDLength, MaxOrdIDLength),
			validation.Match(regexp.MustCompile(AspectEventResourceRegex)),
		},
		"minVersion": {
			validation.Match(regexp.MustCompile(AspectResourcesMinVersionRegex)),
		},
		"subset": {
			validation.By(validateAspectEventResourceSubset),
		},
	})
}

// validateAspectEventResourceSubset validates the Aspect Event Resource fields
func validateAspectEventResourceSubset(value interface{}) error {
	if value == nil {
		return nil
	}

	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var rawValue json.RawMessage = v

	return ValidateJSONArrayOfObjects(rawValue, map[string][]validation.Rule{
		"eventType": {
			validation.Required,
			validation.Match(regexp.MustCompile(EventResourceEventTypeRegex)),
		},
	})
}

// ValidateJSONArrayOfObjects validates JSON array of objects
func ValidateJSONArrayOfObjects(arr interface{}, elementFieldRules map[string][]validation.Rule, crossFieldRules ...func(gjson.Result) error) error {
	if arr == nil {
		return nil
	}

	jsonArr, ok := arr.(json.RawMessage)
	if !ok {
		return errors.New("should be json")
	}

	if len(jsonArr) == 0 {
		return nil
	}

	if !gjson.ValidBytes(jsonArr) {
		return errors.New("should be valid json")
	}

	parsedArr := gjson.ParseBytes(jsonArr)
	if !parsedArr.IsArray() {
		return errors.New("should be json array")
	}

	if len(parsedArr.Array()) == 0 {
		return nil
	}

	for _, el := range parsedArr.Array() {
		for field, rules := range elementFieldRules {
			if err := validation.Validate(el.Get(field).Value(), rules...); err != nil {
				return errors.Wrapf(err, "error validating field %s", field)
			}
			for _, f := range crossFieldRules {
				if err := f(el); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ValidateFieldMandatory validates the mandatory field
func ValidateFieldMandatory(value interface{}, resource string) error {
	switch v := value.(type) {
	case *bool:
		if v == nil {
			return errors.New(fmt.Sprintf("%s mandatory field is required", resource))
		}
		return nil
	default:
		return errors.New(fmt.Sprintf("%s mandatory field is not a boolean pointer", resource))
	}
}

func NoNewLines(s string) bool {
	return !strings.Contains(s, "\\n")
}
