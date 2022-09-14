package application

// ORDWebhookMapping missing godoc
type ORDWebhookMapping struct {
	Type                string   `json:"Type"`
	PpmsProductVersions []string `json:"PpmsProductVersions"`
	OrdURLPath          string   `json:"OrdUrlPath"`
	SubdomainSuffix     string   `json:"SubdomainSuffix"`
}
