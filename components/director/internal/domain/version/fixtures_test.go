package version_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelVersion(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixGQLVersion(value string, deprecated bool, deprecatedSince string, forRemoval bool) *graphql.Version {
	return &graphql.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixModelVersionInput(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.VersionInput {
	return &model.VersionInput{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixGQLVersionInput(value string, deprecated bool, deprecatedSince string, forRemoval bool) *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixVersionEntity(value string, deprecated bool, deprecatedSince string, forRemoval bool) *version.Version {
	return &version.Version{
		Value:           repo.NewNullableString(&value),
		Deprecated:      repo.NewNullableBool(&deprecated),
		DeprecatedSince: repo.NewNullableString(&deprecatedSince),
		ForRemoval:      repo.NewNullableBool(&forRemoval),
	}
}
