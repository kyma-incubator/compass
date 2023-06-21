package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Product missing godoc
type Product struct {
	ID                           string
	OrdID                        string
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	Title                        string
	ShortDescription             string
	Vendor                       string
	Parent                       *string
	CorrelationIDs               json.RawMessage
	Tags                         json.RawMessage
	Labels                       json.RawMessage
	DocumentationLabels          json.RawMessage
}

// ProductInput missing godoc
type ProductInput struct {
	OrdID               string          `json:"ordId"`
	Title               string          `json:"title"`
	ShortDescription    string          `json:"shortDescription"`
	Vendor              string          `json:"vendor"`
	Parent              *string         `json:"parent"`
	CorrelationIDs      json.RawMessage `json:"correlationIds"`
	Tags                json.RawMessage `json:"tags"`
	Labels              json.RawMessage `json:"labels"`
	DocumentationLabels json.RawMessage `json:"documentationLabels"`
}

// ToProduct missing godoc
func (i *ProductInput) ToProduct(id string, resourceType resource.Type, resourceID string) *Product {
	if i == nil {
		return nil
	}

	product := &Product{
		ID:                  id,
		OrdID:               i.OrdID,
		Title:               i.Title,
		ShortDescription:    i.ShortDescription,
		Vendor:              i.Vendor,
		Parent:              i.Parent,
		CorrelationIDs:      i.CorrelationIDs,
		Tags:                i.Tags,
		Labels:              i.Labels,
		DocumentationLabels: i.DocumentationLabels,
	}

	if resourceType == resource.ApplicationTemplateVersion {
		product.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		product.ApplicationID = &resourceID
	}

	return product
}

// SetFromUpdateInput missing godoc
func (p *Product) SetFromUpdateInput(update ProductInput) {
	p.Title = update.Title
	p.ShortDescription = update.ShortDescription
	p.Vendor = update.Vendor
	p.Parent = update.Parent
	p.CorrelationIDs = update.CorrelationIDs
	p.Tags = update.Tags
	p.Labels = update.Labels
	p.DocumentationLabels = update.DocumentationLabels
}
