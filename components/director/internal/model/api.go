package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID            string
	ApplicationID string
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
	Data         *[]byte
	Format       *SpecFormat
	Type         APISpecType
	FetchRequest *FetchRequest
}

type APISpecType string

const (
	APISpecTypeOdata   APISpecType = "ODATA"
	APISpecTypeOpenAPI APISpecType = "OPEN_API"
)

type SpecFormat string

const (
	SpecFormatYaml SpecFormat = "YAML"
	SpecFormatJSON SpecFormat = "JSON"
)

type Timestamp time.Time

type Version struct {
	// for example 4.6
	Value      string
	Deprecated *bool
	// for example 4.5
	DeprecatedSince *string
	// if true, will be removed in the next version
	ForRemoval *bool
}

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
	Data         *[]byte
	Type         APISpecType
	Format       *SpecFormat
	FetchRequest *FetchRequestInput
}

type VersionInput struct {
	Value           string
	Deprecated      *bool
	DeprecatedSince *string
	ForRemoval      *bool
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinition(id string, appID string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          a.Name,
		Description:   a.Description,
		Spec:          a.Spec.ToAPISpec(),
		TargetURL:     a.TargetURL,
		Group:         a.Group,
		Auths:         nil,
		DefaultAuth:   a.DefaultAuth.ToAuth(),
		Version:       a.Version.ToVersion(),
	}
}

func (a *APISpecInput) ToAPISpec() *APISpec {
	if a == nil {
		return nil
	}

	return &APISpec{
		Data:         a.Data,
		Format:       a.Format,
		Type:         a.Type,
		FetchRequest: a.FetchRequest.ToFetchRequest(time.Now()),
	}
}

func (v *VersionInput) ToVersion() *Version {
	if v == nil {
		return nil
	}

	return &Version{
		Value:           v.Value,
		Deprecated:      v.Deprecated,
		DeprecatedSince: v.DeprecatedSince,
		ForRemoval:      v.ForRemoval,
	}
}
