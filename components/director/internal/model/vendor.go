package model

import (
	"encoding/json"
)

// Vendor missing godoc
type Vendor struct {
	ID            string
	OrdID         string
	ApplicationID string
	Title         string
	Partners      json.RawMessage
	Labels        json.RawMessage
}

// VendorInput missing godoc
type VendorInput struct {
	OrdID    string          `json:"ordId"`
	Title    string          `json:"title"`
	Partners json.RawMessage `json:"partners"`
	Labels   json.RawMessage `json:"labels"`
}

// ToVendor missing godoc
func (i *VendorInput) ToVendor(id, appID string) *Vendor {
	if i == nil {
		return nil
	}

	return &Vendor{
		ID:            id,
		OrdID:         i.OrdID,
		ApplicationID: appID,
		Title:         i.Title,
		Partners:      i.Partners,
		Labels:        i.Labels,
	}
}

// SetFromUpdateInput missing godoc
func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.Partners = update.Partners
	p.Labels = update.Labels
}
