package document_test

import (
	"time"

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

func fixModelDocumentInputWithFetchRequest(fetchRequestURL string) *model.DocumentInput {
	return &model.DocumentInput{
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      model.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docData,
		FetchRequest: &model.FetchRequestInput{
			URL: fetchRequestURL,
		},
	}
}

func fixModelFetchRequest(id, url string, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		Tenant: "tenant",
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.DocumentFetchRequestReference,
		ObjectID:   "foo",
	}
}

func fixGQLFetchRequest(url string, timestamp time.Time) *graphql.FetchRequest {
	return &graphql.FetchRequest{
		Filter: nil,
		Mode:   graphql.FetchModeSingle,
		Auth:   nil,
		URL:    url,
		Status: &graphql.FetchRequestStatus{
			Timestamp: graphql.Timestamp(timestamp),
			Condition: graphql.FetchRequestStatusConditionInitial,
		},
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
