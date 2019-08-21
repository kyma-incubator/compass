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

func (c *converter) FromEntity(version *Version) (*model.Version, error) {
	return &model.Version{
		Value:           version.VersionValue,
		Deprecated:      repo.BoolFromSqlNullBool(&version.VersionDepracated),
		DeprecatedSince: repo.StringFromSqlNullString(&version.VersionDepracatedSince),
		ForRemoval:      repo.BoolFromSqlNullBool(&version.VersionForRemoval),
	}, nil
}
func (c *converter) ToEntity(version *model.Version) (*Version, error) {
	return &Version{
		VersionValue:           version.Value,
		VersionDepracated:      repo.NewSqlNullBool(version.Deprecated),
		VersionDepracatedSince: repo.NewSqlNullString(version.DeprecatedSince),
		VersionForRemoval:      repo.NewSqlNullBool(version.ForRemoval),
	}, nil
}
