package resync

import (
	"encoding/json"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// GlobalAccountRegex determines whether event entity type is global account
const GlobalAccountRegex = "^GLOBALACCOUNT_.*|GlobalAccount"

// EventsPage represents a page returned by the Events API - its payload,
// along with the mappings required for converting the page to a set of tenants
type EventsPage struct {
	FieldMapping                 TenantFieldMapping
	MovedSubaccountsFieldMapping MovedSubaccountsFieldMapping
	ProviderName                 string
	Payload                      []byte
}

// GetMovedSubaccounts parses the data from the page payload to MovedSubaccountMappingInput
func (ep EventsPage) GetMovedSubaccounts() []model.MovedSubaccountMappingInput {
	eds := ep.getEventsDetails()
	mappings := make([]model.MovedSubaccountMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToMovedSubaccount(detail)
		if err != nil {
			log.D().Warnf("Error: %s. Could not convert tenant: %s", err.Error(), string(detail))
			continue
		}

		mappings = append(mappings, *mapping)
	}

	return mappings
}

// GetTenantMappings parses the data from the page payload to BusinessTenantMappingInput
func (ep EventsPage) GetTenantMappings(eventsType EventsType) []model.BusinessTenantMappingInput {
	eds := ep.getEventsDetails()
	tenants := make([]model.BusinessTenantMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToTenant(eventsType, detail)
		if err != nil {
			log.D().Warnf("Error: %s. Could not convert tenant: %s", err.Error(), string(detail))
			continue
		}

		tenants = append(tenants, *mapping)
	}

	return tenants
}

func (ep EventsPage) getEventsDetails() [][]byte {
	tenantDetails := make([][]byte, 0)
	gjson.GetBytes(ep.Payload, ep.FieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		entityType := event.Get(ep.FieldMapping.EntityTypeField)
		globalAccountGUID := event.Get(ep.FieldMapping.GlobalAccountGUIDField)
		details := event.Get(ep.FieldMapping.DetailsField).Map()
		details[ep.FieldMapping.EntityTypeField] = entityType
		details[ep.FieldMapping.GlobalAccountKey] = globalAccountGUID
		allDetails := make(map[string]interface{})
		for key, result := range details {
			switch result.Type {
			case gjson.String:
				allDetails[key] = result.String()
			case gjson.Number:
				allDetails[key] = result.Float()
			case gjson.True:
				allDetails[key] = true
			case gjson.False:
				allDetails[key] = false
			case gjson.Null:
				allDetails[key] = nil
			default:
				log.D().Warnf("Unknown property type %s", result.Type)
			}
		}
		currentTenantDetails, err := json.Marshal(allDetails)
		if err != nil {
			log.D().Errorf("failed to marshal tenant details: %v", err)
			return false
		}
		tenantDetails = append(tenantDetails, currentTenantDetails)
		return true
	})
	return tenantDetails
}

func (ep EventsPage) eventDataToMovedSubaccount(eventData []byte) (*model.MovedSubaccountMappingInput, error) {
	jsonPayload := string(eventData)
	if !gjson.Valid(jsonPayload) {
		return nil, errors.Errorf("invalid json Payload")
	}

	id, ok := gjson.GetBytes(eventData, ep.MovedSubaccountsFieldMapping.SubaccountID).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.MovedSubaccountsFieldMapping.SubaccountID)
	}

	source, ok := gjson.GetBytes(eventData, ep.MovedSubaccountsFieldMapping.SourceTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.MovedSubaccountsFieldMapping.SourceTenant)
	}

	target, ok := gjson.GetBytes(eventData, ep.MovedSubaccountsFieldMapping.TargetTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.MovedSubaccountsFieldMapping.TargetTenant)
	}

	nameResult := gjson.Get(jsonPayload, ep.FieldMapping.NameField)
	if !nameResult.Exists() {
		return nil, invalidFieldFormatError(ep.FieldMapping.NameField)
	}

	subdomain := gjson.Get(jsonPayload, ep.FieldMapping.SubdomainField)
	if !subdomain.Exists() {
		log.D().Warnf("Missig or invalid format of field: %s for tenant with ID: %s", ep.FieldMapping.SubdomainField, id)
	}

	subaccountInput, err := constructSubaccountTenant(jsonPayload, nameResult.String(), subdomain.String(), id, ep)
	if err != nil {
		return nil, err
	}

	return &model.MovedSubaccountMappingInput{
		TenantMappingInput: *subaccountInput,
		SubaccountID:       id,
		SourceTenant:       source,
		TargetTenant:       target,
	}, nil
}

