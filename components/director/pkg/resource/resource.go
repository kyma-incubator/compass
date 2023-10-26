package resource

// Type represents a resource type in compass.
type Type string

const (
	// Destination type represents destination resource.
	Destination Type = "destination"
	// Application type represents application resource.
	Application Type = "application"
	// ApplicationTemplate type represents application template resource.
	ApplicationTemplate Type = "applicationTemplate"
	// ApplicationTemplateVersion type represents application template version resource.
	ApplicationTemplateVersion Type = "applicationTemplateVersion"
	// Runtime type represents runtime resource.
	Runtime Type = "runtime"
	// RuntimeContext type represents runtime context resource.
	RuntimeContext Type = "runtimeContext"
	// LabelDefinition type represents label definition resource.
	LabelDefinition Type = "labelDefinition"
	// Label type represents generic label resource. This resource type does not assume the referenced resource type of the label.
	Label Type = "label"
	// RuntimeLabel type represents runtime label resource.
	RuntimeLabel Type = "runtimeLabel"
	// RuntimeContextLabel type represents runtime context label resource.
	RuntimeContextLabel Type = "runtimeContextLabel"
	// ApplicationLabel type represents application label resource.
	ApplicationLabel Type = "applicationLabel"
	// TenantLabel type represents tenant label resource.
	TenantLabel Type = "tenantLabel"
	// Bundle type represents bundle resource.
	Bundle Type = "bundle"
	// BundleReference type represents bundle reference resource.
	BundleReference Type = "bundleReference"
	// Package type represents package resource.
	Package Type = "package"
	// Product type represents product resource.
	Product Type = "product"
	// Vendor type represents vendor resource.
	Vendor Type = "vendor"
	// Tombstone type represents tombstone resource.
	Tombstone Type = "tombstone"
	// IntegrationSystem type represents integration system resource.
	IntegrationSystem Type = "integrationSystem"
	// SystemAuth type represents system auth resource.
	SystemAuth Type = "systemAuth"
	// FetchRequest type represents generic fetch request resource. This resource does not assume the referenced resource type of the FR.
	FetchRequest Type = "fetchRequest"
	// DocFetchRequest type represents document fetch request resource.
	DocFetchRequest Type = "docFetchRequest"
	// APISpecFetchRequest type represents API specification fetch request resource.
	APISpecFetchRequest Type = "apiSpecFetchRequest"
	// EventSpecFetchRequest type represents Event specification fetch request resource.
	EventSpecFetchRequest Type = "eventSpecFetchRequest"
	// CapabilitySpecFetchRequest type represents Capability specification fetch request resource.
	CapabilitySpecFetchRequest Type = "capabilitySpecFetchRequest"
	// Specification type represents generic specification resource. This resource does not assume the referenced resource type of the Spec.
	Specification Type = "specification"
	// EntityType type represents Entity Type resource.
	EntityType Type = "entityType"
	// EntityTypeMapping type represents Entity Type Mapping resource.
	EntityTypeMapping Type = "entityTypeMapping"
	// APISpecification type represents API specification resource.
	APISpecification Type = "apiSpecification"
	// EventSpecification type represents Event specification resource.
	EventSpecification Type = "eventSpecification"
	// CapabilitySpecification type represents Capability specification resource.
	CapabilitySpecification Type = "capabilitySpecification"
	// Document type represents document resource.
	Document Type = "document"
	// BundleInstanceAuth type represents bundle instance auth resource.
	BundleInstanceAuth Type = "bundleInstanceAuth"
	// API type represents api resource.
	API Type = "api"
	// EventDefinition type represents event resource.
	EventDefinition Type = "eventDefinition"
	// Capability type represents capability resource.
	Capability Type = "capability"
	// AutomaticScenarioAssigment type represents ASA resource.
	AutomaticScenarioAssigment Type = "automaticScenarioAssigment"
	// CertSubjectMapping type represents certificate subject mapping resource
	CertSubjectMapping Type = "certSubjectMapping"
	// Formations type represents formations resource.
	Formations Type = "formations"
	// FormationTemplate type represents formation template resource.
	FormationTemplate Type = "formationTemplate"
	// FormationAssignment type represents formation assignment resource.
	FormationAssignment Type = "formationAssignment"
	// FormationConstraint type represents formation constraint resource.
	FormationConstraint Type = "formationConstraint"
	// FormationTemplateConstraintReference type represents formationTemplate-constraint reference resource.
	FormationTemplateConstraintReference Type = "formationTemplateConstraintReference"
	// Webhook type represents generic webhook resource. This resource does not assume the referenced resource type of the Webhook.
	Webhook Type = "webhook"
	// AppWebhook type represents application webhook resource.
	AppWebhook Type = "appWebhook"
	// RuntimeWebhook type represents runtime webhook resource.
	RuntimeWebhook Type = "runtimeWebhook"
	// Operation type represents operation resource.
	Operation Type = "operation"
	// FormationTemplateWebhook type represents formation template webhook resource.
	FormationTemplateWebhook Type = "formationTemplateWebhook"
	// Tenant type represents tenant resource.
	Tenant Type = "tenant"
	// TenantAccess type represents tenant access resource.
	TenantAccess Type = "tenantAccess"
	// TenantBusinessType type represents tenant business type resource.
	TenantBusinessType Type = "tenantBusinessType"
	// Schema type represents schema resource.
	Schema Type = "schemaMigration"
	// SystemsSync type represents systems synchronization timestamps resource
	SystemsSync Type = "systemsSync"
)

