package graphql

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

// Validate missing godoc
func (i ApplicationTemplateInput) Validate() error {
	return validation.Errors{
		"Rule.ValidPlaceholders": validPlaceholders(i.Placeholders, i.ApplicationInput),
		"appInput":               validation.Validate(i.ApplicationInput),
		"name":                   validation.Validate(i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		"description":            validation.Validate(i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		"placeholders":           validation.Validate(i.Placeholders, validation.Each(validation.Required)),
		"accessLevel":            validation.Validate(i.AccessLevel, validation.Required, validation.In(ApplicationTemplateAccessLevelGlobal)),
		"webhooks":               validation.Validate(i.Webhooks, validation.By(webhooksRuleFunc)),
		"applicationNamespace":   validation.Validate(i.ApplicationNamespace, validation.Length(1, longStringLengthLimit)),
	}.Filter()
}

// Validate missing godoc
func (i ApplicationTemplateUpdateInput) Validate() error {
	return validation.Errors{
		"Rule.ValidPlaceholders": validPlaceholders(i.Placeholders, i.ApplicationInput),
		"appInput":               validation.Validate(i.ApplicationInput),
		"name":                   validation.Validate(i.Name, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		"description":            validation.Validate(i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		"placeholders":           validation.Validate(i.Placeholders, validation.Each(validation.Required)),
		"accessLevel":            validation.Validate(i.AccessLevel, validation.Required, validation.In(ApplicationTemplateAccessLevelGlobal)),
		"applicationNamespace":   validation.Validate(i.ApplicationNamespace, validation.Length(1, longStringLengthLimit)),
	}.Filter()
}

// Validate missing godoc
func (i PlaceholderDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
	)
}

// Validate missing godoc
func (i ApplicationFromTemplateInput) Validate() error {
	return validation.Errors{
		"Rule.UniquePlaceholders": i.ensureUniquePlaceholders(),
		"templateName":            validation.Validate(i.TemplateName, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		"values":                  validation.Validate(i.Values, validation.Each(validation.Required)),
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

// Validate missing godoc
func (i TemplateValueInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Placeholder, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.Value, validation.RuneLength(0, shortStringLengthLimit)),
	)
}

func webhooksRuleFunc(value interface{}) error {
	webhookInputs, ok := value.([]*WebhookInput)
	if !ok {
		return errors.New("value could not be cast to WebhookInput slice")
	}

	for _, webhookInput := range webhookInputs {
		if err := webhookInput.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func ensureUniquePlaceholders(placeholders []*PlaceholderDefinitionInput) error {
	keys := make(map[string]interface{})
	for _, item := range placeholders {
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

func ensurePlaceholdersUsed(placeholders []*PlaceholderDefinitionInput, appInput *ApplicationRegisterInput) error {
	placeholdersMarshalled, err := json.Marshal(appInput)
	if err != nil {
		return errors.Wrap(err, "while marshalling placeholders")
	}

	placeholdersString := string(placeholdersMarshalled)

	for _, value := range placeholders {
		if value == nil {
			continue
		}
		if !strings.Contains(placeholdersString, fmt.Sprintf("{{%s}}", value.Name)) {
			return errors.Errorf("application input does not use provided placeholder [name=%s]", value.Name)
		}
	}

	return nil
}

func validPlaceholders(placeholders []*PlaceholderDefinitionInput, appInput *ApplicationRegisterInput) error {
	if err := ensureUniquePlaceholders(placeholders); err != nil {
		return err
	}
	if err := ensurePlaceholdersUsed(placeholders, appInput); err != nil {
		return err
	}
	return nil
}
