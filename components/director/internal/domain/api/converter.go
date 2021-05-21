package api

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery --name=VersionConverter --output=automock --outpkg=automock --case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

//go:generate mockery --name=SpecConverter --output=automock --outpkg=automock --case=underscore
type SpecConverter interface {
	ToGraphQLAPISpec(in *model.Spec) (*graphql.APISpec, error)
	InputFromGraphQLAPISpec(in *graphql.APISpecInput) (*model.SpecInput, error)
}

type converter struct {
	version       VersionConverter
	specConverter SpecConverter
}

func NewConverter(version VersionConverter, specConverter SpecConverter) *converter {
	return &converter{version: version, specConverter: specConverter}
}

func (c *converter) ToGraphQL(in *model.APIDefinition, spec *model.Spec, bundleRef *model.BundleReference) (*graphql.APIDefinition, error) {
	if in == nil {
		return nil, nil
	}

	s, err := c.specConverter.ToGraphQLAPISpec(spec)
	if err != nil {
		return nil, err
	}

	var bundleID string
	if bundleRef.BundleID != nil {
		bundleID = *bundleRef.BundleID
	}

	return &graphql.APIDefinition{
		BundleID:    bundleID,
		Name:        in.Name,
		Description: in.Description,
		Spec:        s,
		TargetURL:   str.PtrStrToStr(bundleRef.APIDefaultTargetURL),
		Group:       in.Group,
		Version:     c.version.ToGraphQL(in.Version),
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: timePtrToTimestampPtr(in.CreatedAt),
			UpdatedAt: timePtrToTimestampPtr(in.UpdatedAt),
			DeletedAt: timePtrToTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

func (c *converter) MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec, bundleRefs []*model.BundleReference) ([]*graphql.APIDefinition, error) {
	if len(in) != len(specs) || len(in) != len(bundleRefs) || len(bundleRefs) != len(specs) {
		return nil, errors.New("different apis, specs and bundleRefs count provided")
	}

	var apis []*graphql.APIDefinition
	for i, a := range in {
		if a == nil {
			continue
		}

		api, err := c.ToGraphQL(a, specs[i], bundleRefs[i])
		if err != nil {
			return nil, err
		}

		apis = append(apis, api)
	}

	return apis, nil
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error) {
	var apiDefs []*model.APIDefinitionInput
	var specs []*model.SpecInput

	for _, item := range in {
		api, spec, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, nil, err
		}

		apiDefs = append(apiDefs, api)
		specs = append(specs, spec)
	}

	return apiDefs, specs, nil
}

func (c *converter) InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, *model.SpecInput, error) {
	if in == nil {
		return nil, nil, nil
	}

	spec, err := c.specConverter.InputFromGraphQLAPISpec(in.Spec)
	if err != nil {
		return nil, nil, err
	}

	return &model.APIDefinitionInput{
		Name:         in.Name,
		Description:  in.Description,
		TargetURLs:   ConvertTargetUrlToJsonArray(in.TargetURL),
		Group:        in.Group,
		VersionInput: c.version.InputFromGraphQL(in.Version),
	}, spec, nil
}

func (c *converter) FromEntity(entity Entity) model.APIDefinition {

	return model.APIDefinition{
		ApplicationID:                           entity.ApplicationID,
		PackageID:                               repo.StringPtrFromNullableString(entity.PackageID),
		Tenant:                                  entity.TenantID,
		Name:                                    entity.Name,
		Description:                             repo.StringPtrFromNullableString(entity.Description),
		TargetURLs:                              repo.JSONRawMessageFromNullableString(entity.TargetURLs),
		Group:                                   repo.StringPtrFromNullableString(entity.Group),
		OrdID:                                   repo.StringPtrFromNullableString(entity.OrdID),
		ShortDescription:                        repo.StringPtrFromNullableString(entity.ShortDescription),
		SystemInstanceAware:                     repo.BoolPtrFromNullableBool(entity.SystemInstanceAware),
		ApiProtocol:                             repo.StringPtrFromNullableString(entity.ApiProtocol),
		Tags:                                    repo.JSONRawMessageFromNullableString(entity.Tags),
		Countries:                               repo.JSONRawMessageFromNullableString(entity.Countries),
		Links:                                   repo.JSONRawMessageFromNullableString(entity.Links),
		APIResourceLinks:                        repo.JSONRawMessageFromNullableString(entity.APIResourceLinks),
		ReleaseStatus:                           repo.StringPtrFromNullableString(entity.ReleaseStatus),
		SunsetDate:                              repo.StringPtrFromNullableString(entity.SunsetDate),
		Successor:                               repo.StringPtrFromNullableString(entity.Successor),
		ChangeLogEntries:                        repo.JSONRawMessageFromNullableString(entity.ChangeLogEntries),
		Labels:                                  repo.JSONRawMessageFromNullableString(entity.Labels),
		Visibility:                              repo.StringPtrFromNullableString(entity.Visibility),
		Disabled:                                repo.BoolPtrFromNullableBool(entity.Disabled),
		PartOfProducts:                          repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LineOfBusiness:                          repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Industry:                                repo.JSONRawMessageFromNullableString(entity.Industry),
		ImplementationStandard:                  repo.StringPtrFromNullableString(entity.ImplementationStandard),
		CustomImplementationStandard:            repo.StringPtrFromNullableString(entity.CustomImplementationStandard),
		CustomImplementationStandardDescription: repo.StringPtrFromNullableString(entity.CustomImplementationStandardDescription),
		Version:                                 c.version.FromEntity(entity.Version),
		Extensible:                              repo.JSONRawMessageFromNullableString(entity.Extensible),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}
}

