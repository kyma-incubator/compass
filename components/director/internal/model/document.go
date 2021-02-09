package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type Document struct {
	BundleID    string
	ID          string
	Tenant      string
	Title       string
	DisplayName string
	Description string
	Format      DocumentFormat
	// for example Service Class, API etc
	Kind *string
	Data *string
	*BaseEntity
}

func (doc *Document) GetID() string {
	return doc.ID
}

func (_ *Document) GetType() string {
	return resource.Document.ToLower()
}

type DocumentInput struct {
	Title        string
	DisplayName  string
	Description  string
	Format       DocumentFormat
	Kind         *string
	Data         *string
	FetchRequest *FetchRequestInput
}

type DocumentFormat string

const (
	DocumentFormatMarkdown DocumentFormat = "MARKDOWN"
)

type DocumentPage struct {
	Data       []*Document
	PageInfo   *pagination.Page
	TotalCount int
}

func (d *DocumentInput) ToDocumentWithinBundle(id, tenant string, bundleID string) *Document {
	if d == nil {
		return nil
	}

	return &Document{
		BundleID:    bundleID,
		ID:          id,
		Tenant:      tenant,
		Title:       d.Title,
		DisplayName: d.DisplayName,
		Description: d.Description,
		Format:      d.Format,
		Kind:        d.Kind,
		Data:        d.Data,
		BaseEntity: &BaseEntity{
			Ready: true,
		},
	}
}
