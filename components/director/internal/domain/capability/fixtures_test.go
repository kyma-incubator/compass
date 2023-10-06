package capability_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"time"
)

const (
	capabilityID                                   = "cccccccccc-cccc-cccc-cccc-cccccccccccc"
	specID                                         = "sssssssss-ssss-ssss-ssss-ssssssssssss"
	tenantID                                       = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID                               = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	packageID                                      = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
	ordID                                          = "com.compass.ord.v1"
	localTenantID                                  = "localTenantID"
	resourceHash                                   = "123456"
	publicVisibility                               = "public"
	CapabilityTypeMDICapabilityDefinitionV1 string = "sap.mdo:mdi-capability:v1"
)

var (
	fixedTimestamp = time.Now()
	appID          = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	appTemplateVersionID = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
)

func fixCapabilityModel(id, name string) *model.Capability {
	return &model.Capability{
		Name:       name,
		Type:       CapabilityTypeMDICapabilityDefinitionV1,
		BaseEntity: &model.BaseEntity{ID: id},
		Visibility: str.Ptr(publicVisibility),
	}
}

func fixCapabilityWithPackageModel(id, name string) *model.Capability {
	return &model.Capability{
		PackageID: str.Ptr(packageID),
		Name:      name,
		Type:      CapabilityTypeMDICapabilityDefinitionV1,
		Version:   &model.Version{},
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}

func fixFullCapabilityModelWithAppID(name string) (model.Capability, model.Spec) {
	capability, spec := fixFullCapabilityModel(capabilityID, name)
	capability.ApplicationID = str.Ptr(appID)

	return capability, spec
}

func fixFullCapabilityModelWithAppTemplateVersionID(name string) (model.Capability, model.Spec) {
	capability, spec := fixFullCapabilityModel(capabilityID, name)
	capability.ApplicationTemplateVersionID = str.Ptr(appTemplateVersionID)

	return capability, spec
}

func fixFullCapabilityModel(capabilityID, name string) (model.Capability, model.Spec) {
	capabilityType := model.CapabilitySpecTypeMDICapabilityDefinitionV1
	spec := model.Spec{
		ID:             specID,
		Data:           str.Ptr("spec_data_" + name),
		Format:         model.SpecFormatYaml,
		ObjectType:     model.CapabilitySpecReference,
		ObjectID:       capabilityID,
		CapabilityType: &capabilityType,
	}

	deprecated := false
	forRemoval := false

	v := &model.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	boolVar := false
	return model.Capability{
		PackageID:           str.Ptr(packageID),
		Name:                name,
		Description:         str.Ptr("desc_" + name),
		OrdID:               str.Ptr(ordID),
		Type:                CapabilityTypeMDICapabilityDefinitionV1,
		CustomType:          nil,
		LocalTenantID:       str.Ptr(localTenantID),
		ShortDescription:    str.Ptr("shortDescription"),
		SystemInstanceAware: &boolVar,
		Tags:                json.RawMessage("[]"),
		Links:               json.RawMessage("[]"),
		ReleaseStatus:       str.Ptr("releaseStatus"),
		Labels:              json.RawMessage("[]"),
		Visibility:          str.Ptr(publicVisibility),
		Version:             v,
		ResourceHash:        str.Ptr(resourceHash),
		DocumentationLabels: json.RawMessage("[]"),
		CorrelationIDs:      json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        capabilityID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}, spec
}

func fixEntityCapability(id, name string) *capability.Entity {
	return &capability.Entity{
		Name:       name,
		Type:       CapabilityTypeMDICapabilityDefinitionV1,
		BaseEntity: &repo.BaseEntity{ID: id},
		Visibility: publicVisibility,
	}
}

func fixFullEntityCapabilityWithAppID(capabilityID, name string) capability.Entity {
	entity := fixFullEntityCapability(capabilityID, name)
	entity.ApplicationID = repo.NewValidNullableString(appID)

	return entity
}

func fixFullEntityCapabilityWithAppTemplateVersionID(capabilityID, name string) capability.Entity {
	entity := fixFullEntityCapability(capabilityID, name)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)

	return entity
}

