package document_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

var (
	docKind        = "fookind"
	docTitle       = "footitle"
	docData        = "foodata"
	docDisplayName = "foodisplay"
	docDescription = "foodesc"
	docCLOB        = graphql.CLOB(docData)
)

func fixModelDocument(applicationID, id string) *model.Document {
	return &model.Document{
		ApplicationID: applicationID,
		ID:            id,
		Title:         docTitle,
		Format:        model.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docData,
		FetchRequest:  &model.FetchRequest{},
	}
}

func fixGQLDocument(id string) *graphql.Document {
	return &graphql.Document{
		ID:           id,
		Title:        docTitle,
		Format:       graphql.DocumentFormatMarkdown,
		Kind:         &docKind,
		Data:         &docCLOB,
		FetchRequest: &graphql.FetchRequest{},
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
