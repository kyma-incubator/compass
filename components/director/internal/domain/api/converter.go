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
	FromEntity(version version.Version) (model.Version, error)
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
		Auths:         c.runtimeAuthArrToGraphQL(in.Auths),
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

func (c *converter) runtimeAuthToGraphQL(in *model.RuntimeAuth) *graphql.RuntimeAuth {
	if in == nil {
		return nil
	}

	return &graphql.RuntimeAuth{
		RuntimeID: in.RuntimeID,
		Auth:      c.auth.ToGraphQL(in.Auth),
	}
}

func (c *converter) runtimeAuthArrToGraphQL(in []*model.RuntimeAuth) []*graphql.RuntimeAuth {
	var auths []*graphql.RuntimeAuth
	for _, item := range in {
		auths = append(auths, &graphql.RuntimeAuth{
			RuntimeID: item.RuntimeID,
			Auth:      c.auth.ToGraphQL(item.Auth),
		})
	}

	return auths
}

func (c *converter) FromEntity(entity APIDefinition) (model.APIDefinition, error) {
	defaultAuth, err := unmarshallDefaultAuth(entity.DefaultAuth)
	if err != nil {
		return model.APIDefinition{}, errors.Wrap(err, "while converting ApiDefinition")
	}

	var vModel *model.Version
	if entity.Version != nil {
		v, err := c.version.FromEntity(*entity.Version)
		if err != nil {
			return model.APIDefinition{}, err
		}
		vModel = &v
	}

	return model.APIDefinition{
		ID:            entity.ID,
		ApplicationID: entity.AppID,
		Name:          entity.Name,
		TargetURL:     entity.TargetURL,
		TenantID:      entity.TenantID,
		DefaultAuth:   defaultAuth,
		Description:   repo.StringPtrFromNullableString(entity.Description),
		Group:         repo.StringPtrFromNullableString(entity.Group),
		//TODO: add spec_fetch_request_ID when resolver will be implemented
		Spec:    c.apiSpecFromEntity(entity.APISpec),
		Version: vModel,
	}, nil
}

func (c *converter) ToEntity(apiModel model.APIDefinition) (APIDefinition, error) {
	defaultAuth, err := marshallDefaultAuth(apiModel.DefaultAuth)
	if err != nil {
		return APIDefinition{}, errors.Wrap(err, "while converting ApiDefinition")
	}

	var versionEntity *version.Version
	if apiModel.Version != nil {
		tmp, err := c.version.ToEntity(*apiModel.Version)
		if err != nil {
			return APIDefinition{}, errors.Wrap(err, "while converting version")
		}
		versionEntity = &tmp
	}

	return APIDefinition{
		ID:          apiModel.ID,
		TenantID:    apiModel.TenantID,
		AppID:       apiModel.ApplicationID,
		Name:        apiModel.Name,
		Description: repo.NewNullableString(apiModel.Description),
		Group:       repo.NewNullableString(apiModel.Group),
		TargetURL:   apiModel.TargetURL,
		APISpec:     c.apiSpecToEntity(apiModel.Spec),
		DefaultAuth: repo.NewNullableString(&defaultAuth),
		Version:     versionEntity,
		//TODO: add spec_fetch_request_ID when resolver will be implemented
	}, nil
}

func (c *converter) apiSpecToEntity(spec *model.APISpec) *APISpec {
	var apiSpecEnt *APISpec
	if spec != nil {
		tmp := APISpec{
			SpecFormat: repo.NewNullableString(strings.Ptr(string(spec.Format))),
			SpecType:   repo.NewNullableString(strings.Ptr(string(spec.Type))),
			SpecData:   repo.NewNullableString(spec.Data),
		}
		apiSpecEnt = &tmp
	}

	return apiSpecEnt
}

func (c *converter) apiSpecFromEntity(specEnt *APISpec) *model.APISpec {
	var apiSpec *model.APISpec

	if specEnt != nil {
		tmp := model.APISpec{}
		specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
		if specFormat != nil {
			tmp.Format = model.SpecFormat(*specFormat)
		}

		specType := repo.StringPtrFromNullableString(specEnt.SpecType)
		if specFormat != nil {
			tmp.Type = model.APISpecType(*specType)
		}
		tmp.Data = repo.StringPtrFromNullableString(specEnt.SpecData)
		apiSpec = &tmp
	}

	return apiSpec
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

func marshallDefaultAuth(defaultAuth *model.Auth) (string, error) {
	marshaledAuth := ""
	if defaultAuth != nil {
		output, err := json.Marshal(defaultAuth)
		if err != nil {
			return "", errors.Wrap(err, "while marshaling default auth")
		}
		marshaledAuth = string(output)
	}
	return marshaledAuth, nil
}
