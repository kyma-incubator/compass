/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package osb

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type Converter struct {
	baseURL string
	MapConverter
}

func (c Converter) Convert(app *schema.ApplicationExt) (*domain.Service, error) {
	plans, err := c.toPlans(app.ID, app.Packages.Data)
	if err != nil {
		return nil, err
	}

	desc := ptrStrToStr(app.Description)
	if desc == "" {
		desc = fmt.Sprintf("service generated from system with name %s", app.Name)
	}

	return &domain.Service{
		ID:                   app.ID,
		Name:                 app.Name,
		Description:          desc,
		Bindable:             true,
		InstancesRetrievable: false,
		BindingsRetrievable:  true,
		PlanUpdatable:        false,
		Plans:                plans,
		Metadata:             c.toServiceMetadata(app),
	}, nil
}

func (c *Converter) toPlans(appID string, packages []*graphql.PackageExt) ([]domain.ServicePlan, error) {
	var plans []domain.ServicePlan
	for _, p := range packages {
		schemas, err := c.toSchemas(p)
		if err != nil {
			return nil, err
		}
		desc := ptrStrToStr(p.Description)
		if desc == "" {
			desc = fmt.Sprintf("plan generated from package with name %s", p.Name)
		}

		metadata, err := c.toPlanMetadata(appID, p)
		if err != nil {
			return nil, err
		}

		plan := domain.ServicePlan{
			ID:          p.ID,
			Name:        p.Name,
			Description: desc,
			Bindable:    boolToPtr(true),
			Metadata:    metadata,
			Schemas:     schemas,
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

//TODO these are probably too hidden, should be abstracted away
func (c *Converter) toPlanMetadata(appID string, pkg *graphql.PackageExt) (*domain.ServicePlanMetadata, error) {
	metadata := &domain.ServicePlanMetadata{
		AdditionalMetadata: make(map[string]interface{}),
	}

	apis := make([]map[string]interface{}, 0, 0)

	for _, apiDef := range pkg.APIDefinitions.Data {
		api, err := c.toApiDefMap(c.baseURL, appID, pkg.ID, apiDef)
		if err != nil {
			return nil, fmt.Errorf("while converting apidef to map: %w", err)
		}
		apis = append(apis, api)
	}
	metadata.AdditionalMetadata["api_specs"] = apis

	events := make([]map[string]interface{}, 0, 0)
	for _, eventDef := range pkg.EventDefinitions.Data {
		event, err := c.toEventDefMap(c.baseURL, appID, pkg.ID, eventDef)
		if err != nil {
			return nil, fmt.Errorf("while converting eventdef to map: %w", err)
		}
		events = append(events, event)
	}
	metadata.AdditionalMetadata["event_specs"] = events

	return metadata, nil
}

func (c *Converter) toServiceMetadata(app *schema.ApplicationExt) *domain.ServiceMetadata {
	if app.Labels == nil {
		app.Labels = map[string]interface{}{}
	}

	return &domain.ServiceMetadata{
		DisplayName:         app.Name,
		ProviderDisplayName: ptrStrToStr(app.ProviderName),
		AdditionalMetadata:  app.Labels,
	}
}

func (c *Converter) toSchemas(pkg *graphql.PackageExt) (*domain.ServiceSchemas, error) {
	if pkg.InstanceAuthRequestInputSchema == nil {
		return nil, nil
	}

	var unmarshalled map[string]interface{}
	err := json.Unmarshal([]byte(*pkg.InstanceAuthRequestInputSchema), &unmarshalled)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling JSON schema: %v", *pkg.InstanceAuthRequestInputSchema)
	}

	return &domain.ServiceSchemas{
		Instance: domain.ServiceInstanceSchema{
			Create: domain.Schema{
				Parameters: unmarshalled,
			},
		},
		Binding: domain.ServiceBindingSchema{},
	}, nil

}

func ptrStrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolToPtr(in bool) *bool {
	return &in
}
