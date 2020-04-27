package edp

type (
	DataTenantPayload struct {
		Name        string `json:"name"`
		Environment string `json:"environment"`
		Secret      string `json:"secret"`
	}

	MetadataTenantPayload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	MetadataItem struct {
		DataTenant DataTenantItem `json:"dataTenant"`
		Key        string         `json:"key"`
		Value      string         `json:"value"`
	}

	DataTenantItem struct {
		Namespace   NamespaceItem `json:"namespace"`
		Name        string        `json:"name"`
		Environment string        `json:"environment"`
	}

	NamespaceItem struct {
		Name string `json:"name"`
	}
)
