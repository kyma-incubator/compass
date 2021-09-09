package eventing

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// RuntimeEventingConfigurationToGraphQL missing godoc
func RuntimeEventingConfigurationToGraphQL(in *model.RuntimeEventingConfiguration) *graphql.RuntimeEventingConfiguration {
	if in == nil {
		return nil
	}

	return &graphql.RuntimeEventingConfiguration{
		DefaultURL: in.DefaultURL.String(),
	}
}

// ApplicationEventingConfigurationToGraphQL missing godoc
func ApplicationEventingConfigurationToGraphQL(in *model.ApplicationEventingConfiguration) *graphql.ApplicationEventingConfiguration {
	if in == nil {
		return nil
	}

	return &graphql.ApplicationEventingConfiguration{
		DefaultURL: in.DefaultURL.String(),
	}
}
