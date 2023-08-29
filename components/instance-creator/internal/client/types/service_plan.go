package types

// ServicePlan represents a Service Plan
type ServicePlan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	CatalogId         string `json:"catalog_id"`
	CatalogName       string `json:"catalog_name"`
	ServiceOfferingId string `json:"service_offering_id"`
}

// ServicePlans represents a collection of Service Plan
type ServicePlans struct {
	NumItems int           `json:"num_items"`
	Items    []ServicePlan `json:"items"`
}
