package eventing

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func RuntimeEventingConfigurationToGraphQL(in *model.RuntimeEventingConfiguration) *graphql.RuntimeEventingConfiguration {
	if in == nil {
		return nil
	}

	return &graphql.RuntimeEventingConfiguration{
		DefaultURL: in.DefaultURL.String(),
	}
}

func ApplicationEventingConfigurationToGraphQL(in *model.ApplicationEventingConfiguration) *graphql.ApplicationEventingConfiguration {
	if in == nil {
		return nil
	}

	return &graphql.ApplicationEventingConfiguration{
		DefaultURL: in.DefaultURL.String(),
	}
}
