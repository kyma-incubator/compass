package destination_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepository_Upsert(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		destRepo := destination.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.destinations ( id, name, type, url, authentication, tenant_id, bundle_id, revision, instance_id, formation_assignment_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ) ON CONFLICT ( name, instance_id, tenant_id ) DO UPDATE SET name=EXCLUDED.name, type=EXCLUDED.type, url=EXCLUDED.url, authentication=EXCLUDED.authentication, revision=EXCLUDED.revision`)
		dbMock.ExpectExec(escapedQuery).WithArgs(destinationID, destinationName, destinationType, destinationURL, destinationNoAuthn, internalDestinationSubaccountID, destinationBundleID, destinationLatestRevision, repo.NewValidNullableString(""), repo.NewValidNullableString("")).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := destRepo.Upsert(ctx, fixDestinationInput(), destinationID, internalDestinationSubaccountID, destinationBundleID, destinationLatestRevision)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_UpsertWithEmbeddedTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", destinationModel).Return(destinationEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		destRepo := destination.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.destinations ( id, name, type, url, authentication, tenant_id, bundle_id, revision, instance_id, formation_assignment_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ) ON CONFLICT ( name, instance_id, tenant_id ) DO UPDATE SET name=EXCLUDED.name, type=EXCLUDED.type, url=EXCLUDED.url, authentication=EXCLUDED.authentication, revision=EXCLUDED.revision WHERE  public.destinations.tenant_id = ?`)
		dbMock.ExpectExec(escapedQuery).WithArgs(destinationID, destinationName, destinationType, destinationURL, destinationNoAuthn, internalDestinationSubaccountID, repo.NewValidNullableString(""), repo.NewValidNullableString(""), destinationInstanceID, destinationFormationAssignmentID, internalDestinationSubaccountID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := destRepo.UpsertWithEmbeddedTenant(ctx, destinationModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when input is nil", func(t *testing.T) {
		testErr := apperrors.NewInternalError("destination model can not be empty")

		destRepo := destination.NewRepository(nil)

		// WHEN
		ctx := context.TODO()
		err := destRepo.UpsertWithEmbeddedTenant(ctx, nil)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
	})
}

func TestRepository_DeleteOld(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name:       "Delete all destinations in a given tenant that do not have latestRevision",
		MethodName: "DeleteOld",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.destinations WHERE revision != $1 AND tenant_id = $2 AND revision IS NOT NULL`),
				Args:          []driver.Value{destinationLatestRevision, internalDestinationSubaccountID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: destination.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},

		MethodArgs:   []interface{}{destinationLatestRevision, internalDestinationSubaccountID},
		IsDeleteMany: true,
		IsGlobal:     true,
	}

	suite.Run(t)
}

func TestRepository_GetDestinationByNameAndTenant(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get Destination By Name and Tenant",
		MethodName: "GetDestinationByNameAndTenant",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, type, url, authentication, tenant_id, bundle_id, revision, instance_id, formation_assignment_id FROM public.destinations WHERE tenant_id = $1 AND name = $2`),
				Args:     []driver.Value{internalDestinationSubaccountID, destinationName},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(destinationID, destinationName, destinationType, destinationURL, destinationNoAuthn, internalDestinationSubaccountID, repo.NewValidNullableString(""), repo.NewValidNullableString(""), destinationInstanceID, destinationFormationAssignmentID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       destination.NewRepository,
		ExpectedModelEntity:       destinationModel,
		ExpectedDBEntity:          destinationEntity,
		MethodArgs:                []interface{}{destinationName, internalDestinationSubaccountID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_ListByAssignmentID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name:       "List Destinations by Formation Assignment ID ",
		MethodName: "ListByAssignmentID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, type, url, authentication, tenant_id, bundle_id, revision, instance_id, formation_assignment_id FROM public.destinations WHERE formation_assignment_id = $1`),
				Args:     []driver.Value{destinationFormationAssignmentID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(destinationID, destinationName, destinationType, destinationURL, destinationNoAuthn, internalDestinationSubaccountID, repo.NewValidNullableString(""), repo.NewValidNullableString(""), destinationInstanceID, destinationFormationAssignmentID)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       destination.NewRepository,
		ExpectedModelEntities:     []interface{}{destinationModel},
		ExpectedDBEntities:        []interface{}{destinationEntity},
		MethodArgs:                []interface{}{destinationFormationAssignmentID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_DeleteByDestinationNameAndAssignmentID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name:       "Delete Destination by name and formation assignment id",
		MethodName: "DeleteByDestinationNameAndAssignmentID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.destinations WHERE tenant_id = $1 AND name = $2 AND formation_assignment_id = $3`),
				Args:          []driver.Value{internalDestinationSubaccountID, destinationName, destinationFormationAssignmentID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: destination.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},

		MethodArgs:   []interface{}{destinationName, destinationFormationAssignmentID, internalDestinationSubaccountID},
		IsDeleteMany: true,
	}

	suite.Run(t)
}
