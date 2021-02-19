package osb_test

import (
	"fmt"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name             string
		app              *schema.ApplicationExt
		expectedServices *domain.Service
		expectedErr      string
	}{
		{
			name: "Success",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundles(1, 1, 1),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(1, 1, 1),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Success with multiple entities",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundles(2, 3, 4),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(2, 3, 4),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Application with no bundles",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundles(0, 0, 0),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(0, 0, 0),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Application with one bundle without definitions",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundles(1, 0, 0),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(1, 0, 0),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Application description is nil",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundles(1, 1, 1),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "service generated from system with name app1",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(1, 1, 1),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Error is returned when bundle has invalid schema",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(func(s *schema.BundleExt) *schema.BundleExt {
						faultySchema := schema.JSONSchema(`NOT A JSON`)
						s.InstanceAuthRequestInputSchema = &faultySchema
						return s
					}),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: nil,
			expectedErr:      "while unmarshaling JSON schema: NOT A JSON",
		},
		{
			name: "Plan description is nil",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(func(ext *schema.BundleExt) *schema.BundleExt {
						ext.Description = nil
						return ext
					}),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "service generated from system with name app1",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans: generatePlansWithModification(func(s domain.ServicePlan) domain.ServicePlan {
					s.Description = fmt.Sprintf("plan generated from bundle with name %s", s.Name)
					return s
				}),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "ApiDef spec format is invalid",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(func(ext *schema.BundleExt) *schema.BundleExt {
						ext.APIDefinitions.Data[0].Spec.Format = "application/I_AM_NOT_A_JSON"
						return ext
					}),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: nil,
			expectedErr:      "unknown spec format application/I_AM_NOT_A_JSON",
		},
		{
			name: "EventDef spec format is invalid",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(func(ext *schema.BundleExt) *schema.BundleExt {
						ext.EventDefinitions.Data[0].Spec.Format = "application/I_AM_NOT_A_JSON"
						return ext
					}),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: nil,
			expectedErr:      "unknown spec format application/I_AM_NOT_A_JSON",
		},
		{
			name: "Application labels are nil",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: nil,
				Bundles: schema.BundlePageExt{
					Data: generateBundles(1, 1, 1),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "service generated from system with name app1",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generateExpectations(1, 1, 1),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  map[string]interface{}{},
				},
			},
			expectedErr: "",
		},
		{
			name: "Bundle instance auth request input schema is nil",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(func(ext *schema.BundleExt) *schema.BundleExt {
						ext.InstanceAuthRequestInputSchema = nil
						return ext
					}),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans: generatePlansWithModification(func(s domain.ServicePlan) domain.ServicePlan {
					s.Schemas = nil
					return s
				}),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
		{
			name: "Success for entities containing group and version",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					BaseEntity:   &schema.BaseEntity{ID: "id"},
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Bundles: schema.BundlePageExt{
					Data: generateBundleWithModification(addGroupAndVersionToBundle),
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: &domain.Service{
				ID:                   "id",
				Name:                 "app1",
				Description:          "description",
				Bindable:             true,
				InstancesRetrievable: false,
				BindingsRetrievable:  true,
				PlanUpdatable:        false,
				Plans:                generatePlansWithModification(addGroupAndVersionToPlan),
				Metadata: &domain.ServiceMetadata{
					DisplayName:         "app1",
					ProviderDisplayName: "provider",
					AdditionalMetadata:  schema.Labels{"key": "value"},
				},
			},
			expectedErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := osb.CatalogConverter{
				BaseURL: "http://specification.com",
			}
			service, err := c.Convert(tt.app)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.Equal(t, tt.expectedServices, service)
			}
		})
	}
}

