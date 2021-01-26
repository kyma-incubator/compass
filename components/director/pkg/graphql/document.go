package graphql

type Document struct {
	ID          string         `json:"id"`
	BundleID    string         `json:"bundleID"`
	Title       string         `json:"title"`
	DisplayName string         `json:"displayName"`
	Description string         `json:"description"`
	Format      DocumentFormat `json:"format"`
	// for example Service Class, API etc
	Kind *string `json:"kind"`
	Data *CLOB   `json:"data"`
}

// Extended types used by external API

type DocumentPageExt struct {
	DocumentPage
	Data []*DocumentExt `json:"data"`
}

type DocumentExt struct {
	Document
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
