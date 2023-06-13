package fetchrequest_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepository_Create(t *testing.T) {
	timestamp := time.Now()
	var nilFrModel *model.FetchRequest
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)
	eventFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.EventSpecFetchRequestReference)
	eventFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.EventSpecFetchRequestReference)
	docFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.DocumentFetchRequestReference)
	docFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.DocumentFetchRequestReference)

	apiFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create API FR",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM api_specifications_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), sql.NullString{}, "foo.bar", apiFREntity.Auth, apiFREntity.Mode, apiFREntity.Filter, apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, refID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         apiFRModel,
		DBEntity:            apiFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	eventFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create Event FR",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM event_specifications_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), sql.NullString{}, "foo.bar", eventFREntity.Auth, eventFREntity.Mode, eventFREntity.Filter, eventFREntity.StatusCondition, eventFREntity.StatusMessage, eventFREntity.StatusTimestamp, refID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         eventFRModel,
		DBEntity:            eventFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	docFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create Doc FR",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM documents_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), refID, "foo.bar", docFREntity.Auth, docFREntity.Mode, docFREntity.Filter, docFREntity.StatusCondition, docFREntity.StatusMessage, docFREntity.StatusTimestamp, sql.NullString{}},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         docFRModel,
		DBEntity:            docFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func TestRepository_CreateGlobal(t *testing.T) {
	timestamp := time.Now()
	var nilFrModel *model.FetchRequest
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)

	apiFRSuite := testdb.RepoCreateTestSuite{
		Name: "Create API FR",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.fetch_requests ( id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), sql.NullString{}, "foo.bar", apiFREntity.Auth, apiFREntity.Mode, apiFREntity.Filter, apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, refID},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         apiFRModel,
		DBEntity:            apiFREntity,
		NilModelEntity:      nilFrModel,
		IsGlobal:            true,
		MethodName:          "CreateGlobal",
	}

	apiFRSuite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	timestamp := time.Now()
	var nilFrModel *model.FetchRequest
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)
	eventFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.EventSpecFetchRequestReference)
	eventFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.EventSpecFetchRequestReference)
	docFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.DocumentFetchRequestReference)
	docFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.DocumentFetchRequestReference)

	apiFRSuite := testdb.RepoUpdateTestSuite{
		Name: "Update API Fetch Request",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.fetch_requests SET status_condition = ?, status_message = ?, status_timestamp = ? WHERE id = ? AND (id IN (SELECT id FROM api_specifications_fetch_requests_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         apiFRModel,
		DBEntity:            apiFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	apiFRSuite.Run(t)

	eventFRSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Event Fetch Request",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.fetch_requests SET status_condition = ?, status_message = ?, status_timestamp = ? WHERE id = ? AND (id IN (SELECT id FROM event_specifications_fetch_requests_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{eventFREntity.StatusCondition, eventFREntity.StatusMessage, eventFREntity.StatusTimestamp, givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         eventFRModel,
		DBEntity:            eventFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	eventFRSuite.Run(t)

	docFRSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Document Fetch Request",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.fetch_requests SET status_condition = ?, status_message = ?, status_timestamp = ? WHERE id = ? AND (id IN (SELECT id FROM document_fetch_requests_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{docFREntity.StatusCondition, docFREntity.StatusMessage, docFREntity.StatusTimestamp, givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         docFRModel,
		DBEntity:            docFREntity,
		NilModelEntity:      nilFrModel,
		TenantID:            tenantID,
	}

	docFRSuite.Run(t)
}

func TestRepository_UpdateGlobal(t *testing.T) {
	timestamp := time.Now()
	var nilFrModel *model.FetchRequest
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)

	apiFRSuite := testdb.RepoUpdateTestSuite{
		Name: "Update API Fetch Request",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.fetch_requests SET status_condition = ?, status_message = ?, status_timestamp = ? WHERE id = ?`),
				Args:          []driver.Value{apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, givenID()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		ModelEntity:         apiFRModel,
		DBEntity:            apiFREntity,
		NilModelEntity:      nilFrModel,
		IsGlobal:            true,
		UpdateMethodName:    "UpdateGlobal",
	}

	apiFRSuite.Run(t)
}

func TestRepository_GetByReferenceObjectID(t *testing.T) {
	timestamp := time.Now()
	apiFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.APISpecFetchRequestReference)
	apiFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.APISpecFetchRequestReference)
	eventFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.EventSpecFetchRequestReference)
	eventFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.EventSpecFetchRequestReference)
	docFRModel := fixFullFetchRequestModel(givenID(), timestamp, model.DocumentFetchRequestReference)
	docFREntity := fixFullFetchRequestEntity(t, givenID(), timestamp, model.DocumentFetchRequestReference)

	apiFRSuite := testdb.RepoGetTestSuite{
		Name: "Get Fetch Request by API ReferenceObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id = $1 AND (id IN (SELECT id FROM api_specifications_fetch_requests_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()).
							AddRow(givenID(), apiFREntity.DocumentID, "foo.bar", apiFREntity.Auth, apiFREntity.Mode, apiFREntity.Filter, apiFREntity.StatusCondition, apiFREntity.StatusMessage, apiFREntity.StatusTimestamp, apiFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntity:     apiFRModel,
		ExpectedDBEntity:        apiFREntity,
		MethodArgs:              []interface{}{tenantID, model.APISpecFetchRequestReference, refID},
		AdditionalConverterArgs: []interface{}{model.APISpecFetchRequestReference},
		MethodName:              "GetByReferenceObjectID",
	}

	eventFRSuite := testdb.RepoGetTestSuite{
		Name: "Get Fetch Request by Event ReferenceObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id = $1 AND (id IN (SELECT id FROM event_specifications_fetch_requests_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()).
							AddRow(givenID(), eventFREntity.DocumentID, "foo.bar", eventFREntity.Auth, eventFREntity.Mode, eventFREntity.Filter, eventFREntity.StatusCondition, eventFREntity.StatusMessage, eventFREntity.StatusTimestamp, eventFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntity:     eventFRModel,
		ExpectedDBEntity:        eventFREntity,
		MethodArgs:              []interface{}{tenantID, model.EventSpecFetchRequestReference, refID},
		AdditionalConverterArgs: []interface{}{model.EventSpecFetchRequestReference},
		MethodName:              "GetByReferenceObjectID",
	}

	docFRSuite := testdb.RepoGetTestSuite{
		Name: "Get Fetch Request by Document ReferenceObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE document_id = $1 AND (id IN (SELECT id FROM document_fetch_requests_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()).
							AddRow(givenID(), docFREntity.DocumentID, "foo.bar", docFREntity.Auth, docFREntity.Mode, docFREntity.Filter, docFREntity.StatusCondition, docFREntity.StatusMessage, docFREntity.StatusTimestamp, docFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntity:     docFRModel,
		ExpectedDBEntity:        docFREntity,
		MethodArgs:              []interface{}{tenantID, model.DocumentFetchRequestReference, refID},
		AdditionalConverterArgs: []interface{}{model.DocumentFetchRequestReference},
		MethodName:              "GetByReferenceObjectID",
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)

	// Additional tests
	t.Run("Error - Invalid Object Reference Type", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := fetchrequest.NewRepository(nil)
		// WHEN
		_, err := repo.GetByReferenceObjectID(ctx, tenantID, "test", givenID())
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("Invalid type of the Fetch Request reference object").Error())
	})
}

func TestRepository_Delete(t *testing.T) {
	apiFRSuite := testdb.RepoDeleteTestSuite{
		Name: "API Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE id = $1 AND (id IN (SELECT id FROM api_specifications_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, givenID(), model.APISpecFetchRequestReference},
	}

	eventFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Event Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE id = $1 AND (id IN (SELECT id FROM event_specifications_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, givenID(), model.EventSpecFetchRequestReference},
	}

	docFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Documents Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE id = $1 AND (id IN (SELECT id FROM document_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenID(), tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, givenID(), model.DocumentFetchRequestReference},
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func TestRepository_DeleteByReferenceObjectID(t *testing.T) {
	apiFRSuite := testdb.RepoDeleteTestSuite{
		Name: "API Fetch Request Delete By ObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE spec_id = $1 AND (id IN (SELECT id FROM api_specifications_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.APISpecFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectID",
		IsDeleteMany:        true,
	}

	eventFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Event Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE spec_id = $1 AND (id IN (SELECT id FROM event_specifications_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.EventSpecFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectID",
		IsDeleteMany:        true,
	}

	docFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Documents Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE document_id = $1 AND (id IN (SELECT id FROM document_fetch_requests_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.DocumentFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectID",
		IsDeleteMany:        true,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func TestRepository_DeleteByReferenceObjectIDGlobal(t *testing.T) {
	apiFRSuite := testdb.RepoDeleteTestSuite{
		Name: "API Fetch Request Delete By ObjectID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE spec_id = $1`),
				Args:          []driver.Value{refID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{model.APISpecFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectIDGlobal",
		IsDeleteMany:        true,
		IsGlobal:            true,
	}

	eventFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Event Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE spec_id = $1`),
				Args:          []driver.Value{refID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{model.EventSpecFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectIDGlobal",
		IsDeleteMany:        true,
		IsGlobal:            true,
	}

	docFRSuite := testdb.RepoDeleteTestSuite{
		Name: "Documents Fetch Request Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.fetch_requests WHERE document_id = $1`),
				Args:          []driver.Value{refID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: fetchrequest.NewRepository,
		MethodArgs:          []interface{}{model.DocumentFetchRequestReference, refID},
		MethodName:          "DeleteByReferenceObjectIDGlobal",
		IsDeleteMany:        true,
		IsGlobal:            true,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func TestRepository_ListByReferenceObjectIDs(t *testing.T) {
	timestamp := time.Now()
	firstFrID := "111111111-1111-1111-1111-111111111111"
	firstRefID := "refID1"
	secondFrID := "222222222-2222-2222-2222-222222222222"
	secondRefID := "refID2"

	firstAPIFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.APISpecFetchRequestReference, firstRefID)
	firstAPIFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.APISpecFetchRequestReference, firstRefID)
	secondAPIFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.APISpecFetchRequestReference, secondRefID)
	secondAPIFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.APISpecFetchRequestReference, secondRefID)

	apiFRSuite := testdb.RepoListTestSuite{
		Name: "List API Fetch Requests by Object IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id IN ($1, $2) AND (id IN (SELECT id FROM api_specifications_fetch_requests_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{firstRefID, secondRefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstAPIFREntity.DocumentID, "foo.bar", firstAPIFREntity.Auth, firstAPIFREntity.Mode, firstAPIFREntity.Filter, firstAPIFREntity.StatusCondition, firstAPIFREntity.StatusMessage, firstAPIFREntity.StatusTimestamp, firstAPIFREntity.SpecID).
						AddRow(secondFrID, secondAPIFREntity.DocumentID, "foo.bar", secondAPIFREntity.Auth, secondAPIFREntity.Mode, secondAPIFREntity.Filter, secondAPIFREntity.StatusCondition, secondAPIFREntity.StatusMessage, secondAPIFREntity.StatusTimestamp, secondAPIFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.APISpecFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstAPIFRModel, secondAPIFRModel},
		ExpectedDBEntities:      []interface{}{firstAPIFREntity, secondAPIFREntity},
		MethodArgs:              []interface{}{tenantID, model.APISpecFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDs",
		DisableEmptySliceTest:   true,
	}

	firstEventFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.EventSpecFetchRequestReference, firstRefID)
	firstEventFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.EventSpecFetchRequestReference, firstRefID)
	secondEventFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.EventSpecFetchRequestReference, secondRefID)
	secondEventFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.EventSpecFetchRequestReference, secondRefID)

	eventFRSuite := testdb.RepoListTestSuite{
		Name: "List Event Fetch Requests by Object IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id IN ($1, $2) AND (id IN (SELECT id FROM event_specifications_fetch_requests_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{firstRefID, secondRefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstEventFREntity.DocumentID, "foo.bar", firstEventFREntity.Auth, firstEventFREntity.Mode, firstEventFREntity.Filter, firstEventFREntity.StatusCondition, firstEventFREntity.StatusMessage, firstEventFREntity.StatusTimestamp, firstEventFREntity.SpecID).
						AddRow(secondFrID, secondEventFREntity.DocumentID, "foo.bar", secondEventFREntity.Auth, secondEventFREntity.Mode, secondEventFREntity.Filter, secondEventFREntity.StatusCondition, secondEventFREntity.StatusMessage, secondEventFREntity.StatusTimestamp, secondEventFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.EventSpecFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstEventFRModel, secondEventFRModel},
		ExpectedDBEntities:      []interface{}{firstEventFREntity, secondEventFREntity},
		MethodArgs:              []interface{}{tenantID, model.EventSpecFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDs",
		DisableEmptySliceTest:   true,
	}

	firstDocFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.DocumentFetchRequestReference, firstRefID)
	firstDocFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.DocumentFetchRequestReference, firstRefID)
	secondDocFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.DocumentFetchRequestReference, secondRefID)
	secondDocFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.DocumentFetchRequestReference, secondRefID)

	docFRSuite := testdb.RepoListTestSuite{
		Name: "List Doc Fetch Requests by Object IDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE document_id IN ($1, $2) AND (id IN (SELECT id FROM document_fetch_requests_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{firstRefID, secondRefID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstDocFREntity.DocumentID, "foo.bar", firstDocFREntity.Auth, firstDocFREntity.Mode, firstDocFREntity.Filter, firstDocFREntity.StatusCondition, firstDocFREntity.StatusMessage, firstDocFREntity.StatusTimestamp, firstDocFREntity.SpecID).
						AddRow(secondFrID, secondDocFREntity.DocumentID, "foo.bar", secondDocFREntity.Auth, secondDocFREntity.Mode, secondDocFREntity.Filter, secondDocFREntity.StatusCondition, secondDocFREntity.StatusMessage, secondDocFREntity.StatusTimestamp, secondDocFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.DocumentFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstDocFRModel, secondDocFRModel},
		ExpectedDBEntities:      []interface{}{firstDocFREntity, secondDocFREntity},
		MethodArgs:              []interface{}{tenantID, model.DocumentFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDs",
		DisableEmptySliceTest:   true,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func TestRepository_ListByReferenceObjectIDsGlobal(t *testing.T) {
	timestamp := time.Now()
	firstFrID := "111111111-1111-1111-1111-111111111111"
	firstRefID := "refID1"
	secondFrID := "222222222-2222-2222-2222-222222222222"
	secondRefID := "refID2"

	firstAPIFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.APISpecFetchRequestReference, firstRefID)
	firstAPIFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.APISpecFetchRequestReference, firstRefID)
	secondAPIFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.APISpecFetchRequestReference, secondRefID)
	secondAPIFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.APISpecFetchRequestReference, secondRefID)

	apiFRSuite := testdb.RepoListTestSuite{
		Name: "List API Fetch Requests by Object IDs Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id IN ($1, $2)`),
				Args:     []driver.Value{firstRefID, secondRefID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstAPIFREntity.DocumentID, "foo.bar", firstAPIFREntity.Auth, firstAPIFREntity.Mode, firstAPIFREntity.Filter, firstAPIFREntity.StatusCondition, firstAPIFREntity.StatusMessage, firstAPIFREntity.StatusTimestamp, firstAPIFREntity.SpecID).
						AddRow(secondFrID, secondAPIFREntity.DocumentID, "foo.bar", secondAPIFREntity.Auth, secondAPIFREntity.Mode, secondAPIFREntity.Filter, secondAPIFREntity.StatusCondition, secondAPIFREntity.StatusMessage, secondAPIFREntity.StatusTimestamp, secondAPIFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.APISpecFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstAPIFRModel, secondAPIFRModel},
		ExpectedDBEntities:      []interface{}{firstAPIFREntity, secondAPIFREntity},
		MethodArgs:              []interface{}{model.APISpecFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDsGlobal",
		DisableEmptySliceTest:   true,
	}

	firstEventFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.EventSpecFetchRequestReference, firstRefID)
	firstEventFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.EventSpecFetchRequestReference, firstRefID)
	secondEventFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.EventSpecFetchRequestReference, secondRefID)
	secondEventFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.EventSpecFetchRequestReference, secondRefID)

	eventFRSuite := testdb.RepoListTestSuite{
		Name: "List Event Fetch Requests by Object IDs Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE spec_id IN ($1, $2)`),
				Args:     []driver.Value{firstRefID, secondRefID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstEventFREntity.DocumentID, "foo.bar", firstEventFREntity.Auth, firstEventFREntity.Mode, firstEventFREntity.Filter, firstEventFREntity.StatusCondition, firstEventFREntity.StatusMessage, firstEventFREntity.StatusTimestamp, firstEventFREntity.SpecID).
						AddRow(secondFrID, secondEventFREntity.DocumentID, "foo.bar", secondEventFREntity.Auth, secondEventFREntity.Mode, secondEventFREntity.Filter, secondEventFREntity.StatusCondition, secondEventFREntity.StatusMessage, secondEventFREntity.StatusTimestamp, secondEventFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.EventSpecFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstEventFRModel, secondEventFRModel},
		ExpectedDBEntities:      []interface{}{firstEventFREntity, secondEventFREntity},
		MethodArgs:              []interface{}{model.EventSpecFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDsGlobal",
		DisableEmptySliceTest:   true,
	}

	firstDocFRModel := fixFullFetchRequestModelWithRefID(firstFrID, timestamp, model.DocumentFetchRequestReference, firstRefID)
	firstDocFREntity := fixFullFetchRequestEntityWithRefID(t, firstFrID, timestamp, model.DocumentFetchRequestReference, firstRefID)
	secondDocFRModel := fixFullFetchRequestModelWithRefID(secondFrID, timestamp, model.DocumentFetchRequestReference, secondRefID)
	secondDocFREntity := fixFullFetchRequestEntityWithRefID(t, secondFrID, timestamp, model.DocumentFetchRequestReference, secondRefID)

	docFRSuite := testdb.RepoListTestSuite{
		Name: "List Doc Fetch Requests by Object IDs Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, document_id, url, auth, mode, filter, status_condition, status_message, status_timestamp, spec_id FROM public.fetch_requests WHERE document_id IN ($1, $2)`),
				Args:     []driver.Value{firstRefID, secondRefID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).
						AddRow(firstFrID, firstDocFREntity.DocumentID, "foo.bar", firstDocFREntity.Auth, firstDocFREntity.Mode, firstDocFREntity.Filter, firstDocFREntity.StatusCondition, firstDocFREntity.StatusMessage, firstDocFREntity.StatusTimestamp, firstDocFREntity.SpecID).
						AddRow(secondFrID, secondDocFREntity.DocumentID, "foo.bar", secondDocFREntity.Auth, secondDocFREntity.Mode, secondDocFREntity.Filter, secondDocFREntity.StatusCondition, secondDocFREntity.StatusMessage, secondDocFREntity.StatusTimestamp, secondDocFREntity.SpecID),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		AdditionalConverterArgs: []interface{}{model.DocumentFetchRequestReference},
		RepoConstructorFunc:     fetchrequest.NewRepository,
		ExpectedModelEntities:   []interface{}{firstDocFRModel, secondDocFRModel},
		ExpectedDBEntities:      []interface{}{firstDocFREntity, secondDocFREntity},
		MethodArgs:              []interface{}{model.DocumentFetchRequestReference, []string{firstRefID, secondRefID}},
		MethodName:              "ListByReferenceObjectIDsGlobal",
		DisableEmptySliceTest:   true,
	}

	apiFRSuite.Run(t)
	eventFRSuite.Run(t)
	docFRSuite.Run(t)
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}
