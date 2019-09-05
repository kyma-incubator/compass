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
	value := repo.StringPtrFromNullableString(version.VersionValue)
	versionValue := ""
	if value != nil {
		versionValue = *value
	}

	if !version.VersionForRemoval.Valid && !version.VersionValue.Valid && !version.VersionDepracated.Valid && !version.VersionDepracatedSince.Valid {
		return nil
	}

	return &model.Version{
		Value:           versionValue,
		Deprecated:      repo.BoolPtrFromNullableBool(version.VersionDepracated),
		DeprecatedSince: repo.StringPtrFromNullableString(version.VersionDepracatedSince),
		ForRemoval:      repo.BoolPtrFromNullableBool(version.VersionForRemoval),
	}
}

func (c *converter) ToEntity(version model.Version) Version {
	return Version{
		VersionValue:           repo.NewNullableString(&version.Value),
		VersionDepracated:      repo.NewNullableBool(version.Deprecated),
		VersionDepracatedSince: repo.NewNullableString(version.DeprecatedSince),
		VersionForRemoval:      repo.NewNullableBool(version.ForRemoval),
	}
}
