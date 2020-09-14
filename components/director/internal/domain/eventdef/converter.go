package eventdef

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"time"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

type converter struct {
	fr FetchRequestConverter
	vc VersionConverter
}

func NewConverter(fr FetchRequestConverter, vc VersionConverter) *converter {
	return &converter{fr: fr, vc: vc}
}

func (c *converter) ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition {
	if in == nil {
		return nil
	}

	return &graphql.EventDefinition{
		ID:               in.ID,
		OpenDiscoveryID:  &in.OpenDiscoveryID,
		BundleID:         in.BundleID,
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Group:            in.Group,
		Specs:            c.EventAPISpecToGraphQL(in.Specs),
		Version:          c.vc.ToGraphQL(in.Version),
		EventDefinitions: graphql.JSON(in.EventDefinitions),
		Documentation:    in.Documentation,
		ChangelogEntries: c.strPtrToJSONPtr(in.ChangelogEntries),
		Logo:             in.Logo,
		Image:            in.Image,
		URL:              in.URL,
		ReleaseStatus:    in.ReleaseStatus,
		Tags:             c.strPtrToJSONPtr(in.Tags),
		LastUpdated:      graphql.Timestamp(in.LastUpdated),
		Extensions:       c.strPtrToJSONPtr(in.Extensions),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition {
	var apis []*graphql.EventDefinition
	for _, a := range in {
		if a == nil {
			continue
		}
		apis = append(apis, c.ToGraphQL(a))
	}

	return apis
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) ([]*model.EventDefinitionInput, error) {
	var arr []*model.EventDefinitionInput
	for _, item := range in {
		api, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, err
		}

		arr = append(arr, api)
	}

	return arr, nil
}

func (c *converter) InputFromGraphQL(in *graphql.EventDefinitionInput) (*model.EventDefinitionInput, error) {
	if in == nil {
		return nil, nil
	}

	specs, err := c.eventAPISpecInputFromGraphQL(in.Specs)
	if err != nil {
		return nil, err
	}

	return &model.EventDefinitionInput{
		ID:               c.strPrtToStr(in.ID),
		OpenDiscoveryID:  c.strPrtToStr(in.OpenDiscoveryID),
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Specs:            specs,
		Group:            in.Group,
		Version:          c.vc.InputFromGraphQL(in.Version),
		EventDefinitions: string(in.EventDefinitions),
		Documentation:    in.Documentation,
		ChangelogEntries: c.jsonPtrToStrPtr(in.ChangelogEntries),
		Logo:             in.Logo,
		Image:            in.Image,
		URL:              in.URL,
		ReleaseStatus:    in.ReleaseStatus,
		Tags:             c.jsonPtrToStrPtr(in.Tags),
		LastUpdated:      time.Time(in.LastUpdated),
		Extensions:       c.jsonPtrToStrPtr(in.Extensions),
	}, nil
}

func (c *converter) EventAPISpecToGraphQL(ins []*model.EventSpec) []*graphql.EventSpec {
	if ins == nil {
		return nil
	}

	result := make([]*graphql.EventSpec, 0, 0)
	for _, in := range ins {
		var data *graphql.CLOB
		if in.Data != nil {
			tmp := graphql.CLOB(*in.Data)
			data = &tmp
		}

		result = append(result, &graphql.EventSpec{
			ID:         in.ID,
			Data:       data,
			Type:       graphql.EventSpecType(in.Type),
			CustomType: in.CustomType,
			Format:     graphql.SpecFormat(in.Format),
		})
	}
	return result
}

func (c *converter) eventAPISpecInputFromGraphQL(ins []*graphql.EventSpecInput) ([]*model.EventSpecInput, error) {
	if ins == nil {
		return nil, nil
	}
	result := make([]*model.EventSpecInput, 0, 0)

	for _, in := range ins {
		fetchReq, err := c.fr.InputFromGraphQL(in.FetchRequest)
		if err != nil {
			return nil, errors.Wrap(err, "while converting FetchRequest from GraphQL input")
		}

		result = append(result, &model.EventSpecInput{
			Data:          (*string)(in.Data),
			Format:        model.SpecFormat(in.Format),
			EventSpecType: model.EventSpecType(in.Type),
			CustomType:    in.CustomType,
			FetchRequest:  fetchReq,
		})
	}
	return result, nil
}

func (c *converter) FromEntity(entity Entity) (model.EventDefinition, error) {
	return model.EventDefinition{
		ID:               entity.ID,
		OpenDiscoveryID:  entity.OpenDiscoveryID,
		Tenant:           entity.TenantID,
		BundleID:         entity.BundleID,
		Title:            entity.Title,
		ShortDescription: entity.ShortDescription,
		Description:      repo.StringPtrFromNullableString(entity.Description),
		Group:            repo.StringPtrFromNullableString(entity.GroupName),
		Version:          c.vc.FromEntity(entity.Version),
		EventDefinitions: entity.EventDefinitions,
		Documentation:    repo.StringPtrFromNullableString(entity.Documentation),
		ChangelogEntries: repo.StringPtrFromNullableString(entity.ChangelogEntries),
		Logo:             repo.StringPtrFromNullableString(entity.Logo),
		Image:            repo.StringPtrFromNullableString(entity.Image),
		URL:              repo.StringPtrFromNullableString(entity.URL),
		ReleaseStatus:    entity.ReleaseStatus,
		Tags:             repo.StringPtrFromNullableString(entity.Tags),
		LastUpdated:      entity.LastUpdated,
		Extensions:       repo.StringPtrFromNullableString(entity.Extensions),
	}, nil
}

func (c *converter) ToEntity(eventModel model.EventDefinition) (Entity, error) {
	return Entity{
		ID:               eventModel.ID,
		OpenDiscoveryID:  eventModel.OpenDiscoveryID,
		TenantID:         eventModel.Tenant,
		BundleID:         eventModel.BundleID,
		Title:            eventModel.Title,
		ShortDescription: eventModel.ShortDescription,
		Description:      repo.NewNullableString(eventModel.Description),
		GroupName:        repo.NewNullableString(eventModel.Group),

		EventDefinitions: eventModel.EventDefinitions,
		Documentation:    repo.NewNullableString(eventModel.Documentation),
		ChangelogEntries: repo.NewNullableString(eventModel.ChangelogEntries),
		Logo:             repo.NewNullableString(eventModel.Logo),
		Image:            repo.NewNullableString(eventModel.Image),
		URL:              repo.NewNullableString(eventModel.URL),
		ReleaseStatus:    eventModel.ReleaseStatus,
		Tags:             repo.NewNullableString(eventModel.Tags),
		LastUpdated:      eventModel.LastUpdated,
		Extensions:       repo.NewNullableString(eventModel.Extensions),

		Version: c.convertVersionToEntity(eventModel.Version),
	}, nil
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.vc.ToEntity(*inVer)
}

func (c *converter) apiSpecToEntity(spec *model.EventSpec) EntitySpec {
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

func (c *converter) EventSpecFromEntity(specEnt EntitySpec) *model.EventSpec {
	if !specEnt.SpecType.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecData.Valid {
		return nil
	}

	apiSpec := model.EventSpec{}
	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		apiSpec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		apiSpec.Type = model.EventSpecType(*specType)
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
