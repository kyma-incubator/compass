package resource

import "strings"

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
	Specification              Type = "Specification"
	Document                   Type = "Document"
	BundleInstanceAuth         Type = "BundleInstanceAuth"
	API                        Type = "Api"
	EventDefinition            Type = "EventDefinition"
	AutomaticScenarioAssigment Type = "AutomaticScenarioAssigment"
	Webhook                    Type = "Webhook"
)

// ToLower returns the lower-case string representation of a resource Type
func (t Type) ToLower() string {
	return strings.ToLower(string(t))
}

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
