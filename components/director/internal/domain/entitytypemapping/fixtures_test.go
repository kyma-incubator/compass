package entitytypemapping_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	entityTypeMappingID = "entity-type-mapping-id"
	tenantID            = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID    = "external-tnt"
	ready               = true
)

var (
	fixedTimestamp        = time.Now()
	testAPIDefinitionID   = "testAPIDefinitionID"
	testEventDefinitionID = "testEventDefinitionID"
	testAPIModelSelectors = removeWhitespace(`[
        {
        	"type": "odata",
            "entitySetName": "A_OperationalAcctgDocItemCube"
        }
    ]`)
	testEntityTypeTargets = removeWhitespace(`[
		{
		  	"ordId": "sap.odm:entityType:WorkforcePerson:v1"
		},
		{
		  	"correlationId": "sap.s4:csnEntity:WorkForcePersonView_v1"
		},
		{
		  	"correlationId": "sap.s4:csnEntity:sap.odm.JobDetails_v1"
		}
	]`)
	errTest = errors.New("test error")
)

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}

func fixEntityTypeMappingEntity(entityTypeMappingID string) *entitytypemapping.Entity {
	return &entitytypemapping.Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        entityTypeMappingID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
		APIDefinitionID:   repo.NewValidNullableString(testAPIDefinitionID),
		EventDefinitionID: repo.NewValidNullableString(testEventDefinitionID),
		APIModelSelectors: repo.NewValidNullableString(testAPIModelSelectors),
		EntityTypeTargets: repo.NewValidNullableString(testEntityTypeTargets),
	}
}

func fixEntityTypeMappingModel(entityTypeMappingID string) *model.EntityTypeMapping {
	return &model.EntityTypeMapping{
		BaseEntity: &model.BaseEntity{
			ID:        entityTypeMappingID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		APIDefinitionID:   &testAPIDefinitionID,
		EventDefinitionID: &testEventDefinitionID,
		APIModelSelectors: json.RawMessage(testAPIModelSelectors),
		EntityTypeTargets: json.RawMessage(testEntityTypeTargets),
	}
}

func fixEntityTypeMappingInputModel() model.EntityTypeMappingInput {
	return model.EntityTypeMappingInput{
		APIModelSelectors: json.RawMessage(testAPIModelSelectors),
		EntityTypeTargets: json.RawMessage(testEntityTypeTargets),
	}
}

func fixEntityTypeMappingColumns() []string {
	return []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "api_definition_id", "event_definition_id", "api_model_selectors", "entity_type_targets"}
}

func fixEntityTypeMappingRow(id string) []driver.Value {
	return []driver.Value{id, ready, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(testAPIDefinitionID), repo.NewValidNullableString(testEventDefinitionID),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage(testAPIModelSelectors)), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(testEntityTypeTargets))}
}

func fixEntityTypeMappingCreateArgs(id string, entityTypeMapping *model.EntityTypeMapping) []driver.Value {
	return []driver.Value{id, ready, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewNullableString(entityTypeMapping.APIDefinitionID), repo.NewNullableString(entityTypeMapping.EventDefinitionID),
		repo.NewNullableStringFromJSONRawMessage(entityTypeMapping.APIModelSelectors), repo.NewNullableStringFromJSONRawMessage(entityTypeMapping.EntityTypeTargets)}
}

func fixEntityTypeMappingUpdateArgs(id string, entityTypeMapping *entitytypemapping.Entity) []driver.Value {
	return []driver.Value{entityTypeMapping.Ready, entityTypeMapping.CreatedAt, entityTypeMapping.UpdatedAt, entityTypeMapping.DeletedAt, entityTypeMapping.Error,
		entityTypeMapping.APIDefinitionID, entityTypeMapping.EventDefinitionID, entityTypeMapping.APIModelSelectors, entityTypeMapping.EntityTypeTargets, entityTypeMapping.ID}
}
