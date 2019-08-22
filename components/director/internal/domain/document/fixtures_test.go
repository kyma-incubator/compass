package document_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	docTenant      = "tenant"
	docKind        = "fookind"
	docTitle       = "footitle"
	docData        = "foodata"
	docDisplayName = "foodisplay"
	docDescription = "foodesc"
	docCLOB        = graphql.CLOB(docData)
)

func fixModelDocument(id, applicationID string) *model.Document {
	return &model.Document{
		ID:            id,
		ApplicationID: applicationID,
		Tenant:        docTenant,
		Title:         docTitle,
		DisplayName:   docDisplayName,
		Description:   docDescription,
		Format:        model.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docData,
	}
}

func fixGQLDocument(id, applicationID string) *graphql.Document {
	return &graphql.Document{
		ID:            id,
		ApplicationID: applicationID,
		Title:         docTitle,
		DisplayName:   docDisplayName,
		Description:   docDescription,
		Format:        graphql.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docCLOB,
	}
}

func fixModelDocumentInput(id string) *model.DocumentInput {
	return &model.DocumentInput{
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      model.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docData,
	}
}

func fixGQLDocumentInput(id string) *graphql.DocumentInput {
	return &graphql.DocumentInput{
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      graphql.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docCLOB,
	}
}