func fixFullEntityCapability(capabilityID, name string) capability.Entity {
	return capability.Entity{
		PackageID:           repo.NewValidNullableString(packageID),
		Name:                name,
		Description:         repo.NewValidNullableString("desc_" + name),
		OrdID:               repo.NewValidNullableString(ordID),
		Type:                CapabilityTypeMDICapabilityDefinitionV1,
		LocalTenantID:       repo.NewValidNullableString(localTenantID),
		ShortDescription:    repo.NewValidNullableString("shortDescription"),
		SystemInstanceAware: repo.NewValidNullableBool(false),
		Tags:                repo.NewValidNullableString("[]"),
		Links:               repo.NewValidNullableString("[]"),
		ReleaseStatus:       repo.NewValidNullableString("releaseStatus"),
		Labels:              repo.NewValidNullableString("[]"),
		Visibility:          publicVisibility,
		Version: version.Version{
			Value:           repo.NewNullableString(str.Ptr("v1.1")),
			Deprecated:      repo.NewValidNullableBool(false),
			DeprecatedSince: repo.NewNullableString(str.Ptr("v1.0")),
			ForRemoval:      repo.NewValidNullableBool(false),
		},
		ResourceHash:        repo.NewValidNullableString(resourceHash),
		DocumentationLabels: repo.NewValidNullableString("[]"),
		CorrelationIDs:      repo.NewValidNullableString("[]"),
		BaseEntity: &repo.BaseEntity{
			ID:        capabilityID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
	}
}

func fixCapabilityColumns() []string {
	return []string{
		"id", "app_id", "app_template_version_id", "package_id", "name", "description", "ord_id", "type", "custom_type", "local_tenant_id",
		"short_description", "system_instance_aware", "tags", "links", "release_status", "labels", "visibility",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash", "documentation_labels", "correlation_ids"}
}

func fixCapabilityRow(id, name string) []driver.Value {
	boolVar := false
	return []driver.Value{id, appID, repo.NewValidNullableString(""), packageID, name, "desc_" + name, ordID, CapabilityTypeMDICapabilityDefinitionV1, nil, localTenantID,
		"shortDescription", &boolVar, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "releaseStatus",
		repo.NewValidNullableString("[]"), publicVisibility, "v1.1", false, "v1.0", false, true, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(resourceHash),
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixCapabilityRowForAppTemplateVersion(id, name string) []driver.Value {
	boolVar := false
	return []driver.Value{id, repo.NewValidNullableString(""), appTemplateVersionID, packageID, name, "desc_" + name, ordID, CapabilityTypeMDICapabilityDefinitionV1, nil, localTenantID,
		"shortDescription", &boolVar, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "releaseStatus",
		repo.NewValidNullableString("[]"), publicVisibility, "v1.1", false, "v1.0", false, true, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(resourceHash),
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixCapabilityCreateArgs(id string, capability *model.Capability) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(""), packageID, capability.Name, capability.Description,
		capability.OrdID, capability.Type, capability.CustomType, capability.LocalTenantID, capability.ShortDescription, capability.SystemInstanceAware, repo.NewNullableStringFromJSONRawMessage(capability.Tags),
		repo.NewNullableStringFromJSONRawMessage(capability.Links), capability.ReleaseStatus, repo.NewNullableStringFromJSONRawMessage(capability.Labels), capability.Visibility,
		capability.Version.Value, capability.Version.Deprecated, capability.Version.DeprecatedSince, capability.Version.ForRemoval, capability.Ready, capability.CreatedAt, capability.UpdatedAt, capability.DeletedAt, capability.Error, capability.ResourceHash, repo.NewNullableStringFromJSONRawMessage(capability.DocumentationLabels),
		repo.NewNullableStringFromJSONRawMessage(capability.CorrelationIDs)}
}
