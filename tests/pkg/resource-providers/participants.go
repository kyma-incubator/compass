package resource_providers

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type Resource interface {
	GetType() string
	GetName() string
	GetArtifactKind() *graphql.ArtifactType // used only for runtimes, otherwise return empty
	GetDisplayName() *string                // used only for runtimes, otherwise return empty
	GetID() string
}

type ApplicationTemplateParticipant struct {
	tpl graphql.ApplicationTemplate
}

func NewApplicationTemplateParticipant(tpl graphql.ApplicationTemplate) *ApplicationTemplateParticipant {
	return &ApplicationTemplateParticipant{
		tpl: tpl,
	}
}

func (p *ApplicationTemplateParticipant) GetType() string {
	return "APPLICATION"
}

func (p *ApplicationTemplateParticipant) GetName() string {
	return p.tpl.Name
}

// GetArtifactKind used only for runtimes, otherwise return empty
func (p *ApplicationTemplateParticipant) GetArtifactKind() *graphql.ArtifactType {
	return nil
}

func (p *ApplicationTemplateParticipant) GetDisplayName() *string {
	return nil
}

func (p *ApplicationTemplateParticipant) GetID() string {
	return p.tpl.ID
}
