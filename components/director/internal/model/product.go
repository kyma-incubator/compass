package model

import (
	"encoding/json"
)

type Product struct {
	OrdID            string
	TenantID         string
	ApplicationID    string
	Title            string
	ShortDescription string
	Vendor           string
	Parent           *string
	PPMSObjectID     *string
	Labels           json.RawMessage
}

type ProductInput struct {
	OrdID            string
	TenantID         string
	ApplicationID    string
	Title            string
	ShortDescription string
	Vendor           string
	Parent           *string
	PPMSObjectID     *string
	Labels           json.RawMessage
}

func (i *ProductInput) ToProduct(tenantID, appID string) *Product {
	if i == nil {
		return nil
	}

	return &Product{
		OrdID:            i.OrdID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            i.Title,
		ShortDescription: i.ShortDescription,
		Vendor:           i.Vendor,
		Parent:           i.Parent,
		PPMSObjectID:     i.PPMSObjectID,
		Labels:           i.Labels,
	}
}

func (p *Product) SetFromUpdateInput(update ProductInput) {
	p.Title = update.Title
	p.ShortDescription = update.ShortDescription
	p.Vendor = update.Vendor
	p.Parent = update.Parent
	p.PPMSObjectID = update.PPMSObjectID
	p.Labels = update.Labels
}
