package types

const ConsumedAPIsPath = "/urn:sap:identity:application:schemas:extension:sci:1.0:Authentication/consumedApis"

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID             string                    `json:"id"`
	Authentication ApplicationAuthentication `json:"urn:sap:identity:application:schemas:extension:sci:1.0:Authentication"`
}

type ApplicationAuthentication struct {
	ConsumedAPIs         []ApplicationConsumedAPI `json:"consumedApis"`
	SAPManagedAttributes SAPManagedAttributes     `json:"sapManagedAttributes"`
}

type SAPManagedAttributes struct {
	AppTenantID string `json:"appTenantId"`
	SAPZoneID   string `json:"sapZoneId"`
}

type ApplicationConsumedAPI struct {
	Name    string `json:"name"`
	APIName string `json:"apiName"`
	AppID   string `json:"appId"`
}

type ApplicationUpdate struct {
	Operations []ApplicationUpdateOperation `json:"operations"`
}

type UpdateOperation string

const (
	ReplaceOp UpdateOperation = "replace"
)

type ApplicationUpdateOperation struct {
	Operation UpdateOperation          `json:"op"`
	Path      string                   `json:"path"`
	Value     []ApplicationConsumedAPI `json:"value"`
}
