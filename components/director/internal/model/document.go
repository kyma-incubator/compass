package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Document struct {
	ApplicationID string
	ID            string
	Tenant        string
	Title         string
	DisplayName   string
	Description   string
	Format        DocumentFormat
	// for example Service Class, API etc
	Kind           *string
	Data           *string
	FetchRequestID *string
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

func (d *DocumentInput) ToDocument(id, tenant, applicationID string, fetchRequestID *string) *Document {
	if d == nil {
		return nil
	}

	return &Document{
		ApplicationID:  applicationID,
		ID:             id,
		Tenant:         tenant,
		Title:          d.Title,
		DisplayName:    d.DisplayName,
		Description:    d.Description,
		Format:         d.Format,
		Kind:           d.Kind,
		Data:           d.Data,
		FetchRequestID: fetchRequestID,
	}
}
