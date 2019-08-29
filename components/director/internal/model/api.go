package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID            string
	ApplicationID string
	TenantID      string
	Name          string
	Description   *string
	Spec          *APISpec
	TargetURL     string
	//  group allows you to find the same API but in different version
	Group *string
	// Returns authentication details for all runtimes, even for a runtime, where Auth is not yet specified.
	Auths []*RuntimeAuth
	// If defaultAuth is specified, it will be used for all Runtimes that does not specify Auth explicitly.
	DefaultAuth *Auth
	Version     *Version
}

type APISpec struct {
	// when fetch request specified, data will be automatically populated
	Data           *string
	Format         SpecFormat
	Type           APISpecType
	FetchRequestID *string
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
	DefaultAuth *AuthInput
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

func (a *APIDefinitionInput) ToAPIDefinition(id string, appID string, fetchRequestID *string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          a.Name,
		Description:   a.Description,
		Spec:          a.Spec.ToAPISpec(fetchRequestID),
		TargetURL:     a.TargetURL,
		Group:         a.Group,
		Auths:         nil,
		DefaultAuth:   a.DefaultAuth.ToAuth(),
		Version:       a.Version.ToVersion(),
	}
}

func (a *APISpecInput) ToAPISpec(fetchRequestID *string) *APISpec {
	if a == nil {
		return nil
	}

	return &APISpec{
		Data:           a.Data,
		Format:         a.Format,
		Type:           a.Type,
		FetchRequestID: fetchRequestID,
	}
}
