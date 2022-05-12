package graphql

// ApplicationTemplate missing godoc
type ApplicationTemplate struct {
	ID               string                         `json:"id"`
	Name             string                         `json:"name"`
	Description      *string                        `json:"description"`
	Webhooks         []Webhook                      `json:"webhooks"`
	ApplicationInput string                         `json:"applicationInput"`
	Placeholders     []*PlaceholderDefinition       `json:"placeholders"`
	AccessLevel      ApplicationTemplateAccessLevel `json:"accessLevel"`
	Labels           Labels                         `json:"labels"`
}
