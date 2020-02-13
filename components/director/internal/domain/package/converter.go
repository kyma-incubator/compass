package mp_package

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

type converter struct {
	auth AuthConverter
}

func NewConverter(auth AuthConverter) *converter {
	return &converter{
		auth: auth,
	}
}

func (c *converter) ToEntity(in *model.Package) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	defaultInstanceAuth, err := marshallDefaultInstanceAuth(in.DefaultInstanceAuth)
	if err != nil {
		return &Entity{}, err
	}

	output := &Entity{
		ID:                  in.ID,
		TenantID:            in.TenantID,
		ApplicationID:       in.ApplicationID,
		Name:                in.Name,
		Description:         repo.NewNullableString(in.Description),
		DefaultInstanceAuth: repo.NewNullableString(defaultInstanceAuth),
		EntityInstanceAuth:  EntityInstanceAuth{},
	}

	if in.InstanceAuthRequestInputSchema != nil {
		b, err := json.Marshal(in.InstanceAuthRequestInputSchema)
		if err != nil {
			return &Entity{}, errors.Wrap(err, "while marshaling schema to JSON")
		}
		output.InstanceAuthRequestJSONSchema = sql.NullString{String: string(b), Valid: true}
	} else {
		output.InstanceAuthRequestJSONSchema = sql.NullString{Valid: false}
	}

	return output, nil
}

func (c *converter) FromEntity(entity *Entity) (*model.Package, error) {
	if entity == nil {
		return nil, errors.New("the Package entity is nil")
	}

	defaultInstanceAuth, err := unmarshallDefaultInstanceAuth(entity.DefaultInstanceAuth)
	if err != nil {
		return nil, err
	}

	output := &model.Package{
		ID:                  entity.ID,
		TenantID:            entity.TenantID,
		ApplicationID:       entity.ApplicationID,
		Name:                entity.Name,
		Description:         &entity.Description.String,
		DefaultInstanceAuth: defaultInstanceAuth,
	}

	if entity.InstanceAuthRequestJSONSchema.Valid {
		mapDest := map[string]interface{}{}
		var tmp interface{}
		err := json.Unmarshal([]byte(entity.InstanceAuthRequestJSONSchema.String), &mapDest)
		if err != nil {
			return &model.Package{}, err
		}
		tmp = mapDest
		output.InstanceAuthRequestInputSchema = &tmp
	}
	return output, nil
}

func (c *converter) ToGraphQL(in *model.Package) (*graphql.Package, error) {
	if in == nil {
		return nil, errors.New("the model Package is nil")
	}

	schema, err := graphql.MarshalSchema(in.InstanceAuthRequestInputSchema)
	if err != nil {
		return &graphql.Package{}, err
	}

	return &graphql.Package{
		ID:                             in.ID,
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: schema,
		DefaultInstanceAuth:            c.auth.ToGraphQL(in.DefaultInstanceAuth),
	}, nil
}

func (c *converter) MultipleToGraphQL(in []*model.Package) ([]*graphql.Package, error) {
	var packages []*graphql.Package
	for _, r := range in {
		if r == nil {
			continue
		}
		pkg, err := c.ToGraphQL(r)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Package to GraphQL")
		}
		packages = append(packages, pkg)
	}

	return packages, nil
}

func (c *converter) CreateInputFromGraphQL(in graphql.PackageCreateInput) (model.PackageCreateInput, error) {
	schema, err := in.InstanceAuthRequestInputSchema.Unmarshal()
	if err != nil {
		return model.PackageCreateInput{}, err
	}

	return model.PackageCreateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: schema,
		DefaultInstanceAuth:            c.auth.InputFromGraphQL(in.DefaultInstanceAuth),
	}, nil
}

func (c *converter) UpdateInputFromGraphQL(in graphql.PackageUpdateInput) (model.PackageUpdateInput, error) {
	schema, err := in.InstanceAuthRequestInputSchema.Unmarshal()
	if err != nil {
		return model.PackageUpdateInput{}, err
	}
	return model.PackageUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: schema,
		DefaultInstanceAuth:            c.auth.InputFromGraphQL(in.DefaultInstanceAuth),
	}, nil
}

func marshallDefaultInstanceAuth(defaultInstanceAuth *model.Auth) (*string, error) {
	if defaultInstanceAuth == nil {
		return nil, nil
	}

	output, err := json.Marshal(defaultInstanceAuth)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling default auth")
	}
	return str.Ptr(string(output)), nil
}

func unmarshallDefaultInstanceAuth(defaultInstanceAuthSql sql.NullString) (*model.Auth, error) {
	var defaultInstanceAuth *model.Auth
	if defaultInstanceAuthSql.Valid && defaultInstanceAuthSql.String != "" {
		defaultInstanceAuth = &model.Auth{}
		err := json.Unmarshal([]byte(defaultInstanceAuthSql.String), defaultInstanceAuth)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshalling default instance auth")
		}
	}

	return defaultInstanceAuth, nil
}
