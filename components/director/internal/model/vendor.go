package model

import (
	"encoding/json"
)

type Vendor struct {
	OrdID         string
	TenantID      string
	ApplicationID string
	Title         string
	Type          string
	Labels        json.RawMessage
}

type VendorInput struct {
	OrdID         string
	TenantID      string
	ApplicationID string
	Title         string
	Type          string
	Labels        json.RawMessage
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
		Type:          i.Type,
		Labels:        i.Labels,
	}
}

func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.Type = update.Type
	p.Labels = update.Labels
}
