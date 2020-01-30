package broker

import (
	"context"
	"encoding/json"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

// OptionalComponentNamesProvider provides optional components names
type OptionalComponentNamesProvider interface {
	GetAllOptionalComponentsNames() []string
}

type ServicesEndpoint struct {
	dumper             StructDumper
	optionalComponents OptionalComponentNamesProvider
	enabledPlanIDs     map[string]struct{}
}

func NewServices(cfg Config, optComponentsSvc OptionalComponentNamesProvider, dumper StructDumper) *ServicesEndpoint {
	enabledPlanIDs := map[string]struct{}{}
	for _, planName := range cfg.EnablePlans {
		id := planIDsMapping[planName]
		enabledPlanIDs[id] = struct{}{}
	}

	return &ServicesEndpoint{
		dumper:             dumper,
		optionalComponents: optComponentsSvc,
		enabledPlanIDs:     enabledPlanIDs,
	}
}

// Services gets the catalog of services offered by the service broker
//   GET /v2/catalog
func (b *ServicesEndpoint) Services(ctx context.Context) ([]domain.Service, error) {
	var availableServicePlans []domain.ServicePlan

	for _, plan := range plans {
		// filter out not enabled plans
		if _, exists := b.enabledPlanIDs[plan.planDefinition.ID]; !exists {
			continue
		}
		p := plan.planDefinition
		err := json.Unmarshal(plan.provisioningRawSchema, &p.Schemas.Instance.Create.Parameters)
		b.addComponentsToSchema(&p.Schemas.Instance.Create.Parameters)
		if err != nil {
			b.dumper.Dump("Could not decode provisioning schema:", err.Error())
			return nil, err
		}
		availableServicePlans = append(availableServicePlans, p)
	}

	return []domain.Service{
		{
			ID:          kymaServiceID,
			Name:        "kymaruntime",
			Description: "[EXPERIMENTAL] Service Class for Kyma Runtime",
			Bindable:    true,
			Plans:       availableServicePlans,
			Metadata: &domain.ServiceMetadata{
				DisplayName:         "Kyma Runtime",
				LongDescription:     "Kyma Runtime experimental service class",
				DocumentationUrl:    "kyma-project.io",
				ProviderDisplayName: "SAP",
			},
			Tags: []string{
				"SAP",
				"Kyma",
			},
		},
	}, nil
}

func (b *ServicesEndpoint) addComponentsToSchema(schema *map[string]interface{}) {
	props := (*schema)["properties"].(map[string]interface{})
	props["components"] = map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
			"enum": b.optionalComponents.GetAllOptionalComponentsNames(),
		},
	}
}
