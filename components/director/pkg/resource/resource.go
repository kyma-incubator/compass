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
	BundleReference            Type = "bundleReference"
	Package                    Type = "package"
	Product                    Type = "product"
	Vendor                     Type = "vendor"
	Tombstone                  Type = "tombstone"
	IntegrationSystem          Type = "integrationSystem"
	SystemAuth                 Type = "systemAuth"
	FetchRequest               Type = "fetchRequest"
	Specification              Type = "specification"
	Document                   Type = "document"
	BundleInstanceAuth         Type = "bundleInstanceAuth"
	API                        Type = "api"
	EventDefinition            Type = "eventDefinition"
	AutomaticScenarioAssigment Type = "automaticScenarioAssigment"
	Webhook                    Type = "webhook"
	Tenant                     Type = "tenant"
	TenantIndex                Type = "tenantIndex"
	Schema                     Type = "schema_migration"
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
