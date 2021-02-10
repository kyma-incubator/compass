package model

import (
	"encoding/json"
)

type Package struct {
	ID                string
	TenantID          string
	ApplicationID     string
	OrdID             string
	Vendor            *string
	Title             string
	ShortDescription  string
	Description       string
	Version           string
	PackageLinks      json.RawMessage
	Links             json.RawMessage
	LicenseType       *string
	Tags              json.RawMessage
	Countries         json.RawMessage
	Labels            json.RawMessage
	PolicyLevel       string
	CustomPolicyLevel *string
	PartOfProducts    json.RawMessage
	LineOfBusiness    json.RawMessage
	Industry          json.RawMessage
}

type PackageInput struct {
	OrdID             string
	Vendor            *string
	Title             string
	ShortDescription  string
	Description       string
	Version           string
	PackageLinks      json.RawMessage
	Links             json.RawMessage
	LicenseType       *string
	Tags              json.RawMessage
	Countries         json.RawMessage
	Labels            json.RawMessage
	PolicyLevel       string
	CustomPolicyLevel *string
	PartOfProducts    json.RawMessage
	LineOfBusiness    json.RawMessage
	Industry          json.RawMessage
}

func (i *PackageInput) ToPackage(id, tenantID, appID string) *Package {
	if i == nil {
		return nil
	}

	return &Package{
		ID:                id,
		TenantID:          tenantID,
		ApplicationID:     appID,
		OrdID:             i.OrdID,
		Vendor:            i.Vendor,
		Title:             i.Title,
		ShortDescription:  i.ShortDescription,
		Description:       i.Description,
		Version:           i.Version,
		PackageLinks:      i.PackageLinks,
		Links:             i.Links,
		LicenseType:       i.LicenseType,
		Tags:              i.Tags,
		Countries:         i.Countries,
		Labels:            i.Labels,
		PolicyLevel:       i.PolicyLevel,
		CustomPolicyLevel: i.CustomPolicyLevel,
		PartOfProducts:    i.PartOfProducts,
		LineOfBusiness:    i.LineOfBusiness,
		Industry:          i.Industry,
	}
}

func (p *Package) SetFromUpdateInput(update PackageInput) {
	p.Vendor = update.Vendor
	p.Title = update.Title
	p.ShortDescription = update.ShortDescription
	p.Description = update.Description
	p.Version = update.Version
	p.PackageLinks = update.PackageLinks
	p.Links = update.Links
	p.LicenseType = update.LicenseType
	p.Tags = update.Tags
	p.Countries = update.Countries
	p.Labels = update.Labels
	p.PolicyLevel = update.PolicyLevel
	p.CustomPolicyLevel = update.CustomPolicyLevel
	p.PartOfProducts = update.PartOfProducts
	p.LineOfBusiness = update.LineOfBusiness
	p.Industry = update.Industry
}
