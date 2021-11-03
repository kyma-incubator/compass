package resource

// Type missing godoc
type Type string

const (
	// Application missing godoc
	Application Type = "application"
	// ApplicationTemplate missing godoc
	ApplicationTemplate Type = "applicationTemplate"
	// Runtime missing godoc
	Runtime Type = "runtime"
	// RuntimeContext missing godoc
	RuntimeContext Type = "runtimeContext"
	// LabelDefinition missing godoc
	LabelDefinition Type = "labelDefinition"
	// Label missing godoc
	Label Type = "label"
	// RuntimeLabel missing godoc
	RuntimeLabel Type = "runtimeLabel"
	// RuntimeContextLabel missing godoc
	RuntimeContextLabel Type = "runtimeContextLabel"
	// ApplicationLabel missing godoc
	ApplicationLabel Type = "applicationLabel"
	// TenantLabel missing godoc
	TenantLabel Type = "tenantLabel"
	// Bundle missing godoc
	Bundle Type = "bundle"
	// BundleReference missing godoc
	BundleReference Type = "bundleReference"
	// Package missing godoc
	Package Type = "package"
	// Product missing godoc
	Product Type = "product"
	// Vendor missing godoc
	Vendor Type = "vendor"
	// Tombstone missing godoc
	Tombstone Type = "tombstone"
	// IntegrationSystem missing godoc
	IntegrationSystem Type = "integrationSystem"
	// SystemAuth missing godoc
	SystemAuth Type = "systemAuth"
	// FetchRequest missing godoc
	FetchRequest Type = "fetchRequest"
	// DocFetchRequest missing godoc
	DocFetchRequest Type = "docFetchRequest"
	// SpecFetchRequest missing godoc
	SpecFetchRequest Type = "specFetchRequest"
	// Specification missing godoc
	Specification Type = "specification"
	// APISpecification missing godoc
	APISpecification Type = "apiSpecification"
	// EventSpecification missing godoc
	EventSpecification Type = "eventSpecification"
	// Document missing godoc
	Document Type = "document"
	// BundleInstanceAuth missing godoc
	BundleInstanceAuth Type = "bundleInstanceAuth"
	// API missing godoc
	API Type = "api"
	// EventDefinition missing godoc
	EventDefinition Type = "eventDefinition"
	// AutomaticScenarioAssigment missing godoc
	AutomaticScenarioAssigment Type = "automaticScenarioAssigment"
	// Webhook missing godoc
	Webhook Type = "webhook"
	// AppWebhook missing godoc
	AppWebhook Type = "appWebhook"
	// RuntimeWebhook missing godoc
	RuntimeWebhook Type = "runtimeWebhook"
	// Tenant missing godoc
	Tenant Type = "tenant"
	// TenantAccess missing godoc
	TenantAccess Type = "tenantAccess"
	// Schema missing godoc
	Schema Type = "schemaMigration"
)

var tenantAccessTable = map[Type]string{
	// Tables

	Application: "tenant_applications",
	Runtime:     "tenant_runtimes",

	// Views

	RuntimeContext:      "runtime_contexts_tenants",
	ApplicationLabel:    "application_labels_tenants",
	RuntimeLabel:        "runtime_labels_tenants",
	RuntimeContextLabel: "runtime_contexts_labels_tenants",
	Bundle:              "bundles_tenants",
	Package:             "packages_tenants",
	Product:             "products_tenants",
	Vendor:              "vendors_tenants",
	Tombstone:           "tombstones_tenants",
	DocFetchRequest:     "document_fetch_requests_tenants",
	SpecFetchRequest:    "specifications_fetch_requests_tenants",

	Specification:       "specifications_tenants",
	APISpecification:    "specifications_tenants",
	EventSpecification:  "specifications_tenants",

	Document:            "documents_tenants",
	BundleInstanceAuth:  "bundle_instance_auths_tenants",
	API:                 "api_definitions_tenants",
	EventDefinition:     "event_api_definitions_tenants",
	AppWebhook:          "application_webhooks_tenants",
	RuntimeWebhook:      "runtime_webhooks_tenants",
}

var parentRelation = map[Type]Type{
	RuntimeContext:      Runtime,
	RuntimeLabel:        Runtime,
	RuntimeContextLabel: RuntimeContext,
	ApplicationLabel:    Application,
	Bundle:              Application,
	Package:             Application,
	Product:             Application,
	Vendor:              Application,
	Tombstone:           Application,
	DocFetchRequest:     Document,
	SpecFetchRequest:    Specification,
	APISpecification:    API,
	EventSpecification:  EventDefinition,
	Document:            Bundle,
	BundleInstanceAuth:  Bundle,
	API:                 Application,
	EventDefinition:     Application,
	AppWebhook:          Application,
	RuntimeWebhook:      Runtime,
}

func (t Type) TenantAccessTable() (string, bool) {
	tbl, ok := tenantAccessTable[t]
	return tbl, ok
}

func (t Type) Parent() (Type, bool) {
	parent, ok := parentRelation[t]
	return parent, ok
}

// SQLOperation missing godoc
type SQLOperation string

const (
	// Create missing godoc
	Create SQLOperation = "Create"
	// Update missing godoc
	Update SQLOperation = "Update"
	// Upsert missing godoc
	Upsert SQLOperation = "Upsert"
	// Delete missing godoc
	Delete SQLOperation = "Delete"
	// Exists missing godoc
	Exists SQLOperation = "Exists"
	// Get missing godoc
	Get SQLOperation = "Get"
	// List missing godoc
	List SQLOperation = "List"
)
