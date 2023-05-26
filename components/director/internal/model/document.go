package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Document missing godoc
type Document struct {
	BundleID                     string
	AppID                        *string
	ApplicationTemplateVersionID *string
	Title                        string
	DisplayName                  string
	Description                  string
	Format                       DocumentFormat
	// for example Service Class, API etc
	Kind *string
	Data *string
	*BaseEntity
}

// GetType missing godoc
func (*Document) GetType() resource.Type {
	return resource.Document
}

// DocumentInput missing godoc
type DocumentInput struct {
	Title        string
	DisplayName  string
	Description  string
	Format       DocumentFormat
	Kind         *string
	Data         *string
	FetchRequest *FetchRequestInput
}

// DocumentFormat missing godoc
type DocumentFormat string

// DocumentFormatMarkdown missing godoc
const (
	DocumentFormatMarkdown DocumentFormat = "MARKDOWN"
)

// DocumentPage missing godoc
type DocumentPage struct {
	Data       []*Document
	PageInfo   *pagination.Page
	TotalCount int
}

// ToDocumentWithinBundle missing godoc
func (d *DocumentInput) ToDocumentWithinBundle(id, bundleID string, resourceType resource.Type, resourceID string) *Document {
	if d == nil {
		return nil
	}

	doc := &Document{
		BundleID:    bundleID,
		Title:       d.Title,
		DisplayName: d.DisplayName,
		Description: d.Description,
		Format:      d.Format,
		Kind:        d.Kind,
		Data:        d.Data,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType == resource.ApplicationTemplateVersion {
		doc.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		doc.AppID = &resourceID
	}

	return doc
}
