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
	Spec             *EventSpec
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
	Spec             *EventSpecInput
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

	return &EventDefinition{
		ID:               e.ID,
		OpenDiscoveryID:  e.OpenDiscoveryID,
		BundleID:         bundleID,
		Tenant:           tenant,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
		Group:            e.Group,
		Spec:             e.Spec.ToEventSpec(),
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
