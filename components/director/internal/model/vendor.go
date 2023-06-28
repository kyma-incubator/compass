package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Vendor missing godoc
type Vendor struct {
	ID                           string
	OrdID                        string
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	Title                        string
	Partners                     json.RawMessage
	Tags                         json.RawMessage
	Labels                       json.RawMessage
	DocumentationLabels          json.RawMessage
}

// VendorInput missing godoc
type VendorInput struct {
	OrdID               string          `json:"ordId"`
	Title               string          `json:"title"`
	Partners            json.RawMessage `json:"partners"`
	Tags                json.RawMessage `json:"tags"`
	Labels              json.RawMessage `json:"labels"`
	DocumentationLabels json.RawMessage `json:"documentationLabels"`
}

// ToVendor missing godoc
func (i *VendorInput) ToVendor(id string, resourceType resource.Type, resourceID string) *Vendor {
	if i == nil {
		return nil
	}

	vendor := &Vendor{
		ID:                  id,
		OrdID:               i.OrdID,
		Title:               i.Title,
		Partners:            i.Partners,
		Tags:                i.Tags,
		Labels:              i.Labels,
		DocumentationLabels: i.DocumentationLabels,
	}

	if resourceType == resource.ApplicationTemplateVersion {
		vendor.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		vendor.ApplicationID = &resourceID
	}

	return vendor
}

// SetFromUpdateInput missing godoc
func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.Partners = update.Partners
	p.Tags = update.Tags
	p.Labels = update.Labels
	p.DocumentationLabels = update.DocumentationLabels
}
