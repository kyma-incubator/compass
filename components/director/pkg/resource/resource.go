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
	// Specification missing godoc
	Specification Type = "specification"
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
	// Tenant missing godoc
	Tenant Type = "tenant"
	// TenantIndex missing godoc
	TenantIndex Type = "tenantIndex"
	// Schema missing godoc
	Schema Type = "schemaMigration"
)

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
