package mp_package

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
	bundle BundleConverter
}

func NewConverter(bundle BundleConverter) *converter {
	return &converter{
		bundle: bundle,
	}
}

func (c *converter) ToEntity(in *model.Package) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	output := &Entity{
		ID:               in.ID,
		TenantID:         in.TenantID,
		ApplicationID:    in.ApplicationID,
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Version:          in.Version,
		Licence:          repo.NewNullableString(in.Licence),
		LicenceType:      repo.NewNullableString(in.LicenceType),
		TermsOfService:   repo.NewNullableString(in.TermsOfService),
		Logo:             repo.NewNullableString(in.Logo),
		Image:            repo.NewNullableString(in.Image),
		Provider:         repo.NewNullableString(in.Provider),
		Actions:          repo.NewNullableString(in.Actions),
		Tags:             repo.NewNullableString(in.Tags),
		LastUpdated:      in.LastUpdated,
		Extensions:       repo.NewNullableString(in.Extensions),
	}

	return output, nil
}

func (c *converter) FromEntity(entity *Entity) (*model.Package, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Bundle entity is nil")
	}

	output := &model.Package{
		ID:               entity.ID,
		TenantID:         entity.TenantID,
		ApplicationID:    entity.ApplicationID,
		Title:            entity.Title,
		ShortDescription: entity.ShortDescription,
		Description:      entity.Description,
		Version:          entity.Version,
		Licence:          repo.StringPtrFromNullableString(entity.Licence),
		LicenceType:      repo.StringPtrFromNullableString(entity.LicenceType),
		TermsOfService:   repo.StringPtrFromNullableString(entity.TermsOfService),
		Logo:             repo.StringPtrFromNullableString(entity.Logo),
		Image:            repo.StringPtrFromNullableString(entity.Image),
		Provider:         repo.StringPtrFromNullableString(entity.Provider),
		Actions:          repo.StringPtrFromNullableString(entity.Actions),
		Tags:             repo.StringPtrFromNullableString(entity.Tags),
		LastUpdated:      entity.LastUpdated,
		Extensions:       repo.StringPtrFromNullableString(entity.Extensions),
	}

	return output, nil
}

func (c *converter) ToGraphQL(in *model.Package) (*graphql.Package, error) {
	if in == nil {
		return nil, apperrors.NewInternalError("the model Bundle is nil")
	}

	return &graphql.Package{
		ID:               in.ID,
		ApplicationID:    in.ApplicationID,
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Version:          in.Version,
		Licence:          in.Licence,
		LicenceType:      in.LicenceType,
		TermsOfService:   in.TermsOfService,
		Logo:             in.Logo,
		Image:            in.Image,
		Provider:         c.strPtrToJSONPtr(in.Provider),
		Actions:          c.strPtrToJSONPtr(in.Actions),
		Tags:             c.strPtrToJSONPtr(in.Tags),
		LastUpdated:      graphql.Timestamp(in.LastUpdated),
		Extensions:       c.strPtrToJSONPtr(in.Extensions),
		Bundles:          nil,
		Bundle:           nil,
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

func (c *converter) InputFromGraphQL(in graphql.PackageInput) (model.PackageInput, error) {
	bundles, err := c.bundle.MultipleCreateInputFromGraphQL(in.Bundles)
	if err != nil {
		return model.PackageInput{}, errors.Wrap(err, "while converting Bundles input")
	}

	return model.PackageInput{
		ID:               c.strPrtToStr(in.ID),
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Description:      in.Description,
		Version:          in.Version,
		Licence:          in.Licence,
		LicenceType:      in.LicenceType,
		TermsOfService:   in.TermsOfService,
		Logo:             in.Logo,
		Image:            in.Image,
		Provider:         c.jsonPtrToStrPtr(in.Provider),
		Actions:          c.jsonPtrToStrPtr(in.Actions),
		Tags:             c.jsonPtrToStrPtr(in.Tags),
		LastUpdated:      time.Time(in.LastUpdated),
		Extensions:       c.jsonPtrToStrPtr(in.Extensions),
		Bundles:          bundles,
	}, nil
}

func (c *converter) MultipleCreateInputFromGraphQL(in []*graphql.PackageInput) ([]*model.PackageInput, error) {
	var packages []*model.PackageInput
	for _, item := range in {
		if item == nil {
			continue
		}
		pkg, err := c.InputFromGraphQL(*item)
		if err != nil {
			return nil, err
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}

func (c *converter) strPtrToJSONSchemaPtr(in *string) *graphql.JSONSchema {
	if in == nil {
		return nil
	}
	out := graphql.JSONSchema(*in)
	return &out
}

func (c *converter) strPtrToJSONPtr(in *string) *graphql.JSON {
	if in == nil {
		return nil
	}
	out := graphql.JSON(*in)
	return &out
}

func (c *converter) jsonSchemaPtrToStrPtr(in *graphql.JSONSchema) *string {
	if in == nil {
		return nil
	}
	out := string(*in)
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
