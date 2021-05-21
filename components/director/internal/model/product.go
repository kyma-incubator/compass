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
	CorrelationIds   json.RawMessage
	Labels           json.RawMessage
}

type ProductInput struct {
	OrdID            string          `json:"ordId"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Vendor           string          `json:"vendor"`
	Parent           *string         `json:"parent"`
	CorrelationIds   json.RawMessage `json:"correlationIds"`
	Labels           json.RawMessage `json:"labels"`
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
		CorrelationIds:   i.CorrelationIds,
		Labels:           i.Labels,
	}
}

func (p *Product) SetFromUpdateInput(update ProductInput) {
	p.Title = update.Title
	p.ShortDescription = update.ShortDescription
	p.Vendor = update.Vendor
	p.Parent = update.Parent
	p.CorrelationIds = update.CorrelationIds
	p.Labels = update.Labels
}
