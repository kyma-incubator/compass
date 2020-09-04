package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"time"
)

type Package struct {
	ID               string
	TenantID         string
	ApplicationID    string
	Title            string
	ShortDescription string
	Description      string
	Version          string
	Licence          *string
	LicenceType      *string
	TermsOfService   *string
	Logo             *string
	Image            *string
	Provider         *string
	Actions          *string
	Tags             *string
	LastUpdated      time.Time
	Extensions       *string
}

func (pkg *Package) SetFromUpdateInput(update PackageUpdateInput) {
	pkg.Title = update.Title
	if update.ShortDescription != nil {
		pkg.ShortDescription = *update.ShortDescription
	}
	if update.Description != nil {
		pkg.Description = *update.Description
	}
	if update.Version != nil {
		pkg.Version = *update.Version
	}
	if update.Version != nil {
		pkg.Version = *update.Version
	}
	pkg.Licence = update.Licence
	pkg.LicenceType = update.LicenceType
	pkg.TermsOfService = update.TermsOfService
	pkg.Logo = update.Logo
	pkg.Image = update.Image
	pkg.Provider = update.Provider
	pkg.Actions = update.Actions
	pkg.Tags = update.Tags
	pkg.LastUpdated = update.LastUpdated
	pkg.Extensions = update.Extensions
}

type PackageCreateInput struct {
	Title            string
	ShortDescription string
	Description      string
	Version          string
	Licence          *string
	LicenceType      *string
	TermsOfService   *string
	Logo             *string
	Image            *string
	Provider         *string
	Actions          *string
	Tags             *string
	LastUpdated      time.Time
	Extensions       *string
	Bundles          []*BundleCreateInput
}

type PackageUpdateInput struct {
	Title            string
	ShortDescription *string
	Description      *string
	Version          *string
	Licence          *string
	LicenceType      *string
	TermsOfService   *string
	Logo             *string
	Image            *string
	Provider         *string
	Actions          *string
	Tags             *string
	LastUpdated      time.Time
	Extensions       *string
}

type PackagePage struct {
	Data       []*Package
	PageInfo   *pagination.Page
	TotalCount int
}

func (PackagePage) IsPageable() {}

func (i *PackageCreateInput) Package(id, applicationID, tenantID string) *Package {
	if i == nil {
		return nil
	}

	return &Package{
		ID:               id,
		TenantID:         tenantID,
		ApplicationID:    applicationID,
		Title:            i.Title,
		ShortDescription: i.ShortDescription,
		Description:      i.Description,
		Version:          i.Version,
		Licence:          i.Licence,
		LicenceType:      i.LicenceType,
		TermsOfService:   i.TermsOfService,
		Logo:             i.Logo,
		Image:            i.Image,
		Provider:         i.Provider,
		Actions:          i.Actions,
		Tags:             i.Tags,
		LastUpdated:      i.LastUpdated,
		Extensions:       i.Extensions,
	}
}
