package tenantfetcher

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type TenantFieldMapping struct {
	NameField          string `envconfig:"default=name,APP_MAPPING_FIELD_NAME"`
	IDField            string `envconfig:"default=id,APP_MAPPING_FIELD_ID"`
	DiscriminatorField string `envconfig:"APP_MAPPING_FIELD_DISCRIMINATOR"`
	DiscriminatorValue string `envconfig:"APP_MAPPING_VALUE_DISCRIMINATOR"`
}

type converter struct {
	providerName string
	fieldMapping TenantFieldMapping
}

func NewConverter(providerName string, fieldMapping TenantFieldMapping) *converter {
	return &converter{
		providerName: providerName,
		fieldMapping: fieldMapping,
	}
}

func (c converter) EventsToTenants(eventsType EventsType, events []Event) []model.BusinessTenantMappingInput {
	var tenants []model.BusinessTenantMappingInput
	for _, event := range events {
		tenant, err := c.EventToTenant(eventsType, event)
		if err != nil {
			log.Warnf("Error: %s, while parsing event: %+v", err.Error(), event)
			continue
		}
		if tenant == nil {
			continue
		}
		tenants = append(tenants, *tenant)
	}
	return tenants
}

func (c converter) EventToTenant(eventType EventsType, event Event) (*model.BusinessTenantMappingInput, error) {
	if event == nil {
		return nil, nil
	}

	eventDataJSON, ok := event["eventData"].(string)
	if !ok {
		return nil, errors.New("invalid event data format")
	}

	var eventData map[string]interface{}
	err := json.Unmarshal([]byte(eventDataJSON), &eventData)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling event data")
	}

	tenant, err := c.eventDataToTenant(eventType, eventData)
	if err != nil {
		return nil, err
	}

	return tenant, nil
}

func (c converter) eventDataToTenant(eventType EventsType, eventData map[string]interface{}) (*model.BusinessTenantMappingInput, error) {
	if eventType == CreatedEventsType && c.fieldMapping.DiscriminatorField != "" {
		discriminator, ok := eventData[c.fieldMapping.DiscriminatorField].(string)
		if !ok {
			return nil, errors.Errorf("invalid format of %s field", c.fieldMapping.DiscriminatorField)
		}

		if discriminator != c.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, ok := eventData[c.fieldMapping.IDField].(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", c.fieldMapping.IDField)
	}

	name, ok := eventData[c.fieldMapping.NameField].(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", c.fieldMapping.NameField)
	}

	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: id,
		Provider:       c.providerName,
	}, nil
}
