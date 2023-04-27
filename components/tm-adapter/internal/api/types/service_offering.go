package types

type ServiceOffering struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BrokerId    string `json:"broker_id"`
	CatalogId   string `json:"catalog_id"`
	CatalogName string `json:"catalog_name"`
}

type ServiceOfferings struct {
	NumItems int               `json:"num_items"`
	Items    []ServiceOffering `json:"items"`
}
