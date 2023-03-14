package model

import (
	"encoding/json"
)

// Vendor missing godoc
type Vendor struct {
	ID                  string
	OrdID               string
	ApplicationID       *string
	Title               string
	Partners            json.RawMessage
	Tags                json.RawMessage
	Labels              json.RawMessage
	DocumentationLabels json.RawMessage
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
func (i *VendorInput) ToVendor(id string, appID *string) *Vendor {
	if i == nil {
		return nil
	}

	return &Vendor{
		ID:                  id,
		OrdID:               i.OrdID,
		ApplicationID:       appID,
		Title:               i.Title,
		Partners:            i.Partners,
		Tags:                i.Tags,
		Labels:              i.Labels,
		DocumentationLabels: i.DocumentationLabels,
	}
}

// SetFromUpdateInput missing godoc
func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.Partners = update.Partners
	p.Tags = update.Tags
	p.Labels = update.Labels
	p.DocumentationLabels = update.DocumentationLabels
}
