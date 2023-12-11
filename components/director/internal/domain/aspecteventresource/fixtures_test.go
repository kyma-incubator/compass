package aspecteventresource_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	aspectEventResourceID = "integrationDependencyID"
	aspectID              = "aspectID"
	tenantID              = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID      = "external-tnt"
	ordID                 = "ordID"
)

var (
	fixedTimestamp       = time.Now()
	appID                = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
	minVersion           = "1.0.0"
	eventType            = "eventType"
	eventTypeRaw         = fmt.Sprintf(`[{"eventType":"%s"}]`, eventType)
	ready                = true
	testErr              = errors.New("test error")
)

func fixAspectEventResourceModel(id string) *model.AspectEventResource {
	return &model.AspectEventResource{
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		AspectID:                     aspectID,
		OrdID:                        ordID,
		MinVersion:                   str.Ptr(minVersion),
		Subset:                       json.RawMessage(eventTypeRaw),
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

func fixEntityAspectEventResource(id, appID, aspectID string) *aspecteventresource.Entity {
	return &aspecteventresource.Entity{
		ApplicationID:                repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID: repo.NewValidNullableString(appTemplateVersionID),
		AspectID:                     aspectID,
		OrdID:                        ordID,
		MinVersion:                   repo.NewValidNullableString(minVersion),
		Subset:                       repo.NewValidNullableString(eventTypeRaw),
		BaseEntity: &repo.BaseEntity{
			ID:        id,
			Ready:     ready,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		}}
}

func fixGQLAspectEventDefinition(id string) *graphql.AspectEventDefinition {
	return &graphql.AspectEventDefinition{
		OrdID: ordID,
		Subset: []*graphql.AspectEventDefinitionSubset{
			{
				EventType: str.Ptr(eventType),
			},
		},
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

func fixGQLAspectEventDefinitionInput() *graphql.AspectEventDefinitionInput {
	return &graphql.AspectEventDefinitionInput{
		OrdID: ordID,
		Subset: []*graphql.AspectEventDefinitionSubsetInput{
			{
				EventType: str.Ptr(eventType),
			},
		},
	}
}

func fixAspectEventResourceInputModel() model.AspectEventResourceInput {
	return model.AspectEventResourceInput{
		OrdID:      ordID,
		MinVersion: str.Ptr(minVersion),
		Subset:     json.RawMessage(eventTypeRaw),
	}
}

func fixAspectEventResourceCreateArgs(id string, aspectEventResource *model.AspectEventResource) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(*aspectEventResource.ApplicationTemplateVersionID), aspectEventResource.AspectID, aspectEventResource.OrdID, repo.NewValidNullableString(*aspectEventResource.MinVersion),
		repo.NewNullableStringFromJSONRawMessage(aspectEventResource.Subset), ready, fixedTimestamp, time.Time{}, time.Time{}, nil,
	}
}

func fixAspectEventResourceColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "aspect_id", "ord_id", "min_version", "subset", "ready", "created_at", "updated_at", "deleted_at", "error"}
}

func fixAspectEventResourceRowWithArgs(id, applicationID, aspectID string) []driver.Value {
	return []driver.Value{id, applicationID, appTemplateVersionID, aspectID, ordID, minVersion, repo.NewValidNullableString(eventTypeRaw), true, fixedTimestamp, time.Time{}, time.Time{}, nil}
}

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}
