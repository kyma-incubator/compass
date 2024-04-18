package resync

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
func (ep EventsPage) GetMovedSubaccounts(ctx context.Context) []model.MovedSubaccountMappingInput {
	eds := ep.getEventsDetails(ctx)
	mappings := make([]model.MovedSubaccountMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToMovedSubaccount(ctx, detail)
		if err != nil {
			log.C(ctx).Warnf("Error: %s. Could not convert tenant: %s", err.Error(), string(detail))
			continue
		}

		mappings = append(mappings, *mapping)
	}

	return mappings
}

// GetTenantMappings parses the data from the page payload to BusinessTenantMappingInput
func (ep EventsPage) GetTenantMappings(ctx context.Context, eventsType EventsType) []model.BusinessTenantMappingInput {
	eds := ep.getEventsDetails(ctx)
	tenants := make([]model.BusinessTenantMappingInput, 0)
	for _, detail := range eds {
		mappings, err := ep.eventDataToTenant(ctx, eventsType, detail)
		if err != nil {
			log.C(ctx).Warnf("Error: %s. Could not convert tenant: %s", err.Error(), string(detail))
			continue
		}

		for _, mapping := range mappings {
			tenants = append(tenants, *mapping)
		}
	}

	return tenants
}

func (ep EventsPage) getEventsDetails(ctx context.Context) [][]byte {
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
			case gjson.JSON:
				allDetails[key] = result.Value()
			case gjson.True:
				allDetails[key] = true
			case gjson.False:
				allDetails[key] = false
			case gjson.Null:
				allDetails[key] = nil
			default:
				log.C(ctx).Debugf("Unknown property type %s", result.Type)
			}
		}
		currentTenantDetails, err := json.Marshal(allDetails)
		if err != nil {
			log.C(ctx).Errorf("failed to marshal tenant details: %v", err)
			return false
		}
		tenantDetails = append(tenantDetails, currentTenantDetails)
		return true
	})
	return tenantDetails
}

func (ep EventsPage) eventDataToMovedSubaccount(ctx context.Context, eventData []byte) (*model.MovedSubaccountMappingInput, error) {
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
		log.C(ctx).Warnf("Missig or invalid format of field: %s for tenant with ID: %s", ep.FieldMapping.SubdomainField, id)
	}

	licenseType := gjson.Get(jsonPayload, ep.FieldMapping.LicenseTypeField)
	var licenseTypeValue *string
	if licenseType.Exists() && licenseType.Type == gjson.String {
		licenseTypeValue = &licenseType.Str
	} else {
		log.C(ctx).Warnf("Missing or invalid format of licenseType field: %s for tenant with ID: %s", ep.FieldMapping.LicenseTypeField, id)
	}

	subaccountInput, err := constructSubaccountTenant(ctx, jsonPayload, nameResult.String(), subdomain.String(), id, licenseTypeValue, ep)
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

func (ep EventsPage) eventDataToTenant(ctx context.Context, eventType EventsType, eventData []byte) ([]*model.BusinessTenantMappingInput, error) {
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
		log.C(ctx).Warnf("Missing or invalid format of name field: %s for tenant with ID: %s", ep.FieldMapping.NameField, id)
	}

	subdomain := gjson.Get(jsonPayload, ep.FieldMapping.SubdomainField)
	if !subdomain.Exists() {
		log.C(ctx).Warnf("Missing or invalid format of subdomain field: %s for tenant with ID: %s", ep.FieldMapping.SubdomainField, id)
	}

	entityType := gjson.Get(jsonPayload, ep.FieldMapping.EntityTypeField)
	if !entityType.Exists() {
		return nil, invalidFieldFormatError(ep.FieldMapping.EntityTypeField)
	}

	licenseType := gjson.Get(jsonPayload, ep.FieldMapping.LicenseTypeField)
	var licenseTypeValue *string
	if licenseType.Exists() && licenseType.Type == gjson.String {
		licenseTypeValue = &licenseType.Str
	} else {
		log.C(ctx).Warnf("Missing or invalid format of licenseType field: %s for tenant with ID: %s", ep.FieldMapping.LicenseTypeField, id)
	}

	globalAccountRegex := regexp.MustCompile(GlobalAccountRegex)
	if globalAccountRegex.MatchString(entityType.String()) {
		globalAccount := constructGlobalAccountTenant(ctx, jsonPayload, nameResult.String(), subdomain.String(), id, licenseTypeValue, ep)

		if !gjson.Get(jsonPayload, ep.FieldMapping.CostObjectIDField).Exists() {
			return []*model.BusinessTenantMappingInput{globalAccount}, nil
		}

		costObjectIDResult := gjson.Get(jsonPayload, ep.FieldMapping.CostObjectIDField).String()
		costObject := constructCostObjectTenant(costObjectIDResult, licenseTypeValue, ep)
		return []*model.BusinessTenantMappingInput{globalAccount, costObject}, nil
	} else {
		subaccount, err := constructSubaccountTenant(ctx, jsonPayload, nameResult.String(), subdomain.String(), id, licenseTypeValue, ep)
		if err != nil {
			return nil, err
		}

		costObjectIDField := gjson.Get(jsonPayload, ep.FieldMapping.SubaccountCostObjectIDField)
		if !costObjectIDField.Exists() || costObjectIDField.String() == "" {
			return []*model.BusinessTenantMappingInput{subaccount}, nil
		}

		subaccount.CostObjectID = str.Ptr(costObjectIDField.String())

		costObject := constructCostObjectTenant(costObjectIDField.String(), licenseTypeValue, ep)
		costObject.CostObjectType = str.Ptr(gjson.Get(jsonPayload, ep.FieldMapping.SubaccountCostObjectTypeField).String())

		return []*model.BusinessTenantMappingInput{subaccount, costObject}, err
	}
}