func (ep EventsPage) eventDataToTenant(eventType EventsType, eventData []byte) (*model.BusinessTenantMappingInput, error) {
	jsonPayload := string(eventData)
	if !gjson.Valid(jsonPayload) {
		return nil, errors.Errorf("invalid json Payload")
	}
	if eventType == CreatedAccountType && ep.FieldMapping.DiscriminatorField != "" {
		discriminatorResult := gjson.Get(jsonPayload, ep.FieldMapping.DiscriminatorField)
		if !discriminatorResult.Exists() {
			return nil, invalidFieldFormatError(ep.FieldMapping.DiscriminatorField)
		}
		if discriminatorResult.String() != ep.FieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, err := determineTenantID(jsonPayload, ep.FieldMapping)
	if err != nil {
		return nil, err
	}

	nameResult := gjson.Get(jsonPayload, ep.FieldMapping.NameField)
	if !nameResult.Exists() {
		log.D().Warnf("Missing or invalid format of name field: %s for tenant with ID: %s", ep.FieldMapping.NameField, id)
	}

	subdomain := gjson.Get(jsonPayload, ep.FieldMapping.SubdomainField)
	if !subdomain.Exists() {
		log.D().Warnf("Missing or invalid format of subdomain field: %s for tenant with ID: %s", ep.FieldMapping.SubdomainField, id)
	}

	entityType := gjson.Get(jsonPayload, ep.FieldMapping.EntityTypeField)
	if !entityType.Exists() {
		return nil, invalidFieldFormatError(ep.FieldMapping.EntityTypeField)
	}

	globalAccountRegex := regexp.MustCompile(GlobalAccountRegex)
	if globalAccountRegex.MatchString(entityType.String()) {
		return constructGlobalAccountTenant(jsonPayload, nameResult.String(), subdomain.String(), id, ep), nil
	} else {
		return constructSubaccountTenant(jsonPayload, nameResult.String(), subdomain.String(), id, ep)
	}
}

func constructGlobalAccountTenant(jsonPayload, name, subdomain, externalTenant string, ep EventsPage) *model.BusinessTenantMappingInput {
	parentID := ""
	customerIDResult := gjson.Get(jsonPayload, ep.FieldMapping.CustomerIDField)
	if !customerIDResult.Exists() {
		log.D().Warnf("Missig or invalid format of field: %s for tenant with id: %s", ep.FieldMapping.CustomerIDField, externalTenant)
	} else {
		parentID = customerIDResult.String()
	}
	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Parent:         parentID,
		Subdomain:      subdomain,
		Region:         "",
		Type:           tenant.TypeToStr(tenant.Account),
		Provider:       ep.ProviderName,
	}
}

func constructSubaccountTenant(jsonPayload, name, subdomain, externalTenant string, ep EventsPage) (*model.BusinessTenantMappingInput, error) {
	regionField := gjson.Get(jsonPayload, ep.FieldMapping.RegionField)
	if !regionField.Exists() {
		log.D().Warnf("Missing or invalid format of region field: %s for tenant with ID: %s", ep.FieldMapping.RegionField, externalTenant)
	}
	region := regionField.String()
	parentIDField := gjson.Get(jsonPayload, ep.FieldMapping.GlobalAccountKey)
	if !parentIDField.Exists() {
		return nil, invalidFieldFormatError(ep.FieldMapping.GlobalAccountKey)
	}
	parentID := parentIDField.String()
	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Parent:         parentID,
		Subdomain:      subdomain,
		Region:         region,
		Type:           tenant.TypeToStr(tenant.Subaccount),
		Provider:       ep.ProviderName,
	}, nil
}

// Returns id of the fetched tenant, since there are multiple possible names for the ID field
func determineTenantID(jsonPayload string, mapping TenantFieldMapping) (string, error) {
	if gjson.Get(jsonPayload, mapping.IDField).Exists() {
		return gjson.Get(jsonPayload, mapping.IDField).String(), nil
	} else if gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField).Exists() {
		return gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField).String(), nil
	} else if gjson.Get(jsonPayload, mapping.SubaccountIDField).Exists() {
		return gjson.Get(jsonPayload, mapping.SubaccountIDField).String(), nil
	}
	return "", errors.Errorf("Missing or invalid format of the ID field")
}

func invalidFieldFormatError(fieldName string) error {
	return errors.Errorf("invalid format of %s field", fieldName)
}
