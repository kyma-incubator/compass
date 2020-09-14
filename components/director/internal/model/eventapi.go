package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"time"
)

type EventDefinition struct {
	ID               string
	OpenDiscoveryID  string
	Tenant           string
	BundleID         string
	Title            string
	ShortDescription string
	Description      *string
	Group            *string
	Specs            []*EventSpec
	Version          *Version
	EventDefinitions string
	Tags             *string
	Documentation    *string
	ChangelogEntries *string
	Logo             *string
	Image            *string
	URL              *string
	ReleaseStatus    string
	LastUpdated      time.Time
	Extensions       *string
}

type EventSpecType string

const (
	EventSpecTypeAsyncAPI EventSpecType = "ASYNC_API"
)

type EventSpec struct {
	ID     string
	Data   *string
	Type   EventSpecType
	Format SpecFormat
}

type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventDefinitionPage) IsPageable() {}

type EventDefinitionInput struct {
	ID               string
	OpenDiscoveryID  string
	Title            string
	ShortDescription string
	Description      *string
	Specs            []*EventSpecInput
	Group            *string
	Version          *VersionInput
	EventDefinitions string
	Tags             *string
	Documentation    *string
	ChangelogEntries *string
	Logo             *string
	Image            *string
	URL              *string
	ReleaseStatus    string
	LastUpdated      time.Time
	Extensions       *string
}

type EventSpecInput struct {
	Data          *string
	EventSpecType EventSpecType
	Format        SpecFormat
	FetchRequest  *FetchRequestInput
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(bundleID string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	specs := make([]*EventSpec, 0, 0)
	for _, spec := range e.Specs {
		specs = append(specs, spec.ToEventSpec())
	}

	return &EventDefinition{
		ID:               e.ID,
		OpenDiscoveryID:  e.OpenDiscoveryID,
		BundleID:         bundleID,
		Tenant:           tenant,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
		Group:            e.Group,
		Specs:            specs,
		Version:          e.Version.ToVersion(),
		EventDefinitions: e.EventDefinitions,
		Tags:             e.Tags,
		Documentation:    e.Documentation,
		ChangelogEntries: e.ChangelogEntries,
		Logo:             e.Logo,
		Image:            e.Image,
		ReleaseStatus:    e.ReleaseStatus,
		LastUpdated:      e.LastUpdated,
		Extensions:       e.Extensions,
	}
}

func (e *EventSpecInput) ToEventSpec() *EventSpec {
	if e == nil {
		return nil
	}

	return &EventSpec{
		Data:   e.Data,
		Type:   e.EventSpecType,
		Format: e.Format,
	}
}

func (e *EventSpecInput) ToSpec() *SpecInput {
	if e == nil {
		return nil
	}

	return &SpecInput{
		Data:         e.Data,
		Type:         SpecType(e.EventSpecType),
		Format:       e.Format,
		FetchRequest: e.FetchRequest,
	}
}
