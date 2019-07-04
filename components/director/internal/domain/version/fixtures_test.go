package version_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelVersion(t *testing.T, value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixGQLVersion(t *testing.T, value string, deprecated bool, deprecatedSince string, forRemoval bool) *graphql.Version {
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