func constructGlobalAccountTenant(ctx context.Context, jsonPayload, name, subdomain, externalTenant string, licenseType *string, ep EventsPage) *model.BusinessTenantMappingInput {
	parents := make([]string, 0)
	customerIDResult := gjson.Get(jsonPayload, ep.FieldMapping.CustomerIDField)
	costObjectIDResult := gjson.Get(jsonPayload, ep.FieldMapping.CostObjectIDField)

	if !customerIDResult.Exists() {
		log.C(ctx).Warnf("Missig or invalid format of field: %s for tenant with id: %s", ep.FieldMapping.CustomerIDField, externalTenant)
	} else {
		parents = append(parents, customerIDResult.String())
	}

	if !costObjectIDResult.Exists() {
		log.C(ctx).Warnf("Missig or invalid format of field: %s for tenant with id: %s", ep.FieldMapping.CostObjectIDField, externalTenant)
	} else {
		parents = append(parents, costObjectIDResult.String())
	}

	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Parents:        parents,
		Subdomain:      subdomain,
		Region:         "",
		Type:           tenant.TypeToStr(tenant.Account),
		Provider:       ep.ProviderName,
		LicenseType:    licenseType,
	}
}

func constructCostObjectTenant(costObjectID string, licenseType *string, ep EventsPage) *model.BusinessTenantMappingInput {
	return &model.BusinessTenantMappingInput{
		Name:           costObjectID,
		ExternalTenant: costObjectID,
		Parents:        []string{},
		Subdomain:      "",
		Region:         "",
		Type:           tenant.TypeToStr(tenant.CostObject),
		Provider:       ep.ProviderName,
		LicenseType:    licenseType,
	}
}

func constructSubaccountTenant(ctx context.Context, jsonPayload, name, subdomain, externalTenant string, licenseType *string, ep EventsPage) (*model.BusinessTenantMappingInput, error) {
	regionField := gjson.Get(jsonPayload, ep.FieldMapping.RegionField)
	if !regionField.Exists() {
		log.C(ctx).Debugf("Missing or invalid format of region field: %s for tenant with ID: %s", ep.FieldMapping.RegionField, externalTenant)
	}
	region := regionField.String()

	parentIDField := gjson.Get(jsonPayload, ep.FieldMapping.GlobalAccountKey)
	if !parentIDField.Exists() {
		return nil, invalidFieldFormatError(ep.FieldMapping.GlobalAccountKey)
	}
	parentID := parentIDField.String()

	var customerIDValue *string
	customerIDField := gjson.Get(jsonPayload, ep.FieldMapping.LabelsField).Get(ep.FieldMapping.CustomerIDField)
	if customerIDFieldArr := customerIDField.Array(); customerIDField.IsArray() && len(customerIDFieldArr) > 0 {
		customerIDValue = str.Ptr(tenant.TrimCustomerIDLeadingZeros(customerIDFieldArr[0].String()))
	}

	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: externalTenant,
		Parents:        []string{parentID},
		Subdomain:      subdomain,
		Region:         region,
		Type:           tenant.TypeToStr(tenant.Subaccount),
		Provider:       ep.ProviderName,
		LicenseType:    licenseType,
		CustomerID:     customerIDValue,
	}, nil
}

// Returns id of the fetched tenant, since there are multiple possible names for the ID field
func determineTenantID(jsonPayload string, mapping TenantFieldMapping) (string, error) {
	id := gjson.Get(jsonPayload, mapping.IDField)
	globalAccountGUID := gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField)
	subaccountID := gjson.Get(jsonPayload, mapping.SubaccountIDField)

	if id.Exists() && id.String() != "" {
		return gjson.Get(jsonPayload, mapping.IDField).String(), nil
	} else if globalAccountGUID.Exists() && globalAccountGUID.String() != "" {
		return gjson.Get(jsonPayload, mapping.GlobalAccountGUIDField).String(), nil
	} else if subaccountID.Exists() && subaccountID.String() != "" {
		return gjson.Get(jsonPayload, mapping.SubaccountIDField).String(), nil
	}
	return "", errors.Errorf("Missing or invalid format of the ID field")
}

func invalidFieldFormatError(fieldName string) error {
	return errors.Errorf("invalid format of %s field", fieldName)
}
