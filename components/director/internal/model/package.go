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

func (pkg *Package) SetFromUpdateInput(update PackageInput) {
	pkg.Title = update.Title
	pkg.ShortDescription = update.ShortDescription
	pkg.Description = update.Description
	pkg.Version = update.Version
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

type PackageInput struct {
	ID               string
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
	Bundles          []*BundleInput
}

type PackagePage struct {
	Data       []*Package
	PageInfo   *pagination.Page
	TotalCount int
}

func (PackagePage) IsPageable() {}

func (i *PackageInput) Package(applicationID, tenantID string) *Package {
	if i == nil {
		return nil
	}

	return &Package{
		ID:               i.ID,
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
