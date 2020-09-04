package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID               string
	BundleID         string
	Tenant           string
	Title            string
	ShortDescription string
	Description      *string
	Spec             *APISpec
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
	Title            string
	ShortDescription string
	Description      *string
	EntryPoint       string
	Group            *string
	Spec             *APISpecInput
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

func (a *APIDefinitionInput) ToAPIDefinitionWithinBundle(id string, bundleID string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		ID:               id,
		BundleID:         bundleID,
		Tenant:           tenant,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Spec:             a.Spec.ToAPISpec(),
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
