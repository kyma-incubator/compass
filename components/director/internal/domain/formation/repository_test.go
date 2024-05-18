package formation_test

import (
	"context"
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

var (
	formationEntity = fixFormationEntity()
	formationModel  = fixFormationModel()
)

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create Formation",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.formations \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ModelEntity:               formationModel,
		DBEntity:                  formationEntity,
		NilModelEntity:            nilFormationModel,
		IsGlobal:                  true,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation by ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TntInternalID, FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntity:       formationModel,
		ExpectedDBEntity:          formationEntity,
		MethodArgs:                []interface{}{FormationID, TntInternalID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetGlobalByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation Globally by ID",
		MethodName: "GetGlobalByID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE id = $1`),
				Args:     []driver.Value{FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntity:       formationModel,
		ExpectedDBEntity:          formationEntity,
		MethodArgs:                []interface{}{FormationID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_GetByName(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Formation By Name",
		MethodName: "GetByName",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE tenant_id = $1 AND name = $2`),
				Args:     []driver.Value{TntInternalID, testFormationName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntity:       formationModel,
		ExpectedDBEntity:          formationEntity,
		MethodArgs:                []interface{}{testFormationName, TntInternalID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_List(t *testing.T) {
	suite := testdb.RepoListPageableTestSuite{
		Name:       "List Formations ",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE tenant_id = $1 ORDER BY id LIMIT 4 OFFSET 0`),
				Args:     []driver.Value{TntInternalID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.formations`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: formation.NewRepository,
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{formationModel},
				ExpectedDBEntities:    []interface{}{formationEntity},
				ExpectedPage: &model.FormationPage{
					Data: []*model.Formation{formationModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		MethodArgs:                []interface{}{TntInternalID, 4, ""},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByFormationNames(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formations ",
		MethodName: "ListByFormationNames",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE tenant_id = $1 AND name IN ($2)`),
				Args:     []driver.Value{TntInternalID, formationModel.Name},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntities:     []interface{}{formationModel},
		ExpectedDBEntities:        []interface{}{formationEntity},
		MethodArgs:                []interface{}{[]string{formationModel.Name}, TntInternalID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByIDsGlobal(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Formations by IDs globally",
		MethodName: "ListByIDsGlobal",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp  FROM public.formations WHERE id IN ($1)`),
				Args:     []driver.Value{formationModel.ID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntities:     []interface{}{formationModel},
		ExpectedDBEntities:        []interface{}{formationEntity},
		MethodArgs:                []interface{}{[]string{formationModel.ID}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByIDs(t *testing.T) {
	tnt := "tnt"
	externalTnt := "wxternal"

	suite := testdb.RepoListTestSuite{
		Name:       "List Formations by IDs globally",
		MethodName: "ListByIDs",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE tenant_id = $1 AND id IN ($2)`),
				Args:     []driver.Value{tnt, formationModel.ID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ExpectedModelEntities:     []interface{}{formationModel},
		ExpectedDBEntities:        []interface{}{formationEntity},
		MethodArgs:                []interface{}{[]string{formationModel.ID}},
		DisableConverterErrorTest: true,
		Context:                   tenant.SaveToContext(context.TODO(), tnt, externalTnt),
	}

	suite.Run(t)
}

func TestRepository_ListObjectIDsOfTypeForFormations(t *testing.T) {
	tnt := "tnt"
	formationName := "formation_name"
	dbQuery := "SELECT DISTINCT fa.source FROM formations f JOIN formation_assignments fa ON f.id = fa.formation_id WHERE f.name IN ($1) AND f.tenant_id = $2 AND fa.source_type = $3;"
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"source"}).AddRow(ApplicationID)
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(formationName, tnt, model.FormationAssignmentTypeApplication).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		res, err := repo.ListObjectIDsOfTypeForFormations(ctx, tnt, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.NoError(t, err)
		require.Equal(t, []string{ApplicationID}, res)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error while executing query", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(formationName, tnt, model.FormationAssignmentTypeApplication).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		_, err := repo.ListObjectIDsOfTypeForFormations(ctx, tnt, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.Error(t, err)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error while getting persistence from context", func(t *testing.T) {
		ctx := context.TODO()
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		_, err := repo.ListObjectIDsOfTypeForFormations(ctx, tnt, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.Error(t, err)
	})
}

func TestRepository_ListObjectIDsOfTypeForFormationsGlobal(t *testing.T) {
	formationName := "formation_name"
	dbQuery := "SELECT DISTINCT fa.source FROM formations f JOIN formation_assignments fa ON f.id = fa.formation_id WHERE f.name IN ($1) AND fa.source_type = $2;"
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"source"}).AddRow(ApplicationID)
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(formationName, model.FormationAssignmentTypeApplication).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		res, err := repo.ListObjectIDsOfTypeForFormationsGlobal(ctx, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.NoError(t, err)
		require.Equal(t, []string{ApplicationID}, res)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error while executing query", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)

		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(formationName, model.FormationAssignmentTypeApplication).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		_, err := repo.ListObjectIDsOfTypeForFormationsGlobal(ctx, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.Error(t, err)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error while getting persistence from context", func(t *testing.T) {
		ctx := context.TODO()
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := formation.NewRepository(nil)
		_, err := repo.ListObjectIDsOfTypeForFormationsGlobal(ctx, []string{formationName}, model.FormationAssignmentTypeApplication)

		require.Error(t, err)
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.formations SET name = ?, state = ?, error = ?, last_state_change_timestamp = ?, last_notification_sent_timestamp = ? WHERE id = ? AND tenant_id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Formation by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE id = $1`),
				Args:     []driver.Value{FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:         updateStmt,
				Args:          []driver.Value{testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime, FormationID, TntInternalID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       formation.NewRepository,
		ModelEntity:               formationModel,
		DBEntity:                  formationEntity,
		NilModelEntity:            nilFormationModel,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)

	t.Run("Success when the formation state is changed and last state change timestamp is updated", func(t *testing.T) {
		// GIVEN
		formationModelWithReadyState := fixFormationModel()
		formationModelWithReadyState.State = model.ReadyFormationState

		formationEntityWithReadyState := fixFormationEntity()
		formationEntityWithReadyState.State = string(model.ReadyFormationState)

		formation.Now = func() time.Time { return defaultTime }

		emptyCtx := context.TODO()
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(emptyCtx, sqlxDB)

		rows := sqlmock.NewRows(fixColumns()).AddRow(FormationID, TntInternalID, FormationTemplateID, testFormationName, initialFormationState, testFormationEmptyError, &defaultTime, &defaultTime)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, formation_template_id, name, state, error, last_state_change_timestamp, last_notification_sent_timestamp FROM public.formations WHERE id = $1`)).
			WithArgs(FormationID).WillReturnRows(rows)

		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.formations SET name = ?, state = ?, error = ?, last_state_change_timestamp = ?, last_notification_sent_timestamp = ? WHERE id = ? AND tenant_id = ?`)).
			WithArgs(testFormationName, readyFormationState, testFormationEmptyError, sqlmock.AnyArg(), &defaultTime, FormationID, TntInternalID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", formationModelWithReadyState).Return(formationEntityWithReadyState, nil).Once()

		repo := formation.NewRepository(mockConverter)

		// WHEN
		err := repo.Update(ctx, formationModelWithReadyState)

		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_UpdateLastNotificationSentTimestamps(t *testing.T) {
	t.Run("Success when updating formation last notification sent timestamp", func(t *testing.T) {
		// GIVEN
		formationModelWithReadyState := fixFormationModel()
		formationModelWithReadyState.State = model.ReadyFormationState

		formationEntityWithReadyState := fixFormationEntity()
		formationEntityWithReadyState.State = string(model.ReadyFormationState)

		formation.Now = func() time.Time { return defaultTime }

		emptyCtx := context.TODO()
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(emptyCtx, sqlxDB)

		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.formations SET last_notification_sent_timestamp = $1 WHERE id = $2`)).
			WithArgs(defaultTime, FormationID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		mockConverter := &automock.EntityConverter{}
		repo := formation.NewRepository(mockConverter)

		// WHEN
		err := repo.UpdateLastNotificationSentTimestamps(ctx, FormationID)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when formation's last notification sent timestamp update fail", func(t *testing.T) {
		// GIVEN
		formationModelWithReadyState := fixFormationModel()
		formationModelWithReadyState.State = model.ReadyFormationState

		formationEntityWithReadyState := fixFormationEntity()
		formationEntityWithReadyState.State = string(model.ReadyFormationState)

		formation.Now = func() time.Time { return defaultTime }

		emptyCtx := context.TODO()
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(emptyCtx, sqlxDB)

		sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.formations SET last_notification_sent_timestamp = $1 WHERE id = $2`)).
			WithArgs(defaultTime, FormationID).
			WillReturnError(testErr)

		mockConverter := &automock.EntityConverter{}
		repo := formation.NewRepository(mockConverter)

		// WHEN
		err := repo.UpdateLastNotificationSentTimestamps(ctx, FormationID)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
	})
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete Formation by name",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.formations WHERE tenant_id = $1 AND name = $2`),
				Args:          []driver.Value{TntInternalID, testFormationName},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: formation.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		MethodName: "DeleteByName",
		MethodArgs: []interface{}{TntInternalID, testFormationName},
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists Formation by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.formations WHERE tenant_id = $1 AND id = $2`),
				Args:     []driver.Value{TntInternalID, FormationID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: formation.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   FormationID,
		TenantID:   TntInternalID,
		MethodName: "Exists",
		MethodArgs: []interface{}{FormationID, TntInternalID},
	}

	suite.Run(t)
}
