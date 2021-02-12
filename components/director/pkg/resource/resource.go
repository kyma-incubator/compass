package resource

type Type string

const (
	Application                Type = "application"
	ApplicationTemplate        Type = "applicationTemplate"
	Runtime                    Type = "runtime"
	RuntimeContext             Type = "runtimeContext"
	LabelDefinition            Type = "labelDefinition"
	Label                      Type = "label"
	Bundle                     Type = "bundle"
	Package                    Type = "package"
	Product                    Type = "product"
	Vendor                     Type = "vendor"
	Tombstone                  Type = "tombstone"
	IntegrationSystem          Type = "integrationSystem"
	Tenant                     Type = "tenant"
	SystemAuth                 Type = "systemAuth"
	FetchRequest               Type = "fetchRequest"
	Specification              Type = "specification"
	Document                   Type = "document"
	BundleInstanceAuth         Type = "bundleInstanceAuth"
	API                        Type = "api"
	EventDefinition            Type = "eventDefinition"
	AutomaticScenarioAssigment Type = "automaticScenarioAssigment"
	Webhook                    Type = "webhook"
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
