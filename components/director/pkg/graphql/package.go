package graphql

type PackageDefinition struct {
	ID                    string      `json:"id"`
	Name                  string      `json:"name"`
	Description           *string     `json:"description"`
	AuthRequestJSONSchema *JSONSchema `json:"authRequestJSONSchema"`
	// When defined, all Auth requests fallback to defaultAuth.
	DefaultAuth *Auth `json:"defaultAuth"`
}
