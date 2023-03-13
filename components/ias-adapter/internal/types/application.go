package types

type Applications struct {
	Applications []Application `json:"applications"`
}

type Application struct {
	ID             string                    `json:"id"`
	Authentication ApplicationAuthentication `json:"urn:sap:identity:application:schemas:extension:sci:1.0:Authentication"`
}

type ApplicationAuthentication struct {
	ConsumedAPIs []ApplicationConsumedAPI `json:"consumedApis"`
}

type ApplicationConsumedAPI struct {
	Name     string `json:"name"`
	APIName  string `json:"apiName"`
	AppID    string `json:"appId"`
	ClientID string `json:"clientId"`
}

type ApplicationUpdate struct {
	Operations []ApplicationUpdateOperation `json:"operations"`
}

type ApplicationUpdateOperation struct {
	Operation string `json:"op"`
	Path      string `json:"path"`
}

type ApplicationUpdateValue struct {
	Name    string `json:"name"`
	APIName string `json:"apiName"`
	AppID   string `json:"appId"`
}
