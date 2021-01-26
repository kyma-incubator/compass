package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID          string
	BundleID    string
	Tenant      string
	Name        string
	Description *string
	Spec        *APISpec
	TargetURL   string
	//  group allows you to find the same API but in different version
	Group   *string
	Version *Version
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
	Name        string
	Description *string
	TargetURL   string
	Group       *string
	Spec        *APISpecInput
	Version     *VersionInput
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
		ID:          id,
		BundleID:    bundleID,
		Tenant:      tenant,
		Name:        a.Name,
		Description: a.Description,
		Spec:        a.Spec.ToAPISpec(),
		TargetURL:   a.TargetURL,
		Group:       a.Group,
		Version:     a.Version.ToVersion(),
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
