package model

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/go-ozzo/ozzo-validation/is"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ID            string
	ApplicationID string
	Tenant        string
	Name          string
	Description   *string
	Spec          *APISpec
	TargetURL     string
	//  group allows you to find the same API but in different version
	Group *string
	// Returns authentication details for all runtimes, even for a runtime, where Auth is not yet specified.
	Auths []*APIRuntimeAuth
	// If defaultAuth is specified, it will be used for all Runtimes that does not specify Auth explicitly.
	DefaultAuth *Auth
	Version     *Version
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
	DefaultAuth *AuthInput
}

func (i *APIDefinitionInput) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.Name, validation.Required, validation.By(inputvalidation.ValidateName)),
		validation.Field(&i.Description, validation.NilOrNotEmpty, validation.Length(1, 128), validation.By(inputvalidation.ValidatePrintableWithWhitespace)),
		validation.Field(&i.TargetURL, validation.Required, is.URL, validation.Length(1, 256), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Group, validation.NilOrNotEmpty, validation.Length(1, 36), validation.By(inputvalidation.ValidatePrintable)),
		validation.Field(&i.Spec),
		validation.Field(&i.Version),
	)
}

type APISpecInput struct {
	Data         *string
	Type         APISpecType
	Format       SpecFormat
	FetchRequest *FetchRequestInput
}

func (i *APISpecInput) Validate() error {
	return validation.Errors{
		"MatchingTypeAndFormat": i.validateMatchingSpecAndType(),
		"Data":                  validation.Validate(i.Data, validation.NilOrNotEmpty, validation.By(inputvalidation.ValidatePrintableWithWhitespace)),
		"Type":                  validation.Validate(i.Type, validation.Required, validation.In(APISpecTypeOdata, APISpecTypeOpenAPI), validation.By(inputvalidation.ValidatePrintable)),
		"Format":                validation.Validate(i.Format, validation.Required, validation.In(SpecFormatYaml, SpecFormatJSON, SpecFormatXML), validation.By(inputvalidation.ValidatePrintable)),
		"FetchRequest":          validation.Validate(i.FetchRequest),
	}.Filter()
}

func (i *APISpecInput) validateMatchingSpecAndType() error {
	switch i.Type {
	case APISpecTypeOdata:
		if !i.formatIsOneOf(i.Format, []SpecFormat{SpecFormatXML, SpecFormatJSON}) {
			return errors.Errorf("format %s is not a valid spec format for spec type %s", i.Format, i.Type)
		}
	case APISpecTypeOpenAPI:
		if !i.formatIsOneOf(i.Format, []SpecFormat{SpecFormatJSON, SpecFormatYaml}) {
			return errors.Errorf("format %s is not a valid spec format for spec type %s", i.Format, i.Type)
		}
	default:
		return errors.New("invalid spec type")
	}
	return nil
}

func (i *APISpecInput) formatIsOneOf(format SpecFormat, formats []SpecFormat) bool {
	var slice []string
	for _, value := range formats {
		slice = append(slice, (string)(value))
	}
	return str.IsInSlice((string)(i.Format), slice)
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinition(id string, appID string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		ID:            id,
		ApplicationID: appID,
		Tenant:        tenant,
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
		Data:   a.Data,
		Format: a.Format,
		Type:   a.Type,
	}
}
