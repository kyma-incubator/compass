package tenantfetcher

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type eventsPage struct {
	fieldMapping                    TenantFieldMapping
	movedRuntimeByLabelFieldMapping MovedRuntimeByLabelFieldMapping
	providerName                    string
	payload                         []byte
}

func (ep eventsPage) getEventsDetails() [][]byte {
	tenantDetails := make([][]byte, 0)
	gjson.GetBytes(ep.payload, ep.fieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		detailsType := event.Get(ep.fieldMapping.DetailsField).Type
		var details []byte
		if detailsType == gjson.String {
			details = []byte(gjson.Parse(event.Get(ep.fieldMapping.DetailsField).String()).Raw)
		} else if detailsType == gjson.JSON {
			details = []byte(event.Get(ep.fieldMapping.DetailsField).Raw)
		} else {
			log.D().Warnf("Invalid event data format: %+v", event)
			return true
		}

		tenantDetails = append(tenantDetails, details)
		return true
	})
	return tenantDetails
}

func (ep eventsPage) getMovedRuntimes() ([]model.MovedRuntimeByLabelMappingInput, error) {
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

	return mappings, nil
}

func (ep eventsPage) getTenantMappings(eventsType EventsType) ([]model.BusinessTenantMappingInput, error) {
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

	return tenants, nil
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
	if eventType == CreatedEventsType && ep.fieldMapping.DiscriminatorField != "" {
		discriminatorResult := gjson.Get(jsonPayload, ep.fieldMapping.DiscriminatorField)
		if !discriminatorResult.Exists() {
			return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.DiscriminatorField)
		}
		if discriminatorResult.String() != ep.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}
	idResult := gjson.Get(jsonPayload, ep.fieldMapping.IDField)
	if !idResult.Exists() {
		return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.IDField)
	}

	nameResult := gjson.Get(jsonPayload, ep.fieldMapping.NameField)
	if !nameResult.Exists() {
		return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.NameField)
	}

	customerIdResult := gjson.Get(jsonPayload, ep.fieldMapping.CustomerIDField)
	if !customerIdResult.Exists() {
		log.D().Warnf("Missig or invalid format of field: %s for tenant with ID: %s", ep.fieldMapping.CustomerIDField, idResult.String())
	}

	return &model.BusinessTenantMappingInput{
		Name:           nameResult.String(),
		ExternalTenant: idResult.String(),
		Parent:         customerIdResult.String(),
		Type:           tenant.TypeToStr(tenant.Account),
		Provider:       ep.providerName,
	}, nil
}
