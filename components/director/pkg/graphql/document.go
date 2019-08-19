package graphql

type Document struct {
	ID            string         `json:"id"`
	ApplicationID string         `json:"applicationID"`
	Title         string         `json:"title"`
	DisplayName   string         `json:"displayName"`
	Description   string         `json:"description"`
	Format        DocumentFormat `json:"format"`
	// for example Service Class, API etc
	Kind *string `json:"kind"`
	Data *CLOB   `json:"data"`
}

// Extended types used by external API

type DocumentPageExt struct {
	Data       []*DocumentExt `json:"data"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

type DocumentExt struct {
	Document
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
