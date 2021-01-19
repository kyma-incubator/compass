package resource

type Type string

const (
	Application                Type = "Application"
	ApplicationTemplate        Type = "ApplicationTemplate"
	Runtime                    Type = "Runtime"
	RuntimeContext             Type = "RuntimeContext"
	LabelDefinition            Type = "LabelDefinition"
	Label                      Type = "Label"
	Bundle                     Type = "Bundle"
	IntegrationSystem          Type = "IntegrationSystem"
	Tenant                     Type = "Tenant"
	SystemAuth                 Type = "SystemAuth"
	FetchRequest               Type = "FetchRequest"
	Document                   Type = "Document"
	BundleInstanceAuth         Type = "BundleInstanceAuth"
	API                        Type = "Api"
	EventDefinition            Type = "EventDefinition"
	AutomaticScenarioAssigment Type = "AutomaticScenarioAssigment"
	Webhook                    Type = "Webhook"
)

type SQLOperation string

const (
	Create SQLOperation = "Create"
	Update SQLOperation = "Update"
	Upsert SQLOperation = "Upsert"
	Delete SQLOperation = "Delete"
	Exists SQLOperation = "Exists"
	Get    SQLOperation = "Get"
	List   SQLOperation = "List"
)
