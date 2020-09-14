package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID               string
	OpenDiscoveryID  string
	BundleID         string
	Tenant           string
	Title            string
	ShortDescription string
	Description      *string
	Specs            []*APISpec
	EntryPoint       string
	//  group allows you to find the same API but in different version
	Group            *string
	Version          *Version
	APIDefinitions   string
	Tags             *string
	Documentation    *string
	ChangelogEntries *string
	Logo             *string
	Image            *string
	URL              *string
	ReleaseStatus    string
	APIProtocol      string
	Actions          string
	LastUpdated      time.Time
	Extensions       *string
}

type APISpec struct {
	ID string
	// when fetch request specified, data will be automatically populated
	Data   *string
	Format SpecFormat
	Type   APISpecType
}

type APISpecType string

const (
	APISpecTypeOdata   APISpecType = "ODATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"
)

type Timestamp time.Time

type APIDefinitionInput struct {
	ID               string
	OpenDiscoveryID  string
	Title            string
	ShortDescription string
	Description      *string
	EntryPoint       string
	Group            *string
	Specs            []*APISpecInput
	Version          *VersionInput
	APIDefinitions   string
	Tags             *string
	Documentation    *string
	ChangelogEntries *string
	Logo             *string
	Image            *string
	URL              *string
	ReleaseStatus    string
	APIProtocol      string
	Actions          string
	LastUpdated      time.Time
	Extensions       *string
}

type APISpecInput struct {
	Data         *string
	Type         APISpecType
	Format       SpecFormat
	FetchRequest *FetchRequestInput
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinitionWithinBundle(bundleID string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	specs := make([]*APISpec, 0, 0)
	for _, spec := range a.Specs {
		specs = append(specs, spec.ToAPISpec())
	}

	return &APIDefinition{
		ID:               a.ID,
		OpenDiscoveryID:  a.OpenDiscoveryID,
		BundleID:         bundleID,
		Tenant:           tenant,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Specs:            specs,
		EntryPoint:       a.EntryPoint,
		Group:            a.Group,
		Version:          a.Version.ToVersion(),
		APIDefinitions:   a.APIDefinitions,
		Tags:             a.Tags,
		Documentation:    a.Documentation,
		ChangelogEntries: a.ChangelogEntries,
		Logo:             a.Logo,
		Image:            a.Image,
		ReleaseStatus:    a.ReleaseStatus,
		APIProtocol:      a.APIProtocol,
		Actions:          a.Actions,
		LastUpdated:      a.LastUpdated,
		Extensions:       a.Extensions,
	}
}

func (a *APISpecInput) ToAPISpec() *APISpec {
	if a == nil {
		return nil
	}

	return &APISpec{
		Data:   a.Data,
		Format: a.Format,
		Type:   a.Type,
	}
}

func (a *APISpecInput) ToSpec() *SpecInput {
	if a == nil {
		return nil
	}

	return &SpecInput{
		Data:         a.Data,
		Format:       a.Format,
		Type:         SpecType(a.Type),
		FetchRequest: a.FetchRequest,
	}
}
