package entitytypemapping_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

func TestPgRepository_GetByID(t *testing.T) {
	entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
	entityTypeMappingEntity := fixEntityTypeMappingEntity(entityTypeMappingID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityTypeMapping",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets FROM public.entity_type_mappings WHERE id = $1 AND (id IN (SELECT id FROM entity_type_mappings_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{entityTypeMappingID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeMappingColumns()).
							AddRow(fixEntityTypeMappingRow(entityTypeMappingID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeMappingColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ExpectedModelEntity:       entityTypeMappingModel,
		ExpectedDBEntity:          entityTypeMappingEntity,
		MethodArgs:                []interface{}{tenantID, entityTypeMappingID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
	entityTypeMappingEntity := fixEntityTypeMappingEntity(entityTypeMappingID)

	suite := testdb.RepoGetTestSuite{
		Name: "Get EntityTypeMapping Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets FROM public.entity_type_mappings WHERE id = $1`),
				Args:     []driver.Value{entityTypeMappingID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeMappingColumns()).
							AddRow(fixEntityTypeMappingRow(entityTypeMappingID)...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixEntityTypeMappingColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ExpectedModelEntity:       entityTypeMappingModel,
		ExpectedDBEntity:          entityTypeMappingEntity,
		MethodArgs:                []interface{}{entityTypeMappingID},
		DisableConverterErrorTest: true,
		MethodName:                "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	firstEntityTypeMappingID := "111111111-1111-1111-1111-111111111111"
	firstEntityTypeModel := fixEntityTypeMappingModel(firstEntityTypeMappingID)
	firstEntityTypeEntity := fixEntityTypeMappingEntity(firstEntityTypeMappingID)
	secondEntityTypeMappingID := "222222222-2222-2222-2222-222222222222"
	secondEntityTypeModel := fixEntityTypeMappingModel(secondEntityTypeMappingID)
	secondEntityTypeEntity := fixEntityTypeMappingEntity(secondEntityTypeMappingID)

	suiteForAPI := testdb.RepoListTestSuite{
		Name: "List EntityTypeMapping for API and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets FROM public.entity_type_mappings WHERE api_definition_id = $1 AND (id IN (SELECT id FROM entity_type_mappings_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{testAPIDefinitionID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeMappingColumns()).AddRow(fixEntityTypeMappingRow(firstEntityTypeMappingID)...).AddRow(fixEntityTypeMappingRow(secondEntityTypeMappingID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeMappingColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ExpectedModelEntities:     []interface{}{firstEntityTypeModel, secondEntityTypeModel},
		ExpectedDBEntities:        []interface{}{firstEntityTypeEntity, secondEntityTypeEntity},
		MethodArgs:                []interface{}{tenantID, testAPIDefinitionID, resource.API},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForEvent := testdb.RepoListTestSuite{
		Name: "List EntityTypeMapping for Event and TenantID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, ready, created_at, updated_at, deleted_at, error, api_definition_id, event_definition_id, api_model_selectors, entity_type_targets FROM public.entity_type_mappings WHERE event_definition_id = $1 AND (id IN (SELECT id FROM entity_type_mappings_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{testEventDefinitionID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeMappingColumns()).AddRow(fixEntityTypeMappingRow(firstEntityTypeMappingID)...).AddRow(fixEntityTypeMappingRow(secondEntityTypeMappingID)...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixEntityTypeMappingColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ExpectedModelEntities:     []interface{}{firstEntityTypeModel, secondEntityTypeModel},
		ExpectedDBEntities:        []interface{}{firstEntityTypeEntity, secondEntityTypeEntity},
		MethodArgs:                []interface{}{tenantID, testEventDefinitionID, resource.EventDefinition},
		MethodName:                "ListByResourceID",
		DisableConverterErrorTest: true,
	}

	suiteForAPI.Run(t)
	suiteForEvent.Run(t)
}

func TestPgRepository_CreateEntityTypeMappingInAPI(t *testing.T) {
	// GIVEN
	var nilEntityTypeMappingModel *model.EntityTypeMapping
	entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
	entityTypeMappingModel.EventDefinitionID = nil
	entityTypeMappingEntity := fixEntityTypeMappingEntity(entityTypeMappingID)
	entityTypeMappingEntity.EventDefinitionID = repo.NewNullableString(nil)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create EntityTypeMapping",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM api_definitions_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, testAPIDefinitionID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.entity_type_mappings \(.+\) VALUES \(.+\)$`,
				Args:        fixEntityTypeMappingCreateArgs(entityTypeMappingID, entityTypeMappingModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ModelEntity:               entityTypeMappingModel,
		DBEntity:                  entityTypeMappingEntity,
		NilModelEntity:            nilEntityTypeMappingModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateEntityTypeMappingInEvent(t *testing.T) {
	// GIVEN
	var nilEntityTypeMappingModel *model.EntityTypeMapping
	entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
	entityTypeMappingModel.APIDefinitionID = nil
	entityTypeMappingEntity := fixEntityTypeMappingEntity(entityTypeMappingID)
	entityTypeMappingEntity.APIDefinitionID = repo.NewNullableString(nil)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create EntityTypeMapping",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM event_api_definitions_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, testEventDefinitionID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.entity_type_mappings \(.+\) VALUES \(.+\)$`,
				Args:        fixEntityTypeMappingCreateArgs(entityTypeMappingID, entityTypeMappingModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ModelEntity:               entityTypeMappingModel,
		DBEntity:                  entityTypeMappingEntity,
		NilModelEntity:            nilEntityTypeMappingModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	var nilEntityTypeModel *model.EntityTypeMapping
	entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
	entityTypeMappingEntity := fixEntityTypeMappingEntity(entityTypeMappingID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create EntityTypeMapping Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.entity_type_mappings \(.+\) VALUES \(.+\)$`,
				Args:        fixEntityTypeMappingCreateArgs(entityTypeMappingID, entityTypeMappingModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc:       entitytypemapping.NewRepository,
		ModelEntity:               entityTypeMappingModel,
		DBEntity:                  entityTypeMappingEntity,
		NilModelEntity:            nilEntityTypeModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "EntityTypeMapping Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.entity_type_mappings WHERE id = $1 AND (id IN (SELECT id FROM entity_type_mappings_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{entityTypeMappingID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc: entitytypemapping.NewRepository,
		MethodArgs:          []interface{}{tenantID, entityTypeMappingID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "EntityTypeMapping Delete Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.entity_type_mappings WHERE id = $1`),
				Args:          []driver.Value{entityTypeMappingID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityTypeMappingConverter{}
		},
		RepoConstructorFunc: entitytypemapping.NewRepository,
		MethodArgs:          []interface{}{entityTypeMappingID},
		IsGlobal:            true,
		MethodName:          "DeleteGlobal",
	}

	suite.Run(t)
}