var ignoredTenantAccessTable = map[Type]string{
	ApplicationTemplateVersion: "application_template_versions",
}

var tenantAccessTable = map[Type]string{
	// Tables

	Application:    "tenant_applications",
	Runtime:        "tenant_runtimes",
	RuntimeContext: "tenant_runtime_contexts",

	// Views

	Label:                      "labels_tenants",
	ApplicationLabel:           "application_labels_tenants",
	RuntimeLabel:               "runtime_labels_tenants",
	RuntimeContextLabel:        "runtime_contexts_labels_tenants",
	Bundle:                     "bundles_tenants",
	Package:                    "packages_tenants",
	Product:                    "products_tenants",
	Vendor:                     "vendors_tenants",
	Tombstone:                  "tombstones_tenants",
	DocFetchRequest:            "document_fetch_requests_tenants",
	APISpecFetchRequest:        "api_specifications_fetch_requests_tenants",
	EventSpecFetchRequest:      "event_specifications_fetch_requests_tenants",
	CapabilitySpecFetchRequest: "capability_specifications_fetch_requests_tenants",
	APISpecification:           "api_specifications_tenants",
	EventSpecification:         "event_specifications_tenants",
	CapabilitySpecification:    "capability_specifications_tenants",
	Document:                   "documents_tenants",
	BundleInstanceAuth:         "bundle_instance_auths_tenants",
	API:                        "api_definitions_tenants",
	EventDefinition:            "event_api_definitions_tenants",
	EntityType:                 "entity_types_tenants",
	EntityTypeMapping:          "entity_type_mappings_tenants",
	Capability:                 "capabilities_tenants",
	Webhook:                    "webhooks_tenants",
	AppWebhook:                 "application_webhooks_tenants",
	RuntimeWebhook:             "runtime_webhooks_tenants",
	FormationTemplateWebhook:   "formation_templates_webhooks_tenants",
}

var tablesWithEmbeddedTenant = map[Type]string{
	// Tables
	FormationTemplate:          "formation_templates",
	AutomaticScenarioAssigment: "automatic_scenario_assignments",
	LabelDefinition:            "label_definitions",
	Label:                      "labels",
	SystemAuth:                 "system_auths",
}

// IgnoredTenantAccessTable returns the table / view with tenant accesses of the given type.
func (t Type) IgnoredTenantAccessTable() (string, bool) {
	tbl, ok := ignoredTenantAccessTable[t]
	return tbl, ok
}

// IsTenantIgnorable returns true if the entity has a Global access and does not need tenant isolation
func (t Type) IsTenantIgnorable() bool {
	_, exists := ignoredTenantAccessTable[t]
	return exists
}

// TenantAccessTable returns the table / view with tenant accesses of the given type.
func (t Type) TenantAccessTable() (string, bool) {
	tbl, ok := tenantAccessTable[t]
	return tbl, ok
}

// EmbeddedTenantTable returns the table of the given type that is with embedded tenant.
func (t Type) EmbeddedTenantTable() (string, bool) {
	tbl, ok := tablesWithEmbeddedTenant[t]
	return tbl, ok
}

// TopLevelEntities is a map of entities that has a many-to-many relationship with the tenants along with their table names.
var TopLevelEntities = map[Type]string{
	Application:                "public.applications",
	ApplicationTemplateVersion: "public.application_template_versions",
	Runtime:                    "public.runtimes",
	RuntimeContext:             "public.runtime_contexts",
}

// IsTopLevel returns true only if the entity has a many-to-many relationship with the tenants.
func (t Type) IsTopLevel() bool {
	_, exists := TopLevelEntities[t]
	return exists
}

// SQLOperation represents an SQL operation
type SQLOperation string

const (
	// Create represents Create SQL operation
	Create SQLOperation = "Create"
	// Update represents Update SQL operation
	Update SQLOperation = "Update"
	// Upsert represents Upsert SQL operation
	Upsert SQLOperation = "Upsert"
	// Delete represents Delete SQL operation
	Delete SQLOperation = "Delete"
	// Exists represents Exists SQL operation
	Exists SQLOperation = "Exists"
	// Get represents Get SQL operation
	Get SQLOperation = "Get"
	// List represents List SQL operation
	List SQLOperation = "List"
)