func generateExpectations(bundlesCount, apiDefCount, eventDefCount int) []domain.ServicePlan {
	plans := make([]domain.ServicePlan, 0, bundlesCount)
	for i := 0; i < bundlesCount; i++ {
		plan := domain.ServicePlan{
			ID:          fmt.Sprintf("id%d", i),
			Name:        fmt.Sprintf("bundle%d", i),
			Description: "description",
			Bindable:    func(b bool) *bool { return &b }(true),
			Metadata:    &domain.ServicePlanMetadata{},
			Schemas: &domain.ServiceSchemas{
				Instance: domain.ServiceInstanceSchema{
					Create: domain.Schema{
						Parameters: map[string]interface{}{"param": "string"},
					},
				},
				Binding: domain.ServiceBindingSchema{},
			},
		}

		apis := make([]map[string]interface{}, 0, 0)
		for j := 0; j < apiDefCount; j++ {
			apiSpec := map[string]interface{}{
				"id":          fmt.Sprintf("id%d", j),
				"name":        fmt.Sprintf("apiDef%d", j),
				"target_url":  fmt.Sprintf("target-url-%d", j),
				"description": strToPtrStr(fmt.Sprintf("description%d", j)),
			}
			specification := make(map[string]interface{})
			specification["type"] = schema.APISpecTypeOdata
			specification["format"] = "application/json"
			specification["url"] = fmt.Sprintf("http://specification.com/specifications?app_id=id&bundle_id=id%d&definition_id=id%d", i, j)
			apiSpec["specification"] = specification

			apis = append(apis, apiSpec)
		}

		events := make([]map[string]interface{}, 0, 0)
		for j := 0; j < eventDefCount; j++ {
			eventSpec := map[string]interface{}{
				"id":          fmt.Sprintf("id%d", j),
				"name":        fmt.Sprintf("eventDef%d", j),
				"description": strToPtrStr(fmt.Sprintf("description%d", j)),
			}
			specification := make(map[string]interface{})
			specification["type"] = schema.EventSpecTypeAsyncAPI
			specification["format"] = "application/json"
			specification["url"] = fmt.Sprintf("http://specification.com/specifications?app_id=id&bundle_id=id%d&definition_id=id%d", i, j)
			eventSpec["specification"] = specification

			events = append(events, eventSpec)
		}

		plan.Metadata.AdditionalMetadata = map[string]interface{}{
			"api_specs":   apis,
			"event_specs": events,
		}
		plans = append(plans, plan)
	}
	if len(plans) == 0 {
		return nil
	}
	return plans
}

func addGroupAndVersionToBundle(ext *schema.BundleExt) *schema.BundleExt {
	ext.APIDefinitions.Data[0].Group = strToPtrStr("group")
	ext.APIDefinitions.Data[0].Version = &schema.Version{
		Value:           "v1",
		Deprecated:      func(b bool) *bool { return &b }(true),
		DeprecatedSince: strToPtrStr("01.01.2021"),
		ForRemoval:      func(b bool) *bool { return &b }(false),
	}

	ext.EventDefinitions.Data[0].Group = strToPtrStr("group")
	ext.EventDefinitions.Data[0].Version = &schema.Version{
		Value:           "v1",
		Deprecated:      func(b bool) *bool { return &b }(true),
		DeprecatedSince: strToPtrStr("01.01.2021"),
		ForRemoval:      func(b bool) *bool { return &b }(false),
	}
	return ext
}

func addGroupAndVersionToPlan(s domain.ServicePlan) domain.ServicePlan {
	apiSpecs := s.Metadata.AdditionalMetadata["api_specs"]
	apiSpecsSlice, ok := apiSpecs.([]map[string]interface{})
	if !ok {
		log.Printf("Failed to convert api specs")
		return s
	}

	eventSpecs := s.Metadata.AdditionalMetadata["event_specs"]
	eventSpecsSlice, ok := eventSpecs.([]map[string]interface{})
	if !ok {
		log.Printf("Failed to convert event specs")
		return s
	}

	versionMap := make(map[string]interface{}, 0)
	versionMap["value"] = "v1"
	versionMap["deprecated"] = func(b bool) *bool { return &b }(true)
	versionMap["deprecated_since"] = strToPtrStr("01.01.2021")
	versionMap["for_removal"] = func(b bool) *bool { return &b }(false)

	apiSpecsSlice[0]["group"] = strToPtrStr("group")
	eventSpecsSlice[0]["group"] = strToPtrStr("group")

	apiSpecsSlice[0]["version"] = versionMap
	eventSpecsSlice[0]["version"] = versionMap

	s.Metadata.AdditionalMetadata["event_specs"] = eventSpecsSlice
	s.Metadata.AdditionalMetadata["api_specs"] = apiSpecsSlice

	return s
}

