package model

// DefaultTemplateName will be used as default formation templane name if no other options are provided
var DefaultTemplateName = "Side-by-side extensibility with Kyma"

// Formation missing godoc
type Formation struct {
	ID                  string
	TenantID            string
	FormationTemplateID string
	Name                string
}
