package document_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
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
		Title:         docTitle,
		Format:        model.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docData,
		FetchRequest:  &model.FetchRequest{},
	}
}

func fixGQLDocument(id, applicationID string) *graphql.Document {
	return &graphql.Document{
		ID:            id,
		ApplicationID: applicationID,
		Title:         docTitle,
		Format:        graphql.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docCLOB,
		FetchRequest:  &graphql.FetchRequest{},
	}
}

func fixModelDocumentInput(id string) *model.DocumentInput {
	return &model.DocumentInput{
		Title:        docTitle,
		DisplayName:  docDisplayName,
		Description:  docDescription,
		Format:       model.DocumentFormatMarkdown,
		Kind:         &docKind,
		Data:         &docData,
		FetchRequest: &model.FetchRequestInput{},
	}
}

func fixGQLDocumentInput(id string) *graphql.DocumentInput {
	return &graphql.DocumentInput{
		Title:        docTitle,
		DisplayName:  docDisplayName,
		Description:  docDescription,
		Format:       graphql.DocumentFormatMarkdown,
		Kind:         &docKind,
		Data:         &docCLOB,
		FetchRequest: &graphql.FetchRequestInput{},
	}
}
