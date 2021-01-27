package osb

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/specs"
)

type MapConverter struct {
}

func (c MapConverter) toApiDefMap(baseURL, appID, pkgID string, apiDef *graphql.APIDefinitionExt) (map[string]interface{}, error) {
	api := make(map[string]interface{})
	api["id"] = apiDef.ID
	api["name"] = apiDef.Name
	api["target_url"] = apiDef.TargetURL
	if apiDef.Spec != nil {
		specsFormatHeader, err := specs.SpecFormatToContentTypeHeader(apiDef.Spec.Format)
		if err != nil {
			return nil, err
		}
		specification := make(map[string]interface{})
		specification["type"] = apiDef.Spec.Type
		specification["format"] = specsFormatHeader
		specification["url"] = fmt.Sprintf("%s%s?%s=%s&%s=%s&%s=%s",
			baseURL, specs.SpecsAPI, specs.AppIDParameter, appID, specs.BundleIDParameter, pkgID, specs.DefinitionIDParameter, apiDef.ID)
		api["specification"] = specification
	}
	if apiDef.Description != nil && *apiDef.Description != "" {
		api["description"] = apiDef.Description
	}
	if apiDef.Group != nil && *apiDef.Group != "" {
		api["group"] = apiDef.Group
	}
	if apiDef.Version != nil {
		versionMap := c.toVersionMap(apiDef.Version)
		api["version"] = versionMap
	}
	return api, nil
}

func (c MapConverter) toEventDefMap(baseURL, appID, pkgID string, eventDef *graphql.EventAPIDefinitionExt) (map[string]interface{}, error) {
	event := make(map[string]interface{})
	event["id"] = eventDef.ID
	event["name"] = eventDef.Name
	if eventDef.Spec != nil {
		specsFormatHeader, err := specs.SpecFormatToContentTypeHeader(eventDef.Spec.Format)
		if err != nil {
			return nil, err
		}
		specification := make(map[string]interface{})
		specification["type"] = eventDef.Spec.Type
		specification["format"] = specsFormatHeader
		specification["url"] = fmt.Sprintf("%s%s?%s=%s&%s=%s&%s=%s",
			baseURL, specs.SpecsAPI, specs.AppIDParameter, appID, specs.BundleIDParameter, pkgID, specs.DefinitionIDParameter, eventDef.ID)
		event["specification"] = specification
	}
	if eventDef.Description != nil && *eventDef.Description != "" {
		event["description"] = eventDef.Description
	}
	if eventDef.Group != nil && *eventDef.Group != "" {
		event["group"] = eventDef.Group
	}
	if eventDef.Version != nil {
		versionMap := c.toVersionMap(eventDef.Version)
		event["version"] = versionMap
	}
	return event, nil
}

func (c MapConverter) toVersionMap(version *graphql.Version) map[string]interface{} {
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
