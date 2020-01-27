package graphql

import (
	"encoding/json"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

func (i ApplicationTemplateInput) Validate() error {
	return validation.Errors{
		"Rule.ValidPlaceholders": i.validPlaceholders(),
		"Name":                   validation.Validate(i.Name, validation.Required, inputvalidation.DNSName),
		"Description":            validation.Validate(i.Description, validation.RuneLength(0, shortStringLengthLimit)),
		"Placeholders":           validation.Validate(i.Placeholders, validation.Each(validation.Required)),
		"AccessLevel":            validation.Validate(i.AccessLevel, validation.Required, validation.In(ApplicationTemplateAccessLevelGlobal)),
	}.Filter()
}

func (i ApplicationTemplateInput) validPlaceholders() error {
	if err := i.ensureUniquePlaceholders(); err != nil {
		return err
	}
	if err := i.ensurePlaceholdersUsed(); err != nil {
		return err
	}
	return nil
}

func (i ApplicationTemplateInput) ensureUniquePlaceholders() error {
	keys := make(map[string]interface{})
	for _, item := range i.Placeholders {
		if item == nil {
			continue
		}

		if _, exist := keys[item.Name]; exist {
			return errors.Errorf("placeholder [name=%s] not unique", item.Name)
		}

		keys[item.Name] = struct{}{}
	}
	return nil
}

func (i ApplicationTemplateInput) ensurePlaceholdersUsed() error {
	placeholdersMarshalled, err := json.Marshal(i.ApplicationInput)
	if err != nil {
		return errors.Wrap(err, "while marshalling placeholders")
	}

	placeholdersString := string(placeholdersMarshalled)

	for _, value := range i.Placeholders {
		if value == nil {
			continue
		}
		if !strings.Contains(placeholdersString, fmt.Sprintf("{{%s}}", value.Name)) {
			return errors.Errorf("application input does not use provided placeholder [name=%s]", value.Name)
		}
	}

	return nil
}

func (i PlaceholderDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.Description, validation.RuneLength(0, shortStringLengthLimit)),
	)
}

func (i ApplicationFromTemplateInput) Validate() error {
	return validation.Errors{
		"Rule.UniquePlaceholders": i.ensureUniquePlaceholders(),
		"TemplateName":            validation.Validate(i.TemplateName, validation.Required, inputvalidation.DNSName),
		"Values":                  validation.Validate(i.Values, validation.Each(validation.Required)),
	}.Filter()
}

func (i ApplicationFromTemplateInput) ensureUniquePlaceholders() error {
	keys := make(map[string]interface{})
	for _, item := range i.Values {
		if item == nil {
			continue
		}
		if _, exist := keys[item.Placeholder]; exist {
			return errors.Errorf("placeholder [name=%s] not unique", item.Placeholder)
		}

		keys[item.Placeholder] = struct{}{}
	}
	return nil
}

func (i TemplateValueInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Placeholder, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.Value, validation.RuneLength(0, shortStringLengthLimit)),
	)
}
