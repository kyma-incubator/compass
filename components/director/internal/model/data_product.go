package model

import (
	"encoding/json"
)

// DataProduct missing godoc
type DataProduct struct {
	ID               string
	ApplicationID    string
	OrdID            *string
	LocalID          *string
	Title            *string
	ShortDescription *string
	Description      *string
	Version          *string
	ReleaseStatus    *string
	Visibility       *string
	OrdPackageID     *string
	Tags             json.RawMessage
	Industry         json.RawMessage
	LineOfBusiness   json.RawMessage
	Type             *string
	DataProductOwner *string
}

// DataProductInput missing godoc
type DataProductInput struct {
	OrdID                    string                        `json:"ordId"`
	Tenant                   string                        `json:",omitempty"`
	LocalID                  *string                       `json:"localId"`
	Title                    *string                       `json:"title"`
	ShortDescription         *string                       `json:"shortDescription"`
	Description              *string                       `json:"description"`
	Version                  *string                       `json:"version"`
	ReleaseStatus            *string                       `json:"releaseStatus"`
	Visibility               *string                       `json:"visibility"`
	OrdPackageID             *string                       `json:"partOfPackage"`
	Tags                     json.RawMessage               `json:"tags"`
	Industry                 json.RawMessage               `json:"industry"`
	LineOfBusiness           json.RawMessage               `json:"lineOfBusiness"`
	Type                     *string                       `json:"type"`
	DataProductOwner         *string                       `json:"dataProductOwner"`
	InputPorts               []*PortInput                  `json:"inputPorts"`
	OutputPorts              []*PortInput                  `json:"outputPorts"`
	PartOfConsumptionBundles []*ConsumptionBundleReference `json:"partOfConsumptionBundles"`
	DefaultConsumptionBundle *string                       `json:"defaultConsumptionBundle"`
}

// Port missing godoc
type Port struct {
	ID                  string
	DataProductID       string
	ApplicationID       string
	Name                *string
	PortType            *string
	Description         *string
	ProducerCardinality *string
	Disabled            bool
}

// PortInput missing godoc
type PortInput struct {
	Name                *string     `json:"portName"`
	Description         *string     `json:"description"`
	ProducerCardinality *string     `json:"producerCardinality"`
	Disabled            bool        `json:"disabled"`
	ApiResources        []*Resource `json:"apiResources"`
	EventResources      []*Resource `json:"eventResources"`
}

type Resource struct {
	OrdID      string  `json:"ordId"`
	MinVersion *string `json:"minVersion"`
}

// ToDataProduct missing godoc
func (a *DataProductInput) ToDataProduct(id, appID string, packageID *string) *DataProduct {
	if a == nil {
		return nil
	}

	return &DataProduct{
		ID:               id,
		ApplicationID:    appID,
		OrdID:            &a.OrdID,
		LocalID:          a.LocalID,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Version:          a.Version,
		ReleaseStatus:    a.ReleaseStatus,
		Visibility:       a.Visibility,
		OrdPackageID:     packageID,
		Tags:             a.Tags,
		Industry:         a.Industry,
		LineOfBusiness:   a.LineOfBusiness,
		Type:             a.Type,
		DataProductOwner: a.DataProductOwner,
	}
}

// ToPort missing godoc
func (a *PortInput) ToPort(id, dataProductID, appID, portType string) *Port {
	if a == nil {
		return nil
	}

	return &Port{
		ID:                  id,
		DataProductID:       dataProductID,
		ApplicationID:       appID,
		Name:                a.Name,
		PortType:            &portType,
		Description:         a.Description,
		ProducerCardinality: a.ProducerCardinality,
		Disabled:            a.Disabled,
	}
}
