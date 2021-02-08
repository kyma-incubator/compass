package document_test

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	docTenant      = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	docKind        = "fookind"
	docTitle       = "footitle"
	docData        = "foodata"
	docDisplayName = "foodisplay"
	docDescription = "foodesc"
	docCLOB        = graphql.CLOB(docData)
)

func fixModelDocument(id, bundleID string) *model.Document {
	return fixModelDocumentWithTimestamp(id, bundleID, time.Now())
}

func fixModelDocumentWithTimestamp(id, bundleID string, createdAt time.Time) *model.Document {
	return &model.Document{
		ID:          id,
		BundleID:    bundleID,
		Tenant:      docTenant,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      model.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docData,
		BaseEntity: &model.BaseEntity{
			Ready:     true,
			Error:     nil,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
			DeletedAt: time.Time{},
		},
	}
}

func fixEntityDocument(id, bundleID string) *document.Entity {
	return fixEntityDocumentWithTimestamp(id, bundleID, time.Now())
}

func fixEntityDocumentWithTimestamp(id, bundleID string, createdAt time.Time) *document.Entity {
	return &document.Entity{
		ID:          id,
		BndlID:      bundleID,
		TenantID:    docTenant,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      string(model.DocumentFormatMarkdown),
		Kind:        repo.NewValidNullableString(docKind),
		Data:        repo.NewValidNullableString(docData),
		BaseEntity: &repo.BaseEntity{

			Ready:     true,
			Error:     sql.NullString{},
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
			DeletedAt: time.Time{},
		},
	}
}

func fixGQLDocument(id, bundleID string) *graphql.Document {
	return fixGQLDocumentWithTimestamp(id, bundleID, time.Now())
}

func fixGQLDocumentWithTimestamp(id, bundleID string, createdAt time.Time) *graphql.Document {
	return &graphql.Document{
		ID:          id,
		BundleID:    bundleID,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      graphql.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docCLOB,
		BaseEntity: &graphql.BaseEntity{
			Ready:     true,
			Error:     nil,
			CreatedAt: graphql.Timestamp(createdAt),
			UpdatedAt: graphql.Timestamp(createdAt),
			DeletedAt: graphql.Timestamp(time.Time{}),
		},
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
		Mode:   model.FetchModeSingle,
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
