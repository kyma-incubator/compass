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

type CatalogConverter struct {
	ORDServiceURL string
}

func (c CatalogConverter) Convert(app *schema.ApplicationExt) (*domain.Service, error) {
	plans, err := toPlans(c.ORDServiceURL, app.Bundles.Data)
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
		Metadata:             toServiceMetadata(app),
	}, nil
}

func toPlans(ORDServiceURL string, bundles []*graphql.BundleExt) ([]domain.ServicePlan, error) {
	var plans []domain.ServicePlan
	for _, p := range bundles {
		schemas, err := toSchemas(p)
		if err != nil {
			return nil, err
		}
		desc := ptrStrToStr(p.Description)
		if desc == "" {
			desc = fmt.Sprintf("plan generated from bundle with name %s", p.Name)
		}

		metadata, err := toPlanMetadata(ORDServiceURL, p)
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

func toPlanMetadata(ORDServiceURL string, pkg *graphql.BundleExt) (*domain.ServicePlanMetadata, error) {
	metadata := &domain.ServicePlanMetadata{
		AdditionalMetadata: make(map[string]interface{}),
	}

	apis := make([]map[string]interface{}, 0, 0)

	for _, apiDef := range pkg.APIDefinitions.Data {
		api, err := toApiDefMap(ORDServiceURL, apiDef)
		if err != nil {
			return nil, fmt.Errorf("while converting apidef to map: %w", err)
		}
		apis = append(apis, api)
	}
	metadata.AdditionalMetadata["api_specs"] = apis

	events := make([]map[string]interface{}, 0, 0)
	for _, eventDef := range pkg.EventDefinitions.Data {
		event, err := toEventDefMap(ORDServiceURL, eventDef)
		if err != nil {
			return nil, fmt.Errorf("while converting eventdef to map: %w", err)
		}
		events = append(events, event)
	}
	metadata.AdditionalMetadata["event_specs"] = events

	return metadata, nil
}

func toServiceMetadata(app *schema.ApplicationExt) *domain.ServiceMetadata {
	if app.Labels == nil {
		app.Labels = map[string]interface{}{}
	}

	return &domain.ServiceMetadata{
		DisplayName:         app.Name,
		ProviderDisplayName: ptrStrToStr(app.ProviderName),
		AdditionalMetadata:  app.Labels,
	}
}

func toSchemas(pkg *graphql.BundleExt) (*domain.ServiceSchemas, error) {
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

func toApiDefMap(ORDServiceURL string, apiDef *graphql.APIDefinitionExt) (map[string]interface{}, error) {
	api := make(map[string]interface{})
	api["id"] = apiDef.ID
	api["name"] = apiDef.Name
	api["target_url"] = apiDef.TargetURL
	if apiDef.Spec != nil {
		specsFormatHeader, err := specFormatToContentTypeHeader(apiDef.Spec.Format)
		if err != nil {
			return nil, err
		}
		specification := make(map[string]interface{})
		specification["type"] = apiDef.Spec.Type
		specification["format"] = specsFormatHeader
		specification["url"] = fmt.Sprintf("%s/api/%s/specification/%s",
			ORDServiceURL, apiDef.ID, apiDef.Spec.ID)
		api["specification"] = specification
	}
	if apiDef.Description != nil && *apiDef.Description != "" {
		api["description"] = apiDef.Description
	}
	if apiDef.Group != nil && *apiDef.Group != "" {
		api["group"] = apiDef.Group
	}
	if apiDef.Version != nil {
		versionMap := toVersionMap(apiDef.Version)
		api["version"] = versionMap
	}
	return api, nil
}

func toEventDefMap(baseURL string, eventDef *graphql.EventAPIDefinitionExt) (map[string]interface{}, error) {
	event := make(map[string]interface{})
	event["id"] = eventDef.ID
	event["name"] = eventDef.Name
	if eventDef.Spec != nil {
		specsFormatHeader, err := specFormatToContentTypeHeader(eventDef.Spec.Format)
		if err != nil {
			return nil, err
		}
		specification := make(map[string]interface{})
		specification["type"] = eventDef.Spec.Type
		specification["format"] = specsFormatHeader
		specification["url"] = fmt.Sprintf("%s/event/%s/specification/%s",
			baseURL, eventDef.ID, eventDef.Spec.ID)
		event["specification"] = specification
	}
	if eventDef.Description != nil && *eventDef.Description != "" {
		event["description"] = eventDef.Description
	}
	if eventDef.Group != nil && *eventDef.Group != "" {
		event["group"] = eventDef.Group
	}
	if eventDef.Version != nil {
		versionMap := toVersionMap(eventDef.Version)
		event["version"] = versionMap
	}
	return event, nil
}

func toVersionMap(version *graphql.Version) map[string]interface{} {
	m := make(map[string]interface{})
	m["value"] = version.Value
	if version.Deprecated != nil {
		m["deprecated"] = version.Deprecated
	}
	if version.DeprecatedSince != nil {
		m["deprecated_since"] = version.DeprecatedSince
	}
	if version.ForRemoval != nil {
		m["for_removal"] = version.ForRemoval
	}
	return m
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

func specFormatToContentTypeHeader(format graphql.SpecFormat) (string, error) {
	switch format {
	case schema.SpecFormatJSON:
		return "application/json", nil
	case schema.SpecFormatXML:
		return "application/xml", nil
	case schema.SpecFormatYaml:
		return "text/yaml", nil
	}

	return "", errors.Errorf("unknown spec format %s", format)
}
