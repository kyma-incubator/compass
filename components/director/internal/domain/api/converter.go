package api

import (
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

type converter struct {
	fr      FetchRequestConverter
	version VersionConverter
}

func NewConverter(fr FetchRequestConverter, version VersionConverter) *converter {
	return &converter{fr: fr, version: version}
}

func (c *converter) ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition {
	if in == nil {
		return nil
	}

	return &graphql.APIDefinition{
		ID:               in.ID,
		OpenDiscoveryID:  &in.OpenDiscoveryID,
		BundleID:         in.BundleID,
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Specs:            c.SpecsToGraphQL(in.Specs),
		EntryPoint:       in.EntryPoint,
		Group:            in.Group,
		Version:          c.version.ToGraphQL(in.Version),
		APIDefinitions:   graphql.JSON(in.APIDefinitions),
		Documentation:    in.Documentation,
		ChangelogEntries: c.strPtrToJSONPtr(in.ChangelogEntries),
		Logo:             in.Logo,
		Image:            in.Image,
		URL:              in.URL,
		ReleaseStatus:    in.ReleaseStatus,
		APIProtocol:      in.APIProtocol,
		Actions:          graphql.JSON(in.Actions),
		Tags:             c.strPtrToJSONPtr(in.Tags),
		LastUpdated:      graphql.Timestamp(in.LastUpdated),
		Extensions:       c.strPtrToJSONPtr(in.Extensions),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition {
	var apis []*graphql.APIDefinition
	for _, a := range in {
		if a == nil {
			continue
		}
		apis = append(apis, c.ToGraphQL(a))
	}

	return apis
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, error) {
	var arr []*model.APIDefinitionInput
	for _, item := range in {
		api, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, err
		}

		arr = append(arr, api)
	}

	return arr, nil
}

func (c *converter) InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, error) {
	if in == nil {
		return nil, nil
	}

	specs, err := c.apiSpecsInputFromGraphQL(in.Specs)
	if err != nil {
		return nil, err
	}

	return &model.APIDefinitionInput{
		ID:               c.strPrtToStr(in.ID),
		OpenDiscoveryID:  c.strPrtToStr(in.OpenDiscoveryID),
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		EntryPoint:       in.EntryPoint,
		Group:            in.Group,
		Specs:            specs,
		Version:          c.version.InputFromGraphQL(in.Version),
		APIDefinitions:   string(in.APIDefinitions),
		Documentation:    in.Documentation,
		ChangelogEntries: c.jsonPtrToStrPtr(in.ChangelogEntries),
		Logo:             in.Logo,
		Image:            in.Image,
		URL:              in.URL,
		ReleaseStatus:    in.ReleaseStatus,
		APIProtocol:      in.APIProtocol,
		Actions:          string(in.Actions),
		Tags:             c.jsonPtrToStrPtr(in.Tags),
		LastUpdated:      time.Time(in.LastUpdated),
		Extensions:       c.jsonPtrToStrPtr(in.Extensions),
	}, nil
}

func (c *converter) SpecsToGraphQL(ins []*model.APISpec) []*graphql.APISpec {
	if ins == nil {
		return nil
	}
	result := make([]*graphql.APISpec, 0, 0)

	for _, in := range ins {
		var data *graphql.CLOB
		if in.Data != nil {
			tmp := graphql.CLOB(*in.Data)
			data = &tmp
		}

		result = append(result, &graphql.APISpec{
			Data:   data,
			Type:   graphql.APISpecType(in.Type),
			Format: graphql.SpecFormat(in.Format),
			ID:     in.ID,
		})
	}
	return result
}

func (c *converter) apiSpecsInputFromGraphQL(ins []*graphql.APISpecInput) ([]*model.APISpecInput, error) {
	if ins == nil {
		return nil, nil
	}
	result := make([]*model.APISpecInput, 0, 0)

	for _, in := range ins {
		fetchReq, err := c.fr.InputFromGraphQL(in.FetchRequest)
		if err != nil {
			return nil, errors.Wrap(err, "while converting FetchRequest from GraphQL input")
		}
		result = append(result, &model.APISpecInput{
			Data:         (*string)(in.Data),
			Type:         model.APISpecType(in.Type),
			Format:       model.SpecFormat(in.Format),
			FetchRequest: fetchReq,
		})
	}
	return result, nil
}

func (c *converter) FromEntity(entity Entity) model.APIDefinition {

	return model.APIDefinition{
		ID:               entity.ID,
		OpenDiscoveryID:  entity.OpenDiscoveryID,
		BundleID:         entity.BundleID,
		Title:            entity.Title,
		EntryPoint:       entity.EntryPoint,
		Tenant:           entity.TenantID,
		ShortDescription: entity.ShortDescription,
		Description:      repo.StringPtrFromNullableString(entity.Description),
		Group:            repo.StringPtrFromNullableString(entity.Group),
		Version:          c.version.FromEntity(entity.Version),
		APIDefinitions:   entity.APIDefinitions,
		Documentation:    repo.StringPtrFromNullableString(entity.Documentation),
		ChangelogEntries: repo.StringPtrFromNullableString(entity.ChangelogEntries),
		Logo:             repo.StringPtrFromNullableString(entity.Logo),
		Image:            repo.StringPtrFromNullableString(entity.Image),
		URL:              repo.StringPtrFromNullableString(entity.URL),
		ReleaseStatus:    entity.ReleaseStatus,
		APIProtocol:      entity.APIProtocol,
		Actions:          entity.Actions,
		Tags:             repo.StringPtrFromNullableString(entity.Tags),
		LastUpdated:      entity.LastUpdated,
		Extensions:       repo.StringPtrFromNullableString(entity.Extensions),
	}
}

func (c *converter) ToEntity(apiModel model.APIDefinition) Entity {

	return Entity{
		ID:               apiModel.ID,
		OpenDiscoveryID:  apiModel.OpenDiscoveryID,
		TenantID:         apiModel.Tenant,
		BundleID:         apiModel.BundleID,
		Title:            apiModel.Title,
		ShortDescription: apiModel.ShortDescription,
		Description:      repo.NewNullableString(apiModel.Description),
		Group:            repo.NewNullableString(apiModel.Group),
		EntryPoint:       apiModel.EntryPoint,

		APIDefinitions:   apiModel.APIDefinitions,
		Documentation:    repo.NewNullableString(apiModel.Documentation),
		ChangelogEntries: repo.NewNullableString(apiModel.ChangelogEntries),
		Logo:             repo.NewNullableString(apiModel.Logo),
		Image:            repo.NewNullableString(apiModel.Image),
		URL:              repo.NewNullableString(apiModel.URL),
		ReleaseStatus:    apiModel.ReleaseStatus,
		APIProtocol:      apiModel.APIProtocol,
		Actions:          apiModel.Actions,
		Tags:             repo.NewNullableString(apiModel.Tags),
		LastUpdated:      apiModel.LastUpdated,
		Extensions:       repo.NewNullableString(apiModel.Extensions),

		Version: c.convertVersionToEntity(apiModel.Version),
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}

func (c *converter) ApiSpecToEntity(spec *model.APISpec) EntitySpec {
	var apiSpecEnt EntitySpec
	if spec != nil {
		apiSpecEnt = EntitySpec{
			SpecFormat: repo.NewNullableString(str.Ptr(string(spec.Format))),
			SpecType:   repo.NewNullableString(str.Ptr(string(spec.Type))),
			SpecData:   repo.NewNullableString(spec.Data),
		}
	}

	return apiSpecEnt
}

func (c *converter) ApiSpecFromEntity(specEnt EntitySpec) *model.APISpec {
	if !specEnt.SpecData.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecType.Valid {
		return nil
	}

	apiSpec := model.APISpec{}
	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		apiSpec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		apiSpec.Type = model.APISpecType(*specType)
	}
	apiSpec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)
	return &apiSpec
}

func (c *converter) strPtrToJSONPtr(in *string) *graphql.JSON {
	if in == nil {
		return nil
	}
	out := graphql.JSON(*in)
	return &out
}

func (c *converter) jsonPtrToStrPtr(in *graphql.JSON) *string {
	if in == nil {
		return nil
	}
	out := string(*in)
	return &out
}

func (c *converter) strPrtToStr(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}
