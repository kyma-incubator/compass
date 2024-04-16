package graphql

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/systemfetcher"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

const (
	// SlisFilterLabelKey is the name of the slis filter label for app template
	SlisFilterLabelKey = "slisFilter"
	// ProductIDKey is the name of the property relating to system roles in system payload
	ProductIDKey = "productId"
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
		"labels":                 validation.Validate(i.Labels, validation.By(labelsRuleFunc)),
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
		"webhooks":               validation.Validate(i.Webhooks, validation.By(webhooksRuleFunc)),
		"applicationNamespace":   validation.Validate(i.ApplicationNamespace, validation.Length(1, longStringLengthLimit)),
	}.Filter()
}

// Validate missing godoc
func (i PlaceholderDefinitionInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required, inputvalidation.DNSName),
		validation.Field(&i.Description, validation.RuneLength(0, descriptionStringLengthLimit)),
		validation.Field(&i.JSONPath, validation.RuneLength(0, jsonPathStringLengthLimit)),
	)
}

// Validate missing godoc
func (i ApplicationFromTemplateInput) Validate() error {
	return validation.Errors{
		"Rule.EitherPlaceholdersOrPlaceholdersPayloadExists": i.ensureEitherPlaceholdersOrPlaceholdersPayloadExists(),
		"Rule.UniquePlaceholders":                            i.ensureUniquePlaceholders(),
		"templateName":                                       validation.Validate(i.TemplateName, validation.Required, is.PrintableASCII, validation.Length(1, 100)),
		"values":                                             validation.Validate(i.Values, validation.Each(validation.Required)),
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

func (i ApplicationFromTemplateInput) ensureEitherPlaceholdersOrPlaceholdersPayloadExists() error {
	if (i.PlaceholdersPayload != nil) == (len(i.Values) != 0) {
		return errors.Errorf("one of values or placeholdersPayload should be provided")
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

func labelsRuleFunc(value interface{}) error {
	labels, ok := value.(Labels)
	if !ok {
		return errors.New("value could not be cast to Labels object")
	}

	systemRolesLabel, hasSystemRoles := labels[systemfetcher.ApplicationTemplateLabelFilter]
	slisFilterLabel, hasSlisFilter := labels[SlisFilterLabelKey]

	if !hasSystemRoles && hasSlisFilter {
		return errors.New("system role is required when slis filter is defined")
	}

	if !hasSystemRoles {
		return nil
	}

	systemRolesLabelValue, ok := systemRolesLabel.([]interface{})
	if !ok {
		return errors.New("invalid format of system roles label")
	}

	if len(systemRolesLabelValue) == 0 && hasSlisFilter {
		return errors.New("system role must not be empty when slis filter is defined")
	}

	if !hasSlisFilter {
		return nil
	}

	systemRoles, err := str.ConvertToStringArray(systemRolesLabelValue)
	if err != nil {
		return err
	}

	slisFilterLabelValue, ok := slisFilterLabel.([]interface{})
	if !ok {
		return errors.Errorf("invalid format of slis filter label")
	}

	productIds := make([]string, 0)

	for _, slisFilterValue := range slisFilterLabelValue {
		filter, ok := slisFilterValue.(map[string]interface{})
		if !ok {
			return errors.New("invalid format of slis filter value")
		}

		productID, ok := filter[ProductIDKey]
		if !ok {
			return errors.New("missing productId in slis filter")
		}

		productIDStr, ok := productID.(string)
		if !ok {
			return errors.New("invalid format of productId value")
		}

		productIds = append(productIds, productIDStr)
	}

	systemRolesCount := len(systemRoles)
	slisFilterProductIdsCount := len(productIds)

	if systemRolesCount != slisFilterProductIdsCount {
		return errors.New("system roles count does not match the product ids count in slis filter")
	}

	sort.Strings(systemRoles)
	sort.Strings(productIds)

	for i, systemRole := range systemRoles {
		if systemRole != productIds[i] {
			return errors.New("system roles don't match with product ids in slis filter")
		}
	}

	return nil
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

func ensurePlaceholdersUsed(placeholders []*PlaceholderDefinitionInput, appInputString string) error {
	for _, value := range placeholders {
		if value == nil {
			continue
		}
		if !strings.Contains(appInputString, fmt.Sprintf("{{%s}}", value.Name)) {
			return errors.Errorf("application input does not use provided placeholder [name=%s]", value.Name)
		}
	}

	return nil
}

func ensurePlaceholdersDefined(placeholders []*PlaceholderDefinitionInput, appInputString string) error {
	placeholdersMap := make(map[string]bool)
	for _, placeholder := range placeholders {
		if placeholder == nil {
			continue
		}

		placeholdersMap[placeholder.Name] = true
	}

	re := regexp.MustCompile(`{{([^\.}]*)}}`)
	matches := re.FindAllStringSubmatch(appInputString, -1)
	for _, match := range matches {
		placeholderName := match[1]
		if placeholderName == "" {
			return errors.New("Empty placeholder [name=] provided in the Application Input")
		}

		if _, ok := placeholdersMap[placeholderName]; !ok {
			return errors.Errorf("Placeholder [name=%s] is used in the application input but it is not defined in the Placeholders array", placeholderName)
		}
	}

	return nil
}

func validPlaceholders(placeholders []*PlaceholderDefinitionInput, appInput *ApplicationJSONInput) error {
	appInputMarshalled, err := json.Marshal(appInput)
	if err != nil {
		return errors.Wrap(err, "while marshalling application json input")
	}

	trimmedAppInput, err := trimAppInputFromWebhookTemplateProperties(appInputMarshalled)
	if err != nil {
		return err
	}

	appInputMarshalledWithoutWebhookTemplates, err := json.Marshal(trimmedAppInput)
	if err != nil {
		return errors.Wrap(err, "while marshalling application json input without webhook templates")
	}

	if err = ensureUniquePlaceholders(placeholders); err != nil {
		return err
	}
	if err = ensurePlaceholdersUsed(placeholders, string(appInputMarshalled)); err != nil {
		return err
	}
	if err = ensurePlaceholdersDefined(placeholders, string(appInputMarshalledWithoutWebhookTemplates)); err != nil {
		return err
	}

	return nil
}

func trimAppInputFromWebhookTemplateProperties(appInput []byte) (ApplicationJSONInput, error) {
	var appIn ApplicationJSONInput
	err := json.Unmarshal(appInput, &appIn)
	if err != nil {
		return ApplicationJSONInput{}, errors.Wrap(err, "while unmarshalling ApplicationJSONInput")
	}

	for idx := range appIn.Webhooks {
		appIn.Webhooks[idx].InputTemplate = nil
		appIn.Webhooks[idx].URLTemplate = nil
		appIn.Webhooks[idx].HeaderTemplate = nil
		appIn.Webhooks[idx].OutputTemplate = nil
		appIn.Webhooks[idx].StatusTemplate = nil
	}

	return appIn, nil
}
