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
	docKind              = "fookind"
	docTitle             = "footitle"
	docData              = "foodata"
	docDisplayName       = "foodisplay"
	docDescription       = "foodesc"
	docCLOB              = graphql.CLOB(docData)
	appID                = "appID"
	appTemplateVersionID = "appTemplateVersionID"
	fixedTimestamp       = time.Now()
	docID                = "foo"
)

func fixModelDocument(id, bundleID string) *model.Document {
	return &model.Document{
		BundleID:    bundleID,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      model.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docData,
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixModelDocumentForApp(id, bundleID string) *model.Document {
	doc := fixModelDocument(id, bundleID)
	doc.AppID = &appID

	return doc
}

func fixModelDocumentForAppTemplateVersion(id, bundleID string) *model.Document {
	doc := fixModelDocument(id, bundleID)
	doc.ApplicationTemplateVersionID = &appTemplateVersionID

	return doc
}

func fixEntityDocument(id, bundleID string) *document.Entity {
	return &document.Entity{
		BndlID:      bundleID,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      string(model.DocumentFormatMarkdown),
		Kind:        repo.NewValidNullableString(docKind),
		Data:        repo.NewValidNullableString(docData),
		BaseEntity: &repo.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     sql.NullString{},
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixEntityDocumentForApp(id, bundleID string) *document.Entity {
	entity := fixEntityDocument(id, bundleID)
	entity.AppID = repo.NewValidNullableString(appID)
	return entity
}

func fixEntityDocumentForAppTemplateVersion(id, bundleID string) *document.Entity {
	entity := fixEntityDocument(id, bundleID)
	entity.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return entity
}

func fixGQLDocument(id, bundleID string) *graphql.Document {
	return &graphql.Document{
		BundleID:    bundleID,
		Title:       docTitle,
		DisplayName: docDisplayName,
		Description: docDescription,
		Format:      graphql.DocumentFormatMarkdown,
		Kind:        &docKind,
		Data:        &docCLOB,
		BaseEntity: &graphql.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
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

func fixModelBundle(id string) *model.Bundle {
	return &model.Bundle{
		ApplicationID: &appID,
		Name:          "name",
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}
