package model

import (
	"encoding/json"
)

type Vendor struct {
	OrdID         string
	TenantID      string
	ApplicationID string
	Title         string
	Partners      json.RawMessage
	Labels        json.RawMessage
}

type VendorInput struct {
	OrdID    string          `json:"ordId"`
	Title    string          `json:"title"`
	Partners json.RawMessage `json:"partners"`
	Labels   json.RawMessage `json:"labels"`
}

func (i *VendorInput) ToVendor(tenantID, appID string) *Vendor {
	if i == nil {
		return nil
	}

	return &Vendor{
		OrdID:         i.OrdID,
		TenantID:      tenantID,
		ApplicationID: appID,
		Title:         i.Title,
		Partners:      i.Partners,
		Labels:        i.Labels,
	}
}

func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.Partners = update.Partners
	p.Labels = update.Labels
}
