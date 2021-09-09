package model

import (
	"encoding/json"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

// Package missing godoc
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
	ResourceHash      *string
}

// PackageInput missing godoc
type PackageInput struct {
	OrdID             string          `json:"ordId"`
	Vendor            *string         `json:"vendor"`
	Title             string          `json:"title"`
	ShortDescription  string          `json:"shortDescription"`
	Description       string          `json:"description"`
	Version           string          `json:"version" hash:"ignore"`
	PackageLinks      json.RawMessage `json:"packageLinks"`
	Links             json.RawMessage `json:"links"`
	LicenseType       *string         `json:"licenseType"`
	Tags              json.RawMessage `json:"tags"`
	Countries         json.RawMessage `json:"countries"`
	Labels            json.RawMessage `json:"labels"`
	PolicyLevel       string          `json:"policyLevel"`
	CustomPolicyLevel *string         `json:"customPolicyLevel"`
	PartOfProducts    json.RawMessage `json:"partOfProducts"`
	LineOfBusiness    json.RawMessage `json:"lineOfBusiness"`
	Industry          json.RawMessage `json:"industry"`
}

// ToPackage missing godoc
func (i *PackageInput) ToPackage(id, tenantID, appID string, pkgHash uint64) *Package {
	if i == nil {
		return nil
	}

	var hash *string
	if pkgHash != 0 {
		hash = str.Ptr(strconv.FormatUint(pkgHash, 10))
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
		ResourceHash:      hash,
	}
}

// SetFromUpdateInput missing godoc
func (p *Package) SetFromUpdateInput(update PackageInput, pkgHash uint64) {
	var hash *string
	if pkgHash != 0 {
		hash = str.Ptr(strconv.FormatUint(pkgHash, 10))
	}

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
	p.ResourceHash = hash
}
