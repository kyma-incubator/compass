package model

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"regexp"
	"strconv"
)

// Capability missing godoc
type Capability struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	PackageID                    *string
	Name                         string
	Description                  *string
	OrdID                        *string
	Type                         string
	CustomType                   *string
	LocalTenantID                *string
	ShortDescription             *string
	SystemInstanceAware          *bool
	Tags                         json.RawMessage
	Links                        json.RawMessage
	ReleaseStatus                *string
	Labels                       json.RawMessage
	Visibility                   *string
	Version                      *Version
	ResourceHash                 *string
	DocumentationLabels          json.RawMessage
	CorrelationIDs               json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*Capability) GetType() resource.Type {
	return resource.Capability
}

// CapabilityInput missing godoc
type CapabilityInput struct {
	OrdPackageID          *string                 `json:"partOfPackage"`
	Tenant                string                  `json:",omitempty"`
	Name                  string                  `json:"title"`
	Description           *string                 `json:"description"`
	OrdID                 *string                 `json:"ordId"`
	Type                  string                  `json:"type"`
	CustomType            *string                 `json:"customType"`
	LocalTenantID         *string                 `json:"localTenantId"`
	ShortDescription      *string                 `json:"shortDescription"`
	SystemInstanceAware   *bool                   `json:"systemInstanceAware"`
	Tags                  json.RawMessage         `json:"tags"`
	Links                 json.RawMessage         `json:"links"`
	ReleaseStatus         *string                 `json:"releaseStatus"`
	Labels                json.RawMessage         `json:"labels"`
	Visibility            *string                 `json:"visibility"`
	CapabilityDefinitions []*CapabilityDefinition `json:"definitions"`
	DocumentationLabels   json.RawMessage         `json:"documentationLabels"`
	CorrelationIDs        json.RawMessage         `json:"correlationIds,omitempty"`

	*VersionInput `hash:"ignore"`
}

// CapabilityDefinition missing godoc
type CapabilityDefinition struct {
	Type           CapabilitySpecType              `json:"type"`
	CustomType     string                          `json:"customType"`
	MediaType      SpecFormat                      `json:"mediaType"`
	URL            string                          `json:"url"`
	AccessStrategy accessstrategy.AccessStrategies `json:"accessStrategies"`
}

// Validate missing godoc
func (cd *CapabilityDefinition) Validate() error {
	const CustomTypeRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	return validation.ValidateStruct(cd,
		validation.Field(&cd.Type, validation.Required, validation.In(CapabilitySpecTypeMDICapabilityDefinitionV1, CapabilitySpecTypeCustom), validation.When(cd.CustomType != "", validation.In(CapabilitySpecTypeCustom))),
		validation.Field(&cd.CustomType, validation.When(cd.CustomType != "", validation.Match(regexp.MustCompile(CustomTypeRegex))),
			validation.Length(0, 255),
			validation.When(cd.Type == CapabilitySpecTypeCustom, validation.Required)),
		validation.Field(&cd.MediaType, validation.Required, validation.In(SpecFormatApplicationJSON, SpecFormatTextYAML, SpecFormatApplicationXML, SpecFormatPlainText, SpecFormatOctetStream),
			validation.When(cd.Type == CapabilitySpecTypeMDICapabilityDefinitionV1, validation.In(SpecFormatApplicationJSON))),
		validation.Field(&cd.URL, validation.Required, is.RequestURI),
		validation.Field(&cd.AccessStrategy, validation.Required),
	)
}

// ToSpec missing godoc
func (cd *CapabilityDefinition) ToSpec() *SpecInput {
	var auth *AuthInput
	if as, ok := cd.AccessStrategy.GetSupported(); ok {
		asString := string(as)
		auth = &AuthInput{
			AccessStrategy: &asString,
		}
	}

	return &SpecInput{
		Format:         cd.MediaType,
		CapabilityType: &cd.Type,
		CustomType:     &cd.CustomType,
		FetchRequest: &FetchRequestInput{
			URL:  cd.URL,
			Auth: auth,
		},
	}
}

// ToCapability missing godoc
func (a *CapabilityInput) ToCapability(id string, resourceType resource.Type, resourceID string, packageID *string, capabilityHash uint64) *Capability {
	if a == nil {
		return nil
	}

	var hash *string
	if capabilityHash != 0 {
		hash = str.Ptr(strconv.FormatUint(capabilityHash, 10))
	}

	capability := &Capability{
		PackageID:           packageID,
		Name:                a.Name,
		Description:         a.Description,
		OrdID:               a.OrdID,
		Type:                a.Type,
		CustomType:          a.CustomType,
		LocalTenantID:       a.LocalTenantID,
		ShortDescription:    a.ShortDescription,
		SystemInstanceAware: a.SystemInstanceAware,
		Tags:                a.Tags,
		Links:               a.Links,
		ReleaseStatus:       a.ReleaseStatus,
		Labels:              a.Labels,
		Visibility:          a.Visibility,
		Version:             a.VersionInput.ToVersion(),
		DocumentationLabels: a.DocumentationLabels,
		CorrelationIDs:      a.CorrelationIDs,
		ResourceHash:        hash,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType.IsTenantIgnorable() {
		capability.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		capability.ApplicationID = &resourceID
	}

	return capability
}
