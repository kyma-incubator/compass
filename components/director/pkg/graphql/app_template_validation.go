package graphql

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
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

type CldFilterOperationType string

const (
	IncludeOperationType CldFilterOperationType = "include"
	ExcludeOperationType CldFilterOperationType = "exclude"
)

type CldFilter struct {
	Key       string
	Value     []string
	Operation CldFilterOperationType
}

type ProductIDFilterMapping struct {
	ProductID string      `json:"productID"`
	Filter    []CldFilter `json:"filter"`
}

// check for cldFilter and cldSystemRole presence and validate product ids
func labelsRuleFunc(value interface{}) error {
	labels, ok := value.(Labels)
	if !ok {
		return errors.New("value could not be cast to Labels object")
	}

	systemRolesLabel, hasCldSystemRoles := labels["cldSystemRole"]
	cldFilterLabel, hasCldFilter := labels["cldFilter"]
	if !hasCldSystemRoles && hasCldFilter {
		return errors.New("cld system role is required when cld filter is defined")
	}

	if !hasCldFilter {
		return nil
	}

	systemRoles := make([]string, 0)
	systemRolesLabelValue, ok := systemRolesLabel.([]interface{})
	if !ok {
		return errors.New("invalid format of cld system roles label")
	}

	for _, systemRoleValue := range systemRolesLabelValue {
		if systemRoleValueStr, ok := systemRoleValue.(string); ok {
			systemRoles = append(systemRoles, systemRoleValueStr)
		}
	}

	cldFilterLabelValue, ok := cldFilterLabel.([]interface{})
	if !ok {
		return errors.New("invalid format of cld filter label")
	}

	productIds := make([]string, 0)

	for _, cldFilterValue := range cldFilterLabelValue {
		filter, ok := cldFilterValue.(map[string]interface{})
		if !ok {
			return errors.New("invalid format of cld filter value")
		}

		productId, ok := filter["productId"]
		if !ok {
			return errors.New("missing productId in cld filter")
		}

		productIdStr, ok := productId.(string)
		if !ok {
			return errors.New("invalid format of productId value")
		}

		productIds = append(productIds, productIdStr)
	}

	systemRolesNumber := len(systemRoles)
	cldFilterProductIdsNumber := len(productIds)

	if systemRolesNumber != cldFilterProductIdsNumber {
		return errors.New("cld system roles number does not match the product ids number in cld filter")
	}

	sort.Strings(systemRoles)
	sort.Strings(productIds)

	fmt.Println(systemRoles, productIds)

	for i, s := range systemRoles {
		if s != productIds[i] {
			return errors.New("cld system roles dont match product ids in cld filter")
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
