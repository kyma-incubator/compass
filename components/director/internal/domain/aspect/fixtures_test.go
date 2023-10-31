package aspect_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"time"
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
	mandatory                = true
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

func fixEntityAspect(id string) *aspect.Entity {
	return &aspect.Entity{
		ApplicationID:                repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID: repo.NewValidNullableString(appTemplateVersionID),
		IntegrationDependencyID:      integrationDependencyID,
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

func fixAspectCreateArgs(id string, aspect *model.Aspect) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(*aspect.ApplicationTemplateVersionID), aspect.IntegrationDependencyID, aspect.Title, repo.NewValidNullableString(*aspect.Description),
		repo.NewNullableBool(aspect.Mandatory), repo.NewNullableBool(aspect.SupportMultipleProviders), repo.NewNullableStringFromJSONRawMessage(aspect.APIResources), repo.NewNullableStringFromJSONRawMessage(aspect.EventResources),
		ready, fixedTimestamp, time.Time{}, time.Time{}, nil,
	}
}
