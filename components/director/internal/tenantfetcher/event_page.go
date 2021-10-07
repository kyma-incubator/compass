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

// GlobalAccountEntityType is the value for entityType property when fetching subaccounts
const GlobalAccountEntityType = "GlobalAccount"

type eventsPage struct {
	fieldMapping                    TenantFieldMapping
	movedRuntimeByLabelFieldMapping MovedRuntimeByLabelFieldMapping
	providerName                    string
	payload                         []byte
}

func (ep eventsPage) getEventsDetails() [][]byte {
	tenantDetails := make([][]byte, 0)
	gjson.GetBytes(ep.payload, ep.fieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		entityType := event.Get(ep.fieldMapping.EntityTypeField)
		details := event.Get(ep.fieldMapping.DetailsField).Map()
		details[ep.fieldMapping.EntityTypeField] = entityType
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

func (ep eventsPage) getMovedRuntimes() []model.MovedRuntimeByLabelMappingInput {
	eds := ep.getEventsDetails()
	mappings := make([]model.MovedRuntimeByLabelMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToMovedRuntime(detail)
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

func (ep eventsPage) eventDataToMovedRuntime(eventData []byte) (*model.MovedRuntimeByLabelMappingInput, error) {
	id, ok := gjson.GetBytes(eventData, ep.movedRuntimeByLabelFieldMapping.LabelValue).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedRuntimeByLabelFieldMapping.LabelValue)
	}

	source, ok := gjson.GetBytes(eventData, ep.movedRuntimeByLabelFieldMapping.SourceTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedRuntimeByLabelFieldMapping.SourceTenant)
	}

	target, ok := gjson.GetBytes(eventData, ep.movedRuntimeByLabelFieldMapping.TargetTenant).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedRuntimeByLabelFieldMapping.TargetTenant)
	}

	return &model.MovedRuntimeByLabelMappingInput{
		LabelValue:   id,
		SourceTenant: source,
		TargetTenant: target,
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
	idResult := gjson.Get(jsonPayload, ep.fieldMapping.IDField)
	if !idResult.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.IDField)
	}

	nameResult := gjson.Get(jsonPayload, ep.fieldMapping.NameField)
	if !nameResult.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.NameField)
	}

	subdomain := gjson.Get(jsonPayload, ep.fieldMapping.SubdomainField)
	if !subdomain.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.SubdomainField)
	}

	entityType := gjson.Get(jsonPayload, ep.fieldMapping.EntityTypeField)
	if !entityType.Exists() {
		return nil, invalidFieldFormatError(ep.fieldMapping.EntityTypeField)
	}

	region := ""
	parentID := ""
	tenantType := tenant.TypeToStr(tenant.Account)
	globalAccountRegex := regexp.MustCompile("^GLOBALACCOUNT_.*|GlobalAccount")
	if globalAccountRegex.MatchString(entityType.String()) {
		customerIDResult := gjson.Get(jsonPayload, ep.fieldMapping.CustomerIDField)
		if !customerIDResult.Exists() {
			log.D().Warnf("Missig or invalid format of field: %s for tenant with ID: %s", ep.fieldMapping.CustomerIDField, idResult.String())
		} else {
			parentID = customerIDResult.String()
		}
	} else {
		regionField := gjson.Get(jsonPayload, ep.fieldMapping.RegionField)
		if !regionField.Exists() {
			return nil, invalidFieldFormatError(ep.fieldMapping.RegionField)
		}
		region = regionField.String()
		parentIDField := gjson.Get(jsonPayload, ep.fieldMapping.ParentIDField)
		if !parentIDField.Exists() {
			return nil, invalidFieldFormatError(ep.fieldMapping.ParentIDField)
		}
		parentID = parentIDField.String()
		tenantType = tenant.TypeToStr(tenant.Subaccount)
	}

	return &model.BusinessTenantMappingInput{
		Name:           nameResult.String(),
		ExternalTenant: idResult.String(),
		Parent:         parentID,
		Subdomain:      subdomain.String(),
		Region:         region,
		Type:           tenantType,
		Provider:       ep.providerName,
	}, nil
}

func invalidFieldFormatError(fieldName string) error {
	return errors.Errorf("invalid format of %s field", fieldName)
}
