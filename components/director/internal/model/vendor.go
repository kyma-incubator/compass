package model

import (
	"encoding/json"
)

type Vendor struct {
	OrdID         string
	TenantID      string
	ApplicationID string
	Title         string
	SapPartner    *bool
	Labels        json.RawMessage
}

type VendorInput struct {
	OrdID      string          `json:"ordId"`
	Title      string          `json:"title"`
	SapPartner *bool           `json:"sapPartner"`
	Labels     json.RawMessage `json:"labels"`
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
		SapPartner:    i.SapPartner,
		Labels:        i.Labels,
	}
}

func (p *Vendor) SetFromUpdateInput(update VendorInput) {
	p.Title = update.Title
	p.SapPartner = update.SapPartner
	p.Labels = update.Labels
}
