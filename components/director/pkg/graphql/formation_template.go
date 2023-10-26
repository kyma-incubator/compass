package graphql

// FormationTemplate represents the Formation Template object
type FormationTemplate struct {
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	ApplicationTypes       []string      `json:"applicationTypes"`
	RuntimeTypes           []string      `json:"runtimeTypes"`
	RuntimeTypeDisplayName *string       `json:"runtimeTypeDisplayName"`
	RuntimeArtifactKind    *ArtifactType `json:"runtimeArtifactKind"`
	Webhooks               []*Webhook    `json:"webhooks"`
	LeadingProductIDs      []string      `json:"leadingProductIDs"`
	SupportsReset          bool          `json:"supportsReset"`
}

// FormationTemplateExt  is an extended types used by external API
type FormationTemplateExt struct {
	FormationTemplate
	FormationConstraints []FormationConstraint `json:"formationConstraints"`
}
