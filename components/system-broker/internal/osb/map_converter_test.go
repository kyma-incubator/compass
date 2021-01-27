package osb

import (
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	log "github.com/sirupsen/logrus"
)

func addGroupAndVersionToBundle(ext *schema.BundleExt) *schema.BundleExt {
	ext.APIDefinitions.Data[0].Group = strToPtrStr("group")
	ext.APIDefinitions.Data[0].Version = &schema.Version{
		Value:           "v1",
		Deprecated:      boolToPtr(true),
		DeprecatedSince: strToPtrStr("01.01.2021"),
		ForRemoval:      boolToPtr(false),
	}

	ext.EventDefinitions.Data[0].Group = strToPtrStr("group")
	ext.EventDefinitions.Data[0].Version = &schema.Version{
		Value:           "v1",
		Deprecated:      boolToPtr(true),
		DeprecatedSince: strToPtrStr("01.01.2021"),
		ForRemoval:      boolToPtr(false),
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
	versionMap["deprecated"] = boolToPtr(true)
	versionMap["deprecated_since"] = strToPtrStr("01.01.2021")
	versionMap["for_removal"] = boolToPtr(false)

	apiSpecsSlice[0]["group"] = strToPtrStr("group")
	eventSpecsSlice[0]["group"] = strToPtrStr("group")

	apiSpecsSlice[0]["version"] = versionMap
	eventSpecsSlice[0]["version"] = versionMap

	s.Metadata.AdditionalMetadata["event_specs"] = eventSpecsSlice
	s.Metadata.AdditionalMetadata["api_specs"] = apiSpecsSlice

	return s
}
