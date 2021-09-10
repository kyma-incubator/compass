package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// Document missing godoc
type Document struct {
	BundleID    string         `json:"bundleID"`
	Title       string         `json:"title"`
	DisplayName string         `json:"displayName"`
	Description string         `json:"description"`
	Format      DocumentFormat `json:"format"`
	// for example Service Class, API etc
	Kind *string `json:"kind"`
	Data *CLOB   `json:"data"`
	*BaseEntity
}

// GetType missing godoc
func (e *Document) GetType() resource.Type {
	return resource.Document
}

// DocumentPageExt is an extended type used by external API
type DocumentPageExt struct {
	DocumentPage
	Data []*DocumentExt `json:"data"`
}

// DocumentExt missing godoc
type DocumentExt struct {
	Document
	FetchRequest *FetchRequest `json:"fetchRequest"`
}
