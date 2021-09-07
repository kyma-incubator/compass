package model

import (
	"encoding/json"
)

type Product struct {
	ID               string
	OrdID            string
	TenantID         string
	ApplicationID    string
	Title            string
	ShortDescription string
	Vendor           string
	Parent           *string
	CorrelationIDs   json.RawMessage
	Labels           json.RawMessage
}

type ProductInput struct {
	OrdID            string          `json:"ordId"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Vendor           string          `json:"vendor"`
	Parent           *string         `json:"parent"`
	CorrelationIDs   json.RawMessage `json:"correlationIds"`
	Labels           json.RawMessage `json:"labels"`
}

func (i *ProductInput) ToProduct(id, tenantID, appID string) *Product {
	if i == nil {
		return nil
	}

	return &Product{
		ID:               id,
		OrdID:            i.OrdID,
		TenantID:         tenantID,
		ApplicationID:    appID,
		Title:            i.Title,
		ShortDescription: i.ShortDescription,
		Vendor:           i.Vendor,
		Parent:           i.Parent,
		CorrelationIDs:   i.CorrelationIDs,
		Labels:           i.Labels,
	}
}

func (p *Product) SetFromUpdateInput(update ProductInput) {
	p.Title = update.Title
	p.ShortDescription = update.ShortDescription
	p.Vendor = update.Vendor
	p.Parent = update.Parent
	p.CorrelationIDs = update.CorrelationIDs
	p.Labels = update.Labels
}
