package graphql

type Package struct {
	ID                             string      `json:"id"`
	Name                           string      `json:"name"`
	Description                    *string     `json:"description"`
	InstanceAuthRequestInputSchema *JSONSchema `json:"InstanceAuthRequestInputSchema"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultInstanceAuth *Auth `json:"defaultInstanceAuth"`
}
