package aspect_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	aspectID                = "aspectID"
	integrationDependencyID = "integrationDependencyID"
	tenantID                = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID        = "external-tnt"
	description             = "description"
	title                   = "title"
)

var (
	fixedTimestamp           = time.Now()
	appID                    = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID     = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
	mandatory                = false
	supportMultipleProviders = true
	ready                    = true
	testErr                  = errors.New("test error")
)

func fixAspectModel(id string) *model.Aspect {
	return &model.Aspect{
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		IntegrationDependencyID:      integrationDependencyID,
		Title:                        title,
		Description:                  str.Ptr(description),
		Mandatory:                    &mandatory,
		SupportMultipleProviders:     &supportMultipleProviders,
		APIResources:                 json.RawMessage("[]"),
		EventResources:               json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     ready,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}
}

func fixEntityAspect(id, appID, intDepID string) *aspect.Entity {
	return &aspect.Entity{
		ApplicationID:                repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID: repo.NewValidNullableString(appTemplateVersionID),
		IntegrationDependencyID:      intDepID,
		Title:                        title,
		Description:                  repo.NewValidNullableString(description),
		Mandatory:                    repo.NewValidNullableBool(mandatory),
		SupportMultipleProviders:     repo.NewValidNullableBool(supportMultipleProviders),
		APIResources:                 repo.NewValidNullableString("[]"),
		EventResources:               repo.NewValidNullableString("[]"),
		BaseEntity: &repo.BaseEntity{
			ID:        id,
			Ready:     ready,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		}}
}

func fixAspectInputModel() model.AspectInput {
	return model.AspectInput{
		Title:                    title,
		Description:              str.Ptr(description),
		Mandatory:                &mandatory,
		SupportMultipleProviders: &supportMultipleProviders,
		APIResources:             json.RawMessage("[]"),
		EventResources:           json.RawMessage("[]"),
	}
}

func fixGQLAspect(id string) *graphql.Aspect {
	return &graphql.Aspect{
		Name:           title,
		Description:    str.Ptr(description),
		Mandatory:      &mandatory,
		APIResources:   []*graphql.AspectAPIDefinition{},
		EventResources: []*graphql.AspectEventDefinition{},
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

func fixGQLAspectInput() *graphql.AspectInput {
	return &graphql.AspectInput{
		Name:           title,
		Description:    str.Ptr(description),
		Mandatory:      &mandatory,
		APIResources:   []*graphql.AspectAPIDefinitionInput{},
		EventResources: []*graphql.AspectEventDefinitionInput{},
	}
}

func fixAspectCreateArgs(id string, aspect *model.Aspect) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(*aspect.ApplicationTemplateVersionID), aspect.IntegrationDependencyID, aspect.Title, repo.NewValidNullableString(*aspect.Description),
		repo.NewNullableBool(aspect.Mandatory), repo.NewNullableBool(aspect.SupportMultipleProviders), repo.NewNullableStringFromJSONRawMessage(aspect.APIResources), repo.NewNullableStringFromJSONRawMessage(aspect.EventResources),
		ready, fixedTimestamp, time.Time{}, time.Time{}, nil,
	}
}

func fixAspectColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "integration_dependency_id", "title", "description", "mandatory", "support_multiple_providers", "api_resources", "event_resources", "ready", "created_at", "updated_at", "deleted_at", "error"}
}

func fixAspectRowWithArgs(id, applicationID, intDepID string) []driver.Value {
	return []driver.Value{id, applicationID, appTemplateVersionID, intDepID, title, description, false, true, repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), true, fixedTimestamp, time.Time{}, time.Time{}, nil}
}

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}
