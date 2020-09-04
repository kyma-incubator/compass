package resource

type Type string

const (
	Application                Type = "Application"
	ApplicationTemplate        Type = "ApplicationTemplate"
	Runtime                    Type = "Runtime"
	LabelDefinition            Type = "LabelDefinition"
	Label                      Type = "Label"
	Bundle                     Type = "Bundle"
	Package                    Type = "Package"
	PackageBundle              Type = "PackageBundle"
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
