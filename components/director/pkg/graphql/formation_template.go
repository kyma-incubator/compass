package graphql

type FormationTemplate struct {
	ID                     string       `json:"id"`
	Name                   string       `json:"name"`
	ApplicationTypes       []string     `json:"applicationTypes"`
	RuntimeTypes           []string     `json:"runtimeTypes"`
	RuntimeTypeDisplayName string       `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    ArtifactType `json:"runtimeArtifactKind"`
	Webhooks               []*Webhook   `json:"webhooks"`
}
