package api

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=AuthConverter -output=automock -outpkg=automock -case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) *graphql.Auth
	InputFromGraphQL(in *graphql.AuthInput) *model.AuthInput
}

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) (*model.Version, error)
	ToEntity(version model.Version) (version.Version, error)
}

type converter struct {
	auth    AuthConverter
	fr      FetchRequestConverter
	version VersionConverter
}

func NewConverter(auth AuthConverter, fr FetchRequestConverter, version VersionConverter) *converter {
	return &converter{auth: auth, fr: fr, version: version}
}

func (c *converter) ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition {
	if in == nil {
		return nil
	}

	return &graphql.APIDefinition{
		ID:            in.ID,
		ApplicationID: in.ApplicationID,
		Name:          in.Name,
		Description:   in.Description,
		Spec:          c.apiSpecToGraphQL(in.ID, in.Spec),
		TargetURL:     in.TargetURL,
		Group:         in.Group,
		DefaultAuth:   c.auth.ToGraphQL(in.DefaultAuth),
		Version:       c.version.ToGraphQL(in.Version),
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

func (c *converter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput {
	var arr []*model.APIDefinitionInput
	for _, item := range in {
		api := c.InputFromGraphQL(item)
		arr = append(arr, api)
	}

	return arr
}

func (c *converter) InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput {
	if in == nil {
		return nil
	}

	return &model.APIDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		TargetURL:   in.TargetURL,
		Group:       in.Group,
		Spec:        c.apiSpecInputFromGraphQL(in.Spec),
		Version:     c.version.InputFromGraphQL(in.Version),
		DefaultAuth: c.auth.InputFromGraphQL(in.DefaultAuth),
	}
}

func (c *converter) apiSpecToGraphQL(definitionID string, in *model.APISpec) *graphql.APISpec {
	if in == nil {
		return nil
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.APISpec{
		Data:         data,
		Type:         graphql.APISpecType(in.Type),
		Format:       graphql.SpecFormat(in.Format),
		DefinitionID: definitionID,
	}
}

func (c *converter) apiSpecInputFromGraphQL(in *graphql.APISpecInput) *model.APISpecInput {
	if in == nil {
		return nil
	}

	return &model.APISpecInput{
		Data:         (*string)(in.Data),
		Type:         model.APISpecType(in.Type),
		Format:       model.SpecFormat(in.Format),
		FetchRequest: c.fr.InputFromGraphQL(in.FetchRequest),
	}
}

func (c *converter) FromEntity(entity Entity) (model.APIDefinition, error) {
	defaultAuth, err := unmarshallDefaultAuth(entity.DefaultAuth)
	if err != nil {
		return model.APIDefinition{}, errors.Wrap(err, "while converting ApiDefinition")
	}

	v, err := c.version.FromEntity(entity.Version)
	if err != nil {
		return model.APIDefinition{}, err
	}

	return model.APIDefinition{
		ID:            entity.ID,
		ApplicationID: entity.AppID,
		Name:          entity.Name,
		TargetURL:     entity.TargetURL,
		Tenant:        entity.TenantID,
		DefaultAuth:   defaultAuth,
		Description:   repo.StringPtrFromNullableString(entity.Description),
		Group:         repo.StringPtrFromNullableString(entity.Group),
		Spec:          c.apiSpecFromEntity(entity.EntitySpec),
		Version:       v,
	}, nil
}

func (c *converter) ToEntity(apiModel model.APIDefinition) (Entity, error) {
	defaultAuth, err := marshallDefaultAuth(apiModel.DefaultAuth)
	if err != nil {
		return Entity{}, errors.Wrap(err, "while converting ApiDefinition")
	}

	versionEntity, err := c.convertVersionToEntity(apiModel.Version)
	if err != nil {
		return Entity{}, err
	}

	return Entity{
		ID:          apiModel.ID,
		TenantID:    apiModel.Tenant,
		AppID:       apiModel.ApplicationID,
		Name:        apiModel.Name,
		Description: repo.NewNullableString(apiModel.Description),
		Group:       repo.NewNullableString(apiModel.Group),
		TargetURL:   apiModel.TargetURL,

		EntitySpec:  c.apiSpecToEntity(apiModel.Spec),
		DefaultAuth: repo.NewNullableString(defaultAuth),
		Version:     versionEntity,
	}, nil
}

func (c *converter) convertVersionToEntity(inVer *model.Version) (version.Version, error) {
	if inVer == nil {
		return version.Version{}, nil
	}

	tmp, err := c.version.ToEntity(*inVer)
	if err != nil {
		return version.Version{}, errors.Wrap(err, "while converting version")
	}
	return tmp, nil
}

func (c *converter) apiSpecToEntity(spec *model.APISpec) EntitySpec {
	var apiSpecEnt EntitySpec
	if spec != nil {
		apiSpecEnt = EntitySpec{
			SpecFormat: repo.NewNullableString(strings.Ptr(string(spec.Format))),
			SpecType:   repo.NewNullableString(strings.Ptr(string(spec.Type))),
			SpecData:   repo.NewNullableString(spec.Data),
		}
	}

	return apiSpecEnt
}

func (c *converter) apiSpecFromEntity(specEnt EntitySpec) *model.APISpec {
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

func unmarshallDefaultAuth(defaultAuthSql sql.NullString) (*model.Auth, error) {
	var defaultAuth *model.Auth
	if defaultAuthSql.Valid && defaultAuthSql.String != "" {
		defaultAuth = &model.Auth{}
		err := json.Unmarshal([]byte(defaultAuthSql.String), defaultAuth)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshalling default auth")
		}
	}

	return defaultAuth, nil
}

func marshallDefaultAuth(defaultAuth *model.Auth) (*string, error) {
	if defaultAuth == nil {
		return nil, nil
	}

	output, err := json.Marshal(defaultAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling default auth")
	}
	return strings.Ptr(string(output)), nil
}
