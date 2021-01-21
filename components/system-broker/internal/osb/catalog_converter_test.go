package osb

import (
	"fmt"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackages(1, 1, 1),
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackages(2, 3, 4),
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
			name: "Application with no packages",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackages(0, 0, 0),
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
			name: "Application with one package without definitions",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackages(1, 0, 0),
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackages(1, 1, 1),
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
			name: "Error is returned when package has invalid schema",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(func(s *schema.PackageExt) *schema.PackageExt {
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(func(ext *schema.PackageExt) *schema.PackageExt {
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
					s.Description = fmt.Sprintf("plan generated from package with name %s", s.Name)
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(func(ext *schema.PackageExt) *schema.PackageExt {
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(func(ext *schema.PackageExt) *schema.PackageExt {
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: nil,
				Packages: schema.PackagePageExt{
					Data: generatePackages(1, 1, 1),
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
			name: "Package instance auth request input schema is nil",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(func(ext *schema.PackageExt) *schema.PackageExt {
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
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  strToPtrStr("description"),
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: generatePackageWithModification(addGroupAndVersionToPackage),
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
			c := Converter{
				baseURL:      "http://specification.com",
				MapConverter: MapConverter{},
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

func generateExpectations(packagesCount, apiDefCount, eventDefCount int) []domain.ServicePlan {
	plans := make([]domain.ServicePlan, 0, packagesCount)
	for i := 0; i < packagesCount; i++ {
		plan := domain.ServicePlan{
			ID:          fmt.Sprintf("id%d", i),
			Name:        fmt.Sprintf("package%d", i),
			Description: "description",
			Bindable:    boolToPtr(true),
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
			specification["url"] = fmt.Sprintf("http://specification.com/specifications?app_id=id&package_id=id%d&definition_id=id%d", i, j)
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
			specification["url"] = fmt.Sprintf("http://specification.com/specifications?app_id=id&package_id=id%d&definition_id=id%d", i, j)
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

func generatePackages(packagesCount, apiDefCount, eventDefCount int) []*schema.PackageExt {
	packages := make([]*schema.PackageExt, 0, packagesCount)
	for i := 0; i < packagesCount; i++ {
		instanceAuthSchema := schema.JSONSchema(`{"param":"string"}`)
		currentPackage := &schema.PackageExt{
			Package: schema.Package{
				ID:                             fmt.Sprintf("id%d", i),
				Name:                           fmt.Sprintf("package%d", i),
				Description:                    strToPtrStr("description"),
				InstanceAuthRequestInputSchema: &instanceAuthSchema,
			},
			APIDefinitions:   generateAPIDefinitions(apiDefCount),
			EventDefinitions: generateEventDefinitions(eventDefCount),
		}
		packages = append(packages, currentPackage)
	}
	return packages
}

func generateAPIDefinitions(count int) schema.APIDefinitionPageExt {
	apiDefinitions := make([]*schema.APIDefinitionExt, 0, count)
	for i := 0; i < count; i++ {
		currentAPIDefinition := &schema.APIDefinitionExt{
			APIDefinition: schema.APIDefinition{
				ID:          fmt.Sprintf("id%d", i),
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
				ID:          fmt.Sprintf("id%d", i),
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

func generatePackageWithModification(f func(*schema.PackageExt) *schema.PackageExt) []*schema.PackageExt {
	pkg := generatePackages(1, 1, 1)
	pkg[0] = f(pkg[0])
	return pkg
}

func strToPtrStr(s string) *string {
	return &s
}

func boolToPtrStr(b bool) *bool {
	return &b
}
