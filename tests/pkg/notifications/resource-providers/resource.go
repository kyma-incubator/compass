package resource_providers

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type Resource interface {
	GetType() graphql.ResourceType
	GetName() string
	GetArtifactKind() *graphql.ArtifactType // used only for runtimes, otherwise return empty
	GetDisplayName() *string                // used only for runtimes, otherwise return empty
	GetID() string
}
