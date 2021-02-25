package version

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToGraphQL(in *model.Version) *graphql.Version {
	if in == nil {
		return nil
	}

	return &graphql.Version{
		Value:           in.Value,
		Deprecated:      in.Deprecated,
		DeprecatedSince: in.DeprecatedSince,
		ForRemoval:      in.ForRemoval,
	}
}

func (c *converter) InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput {
	if in == nil {
		return nil
	}

	return &model.VersionInput{
		Value:           in.Value,
		Deprecated:      in.Deprecated,
		DeprecatedSince: in.DeprecatedSince,
		ForRemoval:      in.ForRemoval,
	}
}

func (c *converter) FromEntity(version Version) *model.Version {
	value := repo.StringPtrFromNullableString(version.Value)
	versionValue := ""
	if value != nil {
		versionValue = *value
	}

	if !version.ForRemoval.Valid && !version.Value.Valid && !version.Deprecated.Valid && !version.DeprecatedSince.Valid {
		return nil
	}

	return &model.Version{
		Value:           versionValue,
		Deprecated:      repo.BoolPtrFromNullableBool(version.Deprecated),
		DeprecatedSince: repo.StringPtrFromNullableString(version.DeprecatedSince),
		ForRemoval:      repo.BoolPtrFromNullableBool(version.ForRemoval),
	}
}

func (c *converter) ToEntity(version model.Version) Version {
	return Version{
		Value:           repo.NewNullableString(&version.Value),
		Deprecated:      repo.NewNullableBool(version.Deprecated),
		DeprecatedSince: repo.NewNullableString(version.DeprecatedSince),
		ForRemoval:      repo.NewNullableBool(version.ForRemoval),
	}
}
