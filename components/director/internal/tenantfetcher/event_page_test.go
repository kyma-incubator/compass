package tenantfetcher

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_getMovedSubaccounts(t *testing.T) {
	idField := "id"
	entityTypeField := "type"
	nameField := "name"
	subdomainField := "subdomain"
	regionField := "region"
	parentID := "parent"
	globalAccountKey := "gaID"
	labelFieldMappingValue := "moved-label"
	sourceTenantField := "source-tenant"
	targetTenantField := "target-tenant"
	expectedRuntime := model.MovedSubaccountMappingInput{
		SubaccountID: "label-value",
		SourceTenant: "123",
		TargetTenant: "456",
		TenantMappingInput: model.BusinessTenantMappingInput{
			Name:           "subaccount-name",
			ExternalTenant: "label-value",
			Parent:         "parent",
			Subdomain:      "subdomain",
			Region:         "region",
			Type:           string(tenant.Subaccount),
			Provider:       "",
		},
	}
	fieldMapping := TenantFieldMapping{
		IDField:          idField,
		NameField:        nameField,
		SubdomainField:   subdomainField,
		EntityTypeField:  entityTypeField,
		RegionField:      regionField,
		GlobalAccountKey: globalAccountKey,
		EventsField:      "events",
		DetailsField:     "details",
	}
	tests := []struct {
		name               string
		detailsPairs       [][]Pair
		errorFunc          func(*testing.T, error)
		assertRuntimesFunc func(*testing.T, []model.MovedSubaccountMappingInput)
	}{
		{
			name: "successfully gets MovedSubaccountsMappingInputs for correct eventPage format",
			detailsPairs: [][]Pair{
				{
					{labelFieldMappingValue, "label-value"},
					{sourceTenantField, "123"},
					{targetTenantField, "456"},
					{nameField, "subaccount-name"},
					{subdomainField, "subdomain"},
					{regionField, "region"},
				},
			},
			assertRuntimesFunc: func(t *testing.T, runtimes []model.MovedSubaccountMappingInput) {
				assert.Equal(t, 1, len(runtimes))
				assert.Equal(t, expectedRuntime, runtimes[0])
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "empty mappings for get MovedSubaccountMappingInput when id field is invalid",
			detailsPairs: [][]Pair{
				{
					{"wrong", "label-value"},
					{sourceTenantField, "123"},
					{targetTenantField, "456"},
				},
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertRuntimesFunc: func(t *testing.T, inputs []model.MovedSubaccountMappingInput) {
				assert.Len(t, inputs, 0)
			},
		},
		{
			name: "empty mappings for get MovedSubaccountMappingInput when sourceTenant field is invalid",
			detailsPairs: [][]Pair{
				{
					{labelFieldMappingValue, "label-value"},
					{"wrong", "123"},
					{targetTenantField, "456"},
				},
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertRuntimesFunc: func(t *testing.T, inputs []model.MovedSubaccountMappingInput) {
				assert.Len(t, inputs, 0)
			},
		},
		{
			name: "empty mappings for get MovedSubaccountMappingInput when targetTenant field is invalid",
			detailsPairs: [][]Pair{
				{
					{labelFieldMappingValue, "label-value"},
					{sourceTenantField, "123"},
					{"wrong", "456"},
				},
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertRuntimesFunc: func(t *testing.T, inputs []model.MovedSubaccountMappingInput) {
				assert.Len(t, inputs, 0)
			},
		},
		{
			name: "events are skipped if some of the fields are invalid",
			detailsPairs: [][]Pair{
				{
					{labelFieldMappingValue, "label-value"},
					{sourceTenantField, "123"},
					{nameField, "name"},
					{regionField, "region"},
					{"wrong", "456"},
				},
				{
					{labelFieldMappingValue, "label-value"},
					{sourceTenantField, "123"},
					{targetTenantField, "456"},
					{nameField, "name"},
					{regionField, "region"},
				},
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertRuntimesFunc: func(t *testing.T, inputs []model.MovedSubaccountMappingInput) {
				assert.Len(t, inputs, 1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make([][]byte, 0, len(test.detailsPairs))
			for i, detailPair := range test.detailsPairs {
				events = append(events, fixEventWithDetails(fmt.Sprintf("id%d", i), fmt.Sprintf("foo%d", i), "GlobalAccount", parentID, constructJSONObject(detailPair...), fieldMapping))
			}
			page := eventsPage{
				fieldMapping: fieldMapping,
				movedSubaccountsFieldMapping: MovedSubaccountsFieldMapping{
					LabelValue:   labelFieldMappingValue,
					SourceTenant: sourceTenantField,
					TargetTenant: targetTenantField,
				},
				payload: []byte(fixTenantEventsResponse(eventsToJSONArray(events...), len(test.detailsPairs), 1)),
			}

			runtimes := page.getMovedSubaccounts()
			test.assertRuntimesFunc(t, runtimes)
		})
	}
}

func Test_getTenantMappings(t *testing.T) {
	idField := "id"
	globalAccountGUIDField := "globalAccountGUID"
	globalAccountKey := "gaID"
	id := "1"
	nameField := "name"
	name := "test-name"
	discriminatorField := "discriminator"
	subdomainField := "subdomain"
	subdomain := "test-subdomain"
	providerName := "test-provider"
	entityTypeField := "type"
	entityType := "account"

	expectedTenantMapping := model.BusinessTenantMappingInput{
		ExternalTenant: id,
		Name:           name,
		Subdomain:      subdomain,
		Type:           entityType,
		Provider:       providerName,
	}

	tests := []struct {
		name                    string
		detailsPairs            [][]Pair
		fieldMapping            TenantFieldMapping
		errorFunc               func(*testing.T, error)
		assertTenantMappingFunc func(*testing.T, []model.BusinessTenantMappingInput)
	}{
		{
			name: "successfully gets businessTenantMappingInputs for correct eventPage format",
			fieldMapping: TenantFieldMapping{
				NameField:              nameField,
				IDField:                idField,
				SubdomainField:         subdomainField,
				EventsField:            "events",
				DetailsField:           "details",
				EntityTypeField:        entityTypeField,
				GlobalAccountGUIDField: globalAccountGUIDField,
				GlobalAccountKey:       globalAccountKey,
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertTenantMappingFunc: func(t *testing.T, tenantMappings []model.BusinessTenantMappingInput) {
				assert.Equal(t, 1, len(tenantMappings))
				assert.Equal(t, expectedTenantMapping, tenantMappings[0])
			},
			detailsPairs: [][]Pair{
				{
					{idField, id},
					{nameField, name},
					{subdomainField, subdomain},
				},
			},
		},
		{
			name: "successfully gets businessTenantMappingInputs for correct eventPage format with discriminator field",
			fieldMapping: TenantFieldMapping{
				NameField:              nameField,
				IDField:                idField,
				EventsField:            "events",
				DetailsField:           "details",
				SubdomainField:         "subdomain",
				DiscriminatorField:     discriminatorField,
				DiscriminatorValue:     "discriminator-value",
				EntityTypeField:        entityTypeField,
				GlobalAccountGUIDField: globalAccountGUIDField,
				GlobalAccountKey:       globalAccountKey,
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertTenantMappingFunc: func(t *testing.T, tenantMappings []model.BusinessTenantMappingInput) {
				assert.Equal(t, 1, len(tenantMappings))
				assert.Equal(t, expectedTenantMapping, tenantMappings[0])
			},
			detailsPairs: [][]Pair{
				{
					{idField, id},
					{nameField, name},
					{discriminatorField, "discriminator-value"},
					{subdomainField, subdomain},
					{entityTypeField, tenant.TypeToStr(tenant.Account)},
				},
			},
		},
		{
			name: "empty mappings for get businessTenantMappingInputs when id field is wrong",
			fieldMapping: TenantFieldMapping{
				NameField:          nameField,
				IDField:            idField,
				EventsField:        "events",
				DetailsField:       "details",
				DiscriminatorField: discriminatorField,
				DiscriminatorValue: "discriminator-value",
				EntityTypeField:    entityTypeField,
				GlobalAccountKey:   globalAccountKey,
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertTenantMappingFunc: func(t *testing.T, tenantMappings []model.BusinessTenantMappingInput) {
				assert.Len(t, tenantMappings, 0)
			},
			detailsPairs: [][]Pair{
				{
					{"wrong", id},
					{nameField, name},
					{discriminatorField, "discriminator-value"},
				},
			},
		},
		{
			name: "empty mappings for get businessTenantMappingInputs when name field is wrong",
			fieldMapping: TenantFieldMapping{
				NameField:          nameField,
				IDField:            idField,
				EventsField:        "events",
				DetailsField:       "details",
				DiscriminatorField: discriminatorField,
				DiscriminatorValue: "discriminator-value",
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertTenantMappingFunc: func(t *testing.T, tenantMappings []model.BusinessTenantMappingInput) {
				assert.Len(t, tenantMappings, 0)
			},
			detailsPairs: [][]Pair{
				{
					{idField, id},
					{"wrong", name},
					{discriminatorField, "discriminator-value"},
				},
			},
		},
		{
			name: "empty mappings for get businessTenantMappingInputs when discriminator field is wrong",
			fieldMapping: TenantFieldMapping{
				NameField:          nameField,
				IDField:            idField,
				EventsField:        "events",
				DetailsField:       "details",
				DiscriminatorField: discriminatorField,
				DiscriminatorValue: "discriminator-value",
			},
			errorFunc: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertTenantMappingFunc: func(t *testing.T, tenantMappings []model.BusinessTenantMappingInput) {
				assert.Len(t, tenantMappings, 0)
			},
			detailsPairs: [][]Pair{
				{
					{idField, id},
					{nameField, name},
					{"wrong", "discriminator-value"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make([][]byte, 0, len(test.detailsPairs))
			for i, detailPair := range test.detailsPairs {
				events = append(events, fixEventWithDetails(fmt.Sprintf("id%d", i), fmt.Sprintf("foo%d", i), "GlobalAccount", fmt.Sprintf("gaID%d", i), constructJSONObject(detailPair...), test.fieldMapping))
			}
			page := eventsPage{
				fieldMapping: test.fieldMapping,
				providerName: providerName,
				payload:      []byte(fixTenantEventsResponse(eventsToJSONArray(events...), len(test.detailsPairs), 1)),
			}
			tenantMappings := page.getTenantMappings(CreatedAccountType)
			test.assertTenantMappingFunc(t, tenantMappings)
		})
	}
}

type Pair struct {
	Key   string
	Value string
}

func constructJSONObject(pairs ...Pair) string {
	var (
		templateName       = "jsonObject"
		jsonObjectTemplate = `{
		{{ $n := (len .) }}
		{{range $i, $e := .}}
  		"{{$e.Key}}": "{{$e.Value}}"{{if ne (plus1 $i) $n }},{{end}}
		{{end}}
	}`
		funcMap = template.FuncMap{
			"plus1": func(i int) int {
				return i + 1
			},
		}
		t      = template.Must(template.New(templateName).Funcs(funcMap).Parse(jsonObjectTemplate))
		buffer = bytes.NewBufferString("")
	)

	template.Must(t, t.ExecuteTemplate(buffer, templateName, pairs))
	return buffer.String()
}

func fixEventWithDetails(id, name, entityType, globalAccountGUID, details string, fieldMapping TenantFieldMapping) []byte {
	return []byte(fmt.Sprintf(`{"%s":"%s", "%s":"%s", "%s":"%s","%s":"%s","%s":%s}`, fieldMapping.IDField, id, fieldMapping.NameField, name, fieldMapping.EntityTypeField, entityType, fieldMapping.GlobalAccountGUIDField, globalAccountGUID, fieldMapping.DetailsField, details))
}

func fixTenantEventsResponse(events []byte, total, pages int) TenantEventsResponse {
	return TenantEventsResponse(fmt.Sprintf(`{
		"events":       %s,
		"total": %d,
		"pages":   %d,
	}`, string(events), total, pages))
}

func eventsToJSONArray(events ...[]byte) []byte {
	return []byte(fmt.Sprintf(`[%s]`, bytes.Join(events, []byte(","))))
}