func (c *converter) ToEntity(apiModel model.APIDefinition) *Entity {
	return &Entity{
		TenantID:                                apiModel.Tenant,
		ApplicationID:                           apiModel.ApplicationID,
		PackageID:                               repo.NewNullableString(apiModel.PackageID),
		Name:                                    apiModel.Name,
		Description:                             repo.NewNullableString(apiModel.Description),
		Group:                                   repo.NewNullableString(apiModel.Group),
		TargetURLs:                              repo.NewNullableStringFromJSONRawMessage(apiModel.TargetURLs),
		OrdID:                                   repo.NewNullableString(apiModel.OrdID),
		ShortDescription:                        repo.NewNullableString(apiModel.ShortDescription),
		SystemInstanceAware:                     repo.NewNullableBool(apiModel.SystemInstanceAware),
		ApiProtocol:                             repo.NewNullableString(apiModel.ApiProtocol),
		Tags:                                    repo.NewNullableStringFromJSONRawMessage(apiModel.Tags),
		Countries:                               repo.NewNullableStringFromJSONRawMessage(apiModel.Countries),
		Links:                                   repo.NewNullableStringFromJSONRawMessage(apiModel.Links),
		APIResourceLinks:                        repo.NewNullableStringFromJSONRawMessage(apiModel.APIResourceLinks),
		ReleaseStatus:                           repo.NewNullableString(apiModel.ReleaseStatus),
		SunsetDate:                              repo.NewNullableString(apiModel.SunsetDate),
		Successor:                               repo.NewNullableString(apiModel.Successor),
		ChangeLogEntries:                        repo.NewNullableStringFromJSONRawMessage(apiModel.ChangeLogEntries),
		Labels:                                  repo.NewNullableStringFromJSONRawMessage(apiModel.Labels),
		Visibility:                              repo.NewNullableString(apiModel.Visibility),
		Disabled:                                repo.NewNullableBool(apiModel.Disabled),
		PartOfProducts:                          repo.NewNullableStringFromJSONRawMessage(apiModel.PartOfProducts),
		LineOfBusiness:                          repo.NewNullableStringFromJSONRawMessage(apiModel.LineOfBusiness),
		Industry:                                repo.NewNullableStringFromJSONRawMessage(apiModel.Industry),
		ImplementationStandard:                  repo.NewNullableString(apiModel.ImplementationStandard),
		CustomImplementationStandard:            repo.NewNullableString(apiModel.CustomImplementationStandard),
		CustomImplementationStandardDescription: repo.NewNullableString(apiModel.CustomImplementationStandardDescription),
		Version:                                 c.convertVersionToEntity(apiModel.Version),
		Extensible:                              repo.NewNullableStringFromJSONRawMessage(apiModel.Extensible),
		BaseEntity: &repo.BaseEntity{
			ID:        apiModel.ID,
			Ready:     apiModel.Ready,
			CreatedAt: apiModel.CreatedAt,
			UpdatedAt: apiModel.UpdatedAt,
			DeletedAt: apiModel.DeletedAt,
			Error:     repo.NewNullableString(apiModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}

func ExtractTargetUrlFromJsonArray(jsonTargetUrl json.RawMessage) string {
	strTargetUrl := string(jsonTargetUrl)
	strTargetUrl = strings.TrimPrefix(strTargetUrl, `["`)
	strTargetUrl = strings.TrimSuffix(strTargetUrl, `"]`)

	return strTargetUrl
}

func ConvertTargetUrlToJsonArray(targetUrl string) json.RawMessage {
	if targetUrl == "" {
		return nil
	}

	return json.RawMessage(`["` + targetUrl + `"]`)
}
