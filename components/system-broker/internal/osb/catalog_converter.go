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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type Converter struct {
}

func (c Converter) Convert(app *graphql.ApplicationExt) ([]domain.Service, error) {
	plans, err := c.toPlans(app.Packages.Data)
	if err != nil {
		return nil, err
	}

	return []domain.Service{
		{
			ID:                   app.ID,
			Name:                 app.Name,
			Description:          ptrStrToStr(app.Description),
			Bindable:             true,
			InstancesRetrievable: false,
			BindingsRetrievable:  false,
			PlanUpdatable:        false,
			Plans:                plans,
			Metadata:             c.toServiceMetadata(app),
		},
	}, nil
}

func (c *Converter) toPlans(packages []*graphql.PackageExt) ([]domain.ServicePlan, error) {
	var plans []domain.ServicePlan
	for _, p := range packages {

		schemas, err := c.toSchemas(p)
		if err != nil {
			return nil, err
		}
		plan := domain.ServicePlan{
			ID:          p.ID,
			Name:        p.Name,
			Description: ptrStrToStr(p.Description),
			Bindable:    boolPtr(true),
			Metadata: &domain.ServicePlanMetadata{
				DisplayName: p.Name,
			},
			Schemas: schemas,
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

func (c *Converter) toServiceMetadata(app *graphql.ApplicationExt) *domain.ServiceMetadata {
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

	var unmarshaled map[string]interface{}
	err := json.Unmarshal([]byte(*pkg.InstanceAuthRequestInputSchema), &unmarshaled)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling JSON schema: %v", *pkg.InstanceAuthRequestInputSchema)
	}

	return &domain.ServiceSchemas{
		Instance: domain.ServiceInstanceSchema{
			Create: domain.Schema{
				Parameters: unmarshaled,
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

func boolPtr(in bool) *bool {
	return &in
}