func generateBundles(bundlesCount, apiDefCount, eventDefCount int) []*schema.BundleExt {
	bundles := make([]*schema.BundleExt, 0, bundlesCount)
	for i := 0; i < bundlesCount; i++ {
		instanceAuthSchema := schema.JSONSchema(`{"param":"string"}`)
		currentBundle := &schema.BundleExt{
			Bundle: schema.Bundle{
				BaseEntity:                     &schema.BaseEntity{ID: fmt.Sprintf("id%d", i)},
				Name:                           fmt.Sprintf("bundle%d", i),
				Description:                    strToPtrStr("description"),
				InstanceAuthRequestInputSchema: &instanceAuthSchema,
			},
			APIDefinitions:   generateAPIDefinitions(apiDefCount),
			EventDefinitions: generateEventDefinitions(eventDefCount),
		}
		bundles = append(bundles, currentBundle)
	}
	return bundles
}

func generateAPIDefinitions(count int) schema.APIDefinitionPageExt {
	apiDefinitions := make([]*schema.APIDefinitionExt, 0, count)
	for i := 0; i < count; i++ {
		currentAPIDefinition := &schema.APIDefinitionExt{
			APIDefinition: schema.APIDefinition{
				BaseEntity:  &schema.BaseEntity{ID: fmt.Sprintf("id%d", i)},
				Name:        fmt.Sprintf("apiDef%d", i),
				TargetURL:   fmt.Sprintf("target-url-%d", i),
				Description: strToPtrStr(fmt.Sprintf("description%d", i)),
			},
			Spec: &schema.APISpecExt{
				APISpec: schema.APISpec{
					Format: schema.SpecFormatJSON,
					Type:   schema.APISpecTypeOdata,
				},
			},
		}
		apiDefinitions = append(apiDefinitions, currentAPIDefinition)
	}
	return schema.APIDefinitionPageExt{
		Data: apiDefinitions,
	}
}

func generateEventDefinitions(count int) schema.EventAPIDefinitionPageExt {
	eventDefinitions := make([]*schema.EventAPIDefinitionExt, 0, count)
	for i := 0; i < count; i++ {
		currentEventDefinition := &schema.EventAPIDefinitionExt{
			EventDefinition: schema.EventDefinition{
				BaseEntity:  &schema.BaseEntity{ID: fmt.Sprintf("id%d", i)},
				Name:        fmt.Sprintf("eventDef%d", i),
				Description: strToPtrStr(fmt.Sprintf("description%d", i)),
			},

			Spec: &schema.EventAPISpecExt{
				EventSpec: schema.EventSpec{
					Format: schema.SpecFormatJSON,
					Type:   schema.EventSpecTypeAsyncAPI,
				},
			},
		}
		eventDefinitions = append(eventDefinitions, currentEventDefinition)
	}

	return schema.EventAPIDefinitionPageExt{
		Data: eventDefinitions,
	}
}

func generatePlansWithModification(f func(plan domain.ServicePlan) domain.ServicePlan) []domain.ServicePlan {
	expectations := generateExpectations(1, 1, 1)
	expectations[0] = f(expectations[0])
	return expectations
}

func generateBundleWithModification(f func(*schema.BundleExt) *schema.BundleExt) []*schema.BundleExt {
	pkg := generateBundles(1, 1, 1)
	pkg[0] = f(pkg[0])
	return pkg
}

func strToPtrStr(s string) *string {
	return &s
}

func boolToPtrStr(b bool) *bool {
	return &b
}
