package osb

import (
	"fmt"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestConverter_Convert(t *testing.T) {

	tests := []struct {
		name             string
		baseURL          string
		app              *schema.ApplicationExt
		expectedServices *domain.Service
		expectedErr      string
	}{
		{
			name:    "Success",
			baseURL: "http://specification.com",
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
			name:    "when description is not provided",
			baseURL: "http://specification.com",
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
			name:    "when error is returned",
			baseURL: "http://specification.com",
			app: &schema.ApplicationExt{
				Application: schema.Application{
					ID:           "id",
					Name:         "app1",
					ProviderName: strToPtrStr("provider"),
					Description:  nil,
				},
				Labels: schema.Labels{"key": "value"},
				Packages: schema.PackagePageExt{
					Data: []*schema.PackageExt{
						generateFaultyPackage(),
					},
				},
				EventingConfiguration: schema.ApplicationEventingConfiguration{},
			},
			expectedServices: nil,
			expectedErr:      "while unmarshaling JSON schema: NOT A JSON",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Converter{
				baseURL: tt.baseURL,
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

func TestConverter_toPlanMetadata(t *testing.T) {
	type fields struct {
		baseURL string
	}
	type args struct {
		appID string
		pkg   *schema.PackageExt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.ServicePlanMetadata
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{
				baseURL: tt.fields.baseURL,
			}
			got, err := c.toPlanMetadata(tt.args.appID, tt.args.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("toPlanMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toPlanMetadata() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConverter_toPlans(t *testing.T) {
	type fields struct {
		baseURL string
	}
	type args struct {
		appID    string
		packages []*schema.PackageExt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []domain.ServicePlan
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{
				baseURL: tt.fields.baseURL,
			}
			got, err := c.toPlans(tt.args.appID, tt.args.packages)
			if (err != nil) != tt.wantErr {
				t.Errorf("toPlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toPlans() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConverter_toSchemas(t *testing.T) {
	type fields struct {
		baseURL string
	}
	type args struct {
		pkg *schema.PackageExt
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *domain.ServiceSchemas
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{
				baseURL: tt.fields.baseURL,
			}
			got, err := c.toSchemas(tt.args.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("toSchemas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toSchemas() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConverter_toServiceMetadata(t *testing.T) {
	type fields struct {
		baseURL string
	}
	type args struct {
		app *schema.ApplicationExt
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *domain.ServiceMetadata
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{
				baseURL: tt.fields.baseURL,
			}
			if got := c.toServiceMetadata(tt.args.app); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toServiceMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_boolPtr(t *testing.T) {
	type args struct {
		in bool
	}
	tests := []struct {
		name string
		args args
		want *bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := boolPtr(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("boolPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ptrStrToStr(t *testing.T) {
	type args struct {
		s *string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ptrStrToStr(tt.args.s); got != tt.want {
				t.Errorf("ptrStrToStr() = %v, want %v", got, tt.want)
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
			Bindable:    boolPtr(true),
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
		specifications := make([]map[string]interface{}, 0, apiDefCount+eventDefCount)

		for j := 0; j < apiDefCount; j++ {
			apiDefSpec := map[string]interface{}{
				"definition_id":          fmt.Sprintf("id%d", j),
				"definition_name":        fmt.Sprintf("apiDef%d", j),
				"specification_category": "api_definition",
				"specification_type":     schema.APISpecTypeOdata,
				"specification_format":   "application/json",
				"specification_url":      fmt.Sprintf("http://specification.com/specifications?app_id=id&package_id=id%d&definition_id=id%d", i, j),
			}
			specifications = append(specifications, apiDefSpec)
		}

		for j := 0; j < eventDefCount; j++ {
			eventDefSpec := map[string]interface{}{
				"definition_id":          fmt.Sprintf("id%d", i),
				"definition_name":        fmt.Sprintf("eventDef%d", i),
				"specification_category": "event_definition",
				"specification_type":     schema.EventSpecTypeAsyncAPI,
				"specification_format":   "application/json",
				"specification_url":      fmt.Sprintf("http://specification.com/specifications?app_id=id&package_id=id%d&definition_id=id%d", i, j),
			}
			specifications = append(specifications, eventDefSpec)
		}

		plan.Metadata.AdditionalMetadata = map[string]interface{}{
			"specifications": specifications,
		}
		plans = append(plans, plan)
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
				ID:   fmt.Sprintf("id%d", i),
				Name: fmt.Sprintf("apiDef%d", i),
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
				ID:   fmt.Sprintf("id%d", i),
				Name: fmt.Sprintf("eventDef%d", i),
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

func generateFaultyPackage() *schema.PackageExt {
	faultySchema := schema.JSONSchema(`NOT A JSON`)
	ext := generatePackages(1, 0, 0)
	ext[0].InstanceAuthRequestInputSchema = &faultySchema
	return ext[0]
}

func strToPtrStr(s string) *string {
	return &s
}
