package tenantfetcher

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

type eventsPage struct {
	fieldMapping                 TenantFieldMapping
	movedSubaccountsFieldMapping MovedSubaccountsFieldMapping
	providerName                 string
	payload                      []byte
}

func (ep eventsPage) getEventsDetails() [][]byte {
	tenantDetails := make([][]byte, 0)
	gjson.GetBytes(ep.payload, ep.fieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		entityType := event.Get(ep.fieldMapping.EntityTypeField)
		globalAccountGUID := event.Get(ep.fieldMapping.GlobalAccountGUIDField)
		details := event.Get(ep.fieldMapping.DetailsField).Map()
		details[ep.fieldMapping.EntityTypeField] = entityType
		details[ep.fieldMapping.GlobalAccountKey] = globalAccountGUID
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
			return false
		}
		tenantDetails = append(tenantDetails, currentTenantDetails)
		return true
	})
	return tenantDetails
}

func (ep eventsPage) getMovedSubaccounts() []model.MovedSubaccountMappingInput {
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

func (ep eventsPage) getTenantMappings(eventsType EventsType) []model.BusinessTenantMappingInput {
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

func (ep eventsPage) eventDataToMovedSubaccount(eventData []byte) (*model.MovedSubaccountMappingInput, error) {
	jsonPayload := string(eventData)
	if !gjson.Valid(jsonPayload) {
		return nil, errors.Errorf("invalid json payload")
	}

	id, ok := gjson.GetBytes(eventData, ep.movedSubaccountsFieldMapping.LabelValue).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountsFieldMapping.LabelValue)
	}

	source, ok := gjson.GetBytes(eventData, ep.movedSubaccountsFieldMapping.SourceTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountsFieldMapping.SourceTenant)
	}

	target, ok := gjson.GetBytes(eventData, ep.movedSubaccountsFieldMapping.TargetTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountsFieldMapping.TargetTenant)
	}

	nameResult := gjson.Get(jsonPayload, ep.fieldMapping.NameField)
	if !nameResult.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.NameField)
	}

	subdomain := gjson.Get(jsonPayload, ep.fieldMapping.SubdomainField)
	if !subdomain.Exists() {
		log.D().Warnf("Missig or invalid format of field: %s for tenant with ID: %s", ep.fieldMapping.SubdomainField, id)
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

func (ep eventsPage) eventDataToTenant(eventType EventsType, eventData []byte) (*model.BusinessTenantMappingInput, error) {
	jsonPayload := string(eventData)
	if !gjson.Valid(jsonPayload) {
		return nil, errors.Errorf("invalid json payload")
	}
	if eventType == CreatedAccountType && ep.fieldMapping.DiscriminatorField != "" {
		discriminatorResult := gjson.Get(jsonPayload, ep.fieldMapping.DiscriminatorField)
		if !discriminatorResult.Exists() {
			return nil, invalidFieldFormatError(ep.fieldMapping.DiscriminatorField)
		}
		if discriminatorResult.String() != ep.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, err := determineTenantID(jsonPayload, ep.fieldMapping)
	if err != nil {
		return nil, err
	}

	nameResult := gjson.Get(jsonPayload, ep.fieldMapping.NameField)
	if !nameResult.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.NameField)
	}

	subdomain := gjson.Get(jsonPayload, ep.fieldMapping.SubdomainField)
	if !subdomain.Exists() {
		log.D().Warnf("Missing or invalid format of field: %s for tenant with ID: %s", ep.fieldMapping.SubdomainField, id)
	}

	entityType := gjson.Get(jsonPayload, ep.fieldMapping.EntityTypeField)
	if !entityType.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.EntityTypeField)
	}

	globalAccountRegex := regexp.MustCompile(GlobalAccountRegex)
	if globalAccountRegex.MatchString(entityType.String()) {
		return constructGlobalAccountTenant(jsonPayload, nameResult.String(), subdomain.String(), id, ep), nil
	} else {
		return constructSubaccountTenant(jsonPayload, nameResult.String(), subdomain.String(), id, ep)
	}
}

func constructGlobalAccountTenant(jsonPayload, name, subdomain, externalTenant string, ep eventsPage) *model.BusinessTenantMappingInput {
	parentID := ""
	customerIDResult := gjson.Get(jsonPayload, ep.fieldMapping.CustomerIDField)
	if !customerIDResult.Exists() {
		log.D().Warnf("Missig or invalid format of field: %s for tenant with id: %s", ep.fieldMapping.CustomerIDField, externalTenant)
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
		Provider:       ep.providerName,
	}
}

func constructSubaccountTenant(jsonPayload, name, subdomain, externalTenant string, ep eventsPage) (*model.BusinessTenantMappingInput, error) {
	regionField := gjson.Get(jsonPayload, ep.fieldMapping.RegionField)
	if !regionField.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.RegionField)
	}
	region := regionField.String()
	parentIDField := gjson.Get(jsonPayload, ep.fieldMapping.GlobalAccountKey)
	if !parentIDField.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.GlobalAccountKey)
	}
	parentID := parentIDField.String()
	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Parent:         parentID,
		Subdomain:      subdomain,
		Region:         region,
		Type:           tenant.TypeToStr(tenant.Subaccount),
		Provider:       ep.providerName,
	}, nil
}

// Returns id of the fetched tenant, since there are multiple possible names for the ID field
func determineTenantID(jsonPayload string, mapping TenantFieldMapping) (string, error) {
	if gjson.Get(jsonPayload, mapping.IDField).Exists() {
		return gjson.Get(jsonPayload, mapping.IDField).String(), nil
	} else if gjson.Get(jsonPayload, mapping.SubaccountIDField).Exists() {
		return gjson.Get(jsonPayload, mapping.SubaccountIDField).String(), nil
	} else if gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField).Exists() {
		return gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField).String(), nil
	} else if gjson.Get(jsonPayload, mapping.SubaccountGUIDField).Exists() {
		return gjson.Get(jsonPayload, mapping.SubaccountGUIDField).String(), nil
	}
	return "", errors.Errorf("Missing or invalid format of the ID field")
}

func invalidFieldFormatError(fieldName string) error {
	return errors.Errorf("invalid format of %s field", fieldName)
}
