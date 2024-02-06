package tenant_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	selfRegDistinguishLabel = "selfRegDistinguishLabel"
	parents                 = []string{testParentID, testParentID2}
	parentRows              = []sqlTenantParentsRow{
		{tenantID: testID, parentID: testParentID},
		{tenantID: testID, parentID: testParentID2},
	}
	tenantMappingModel               = newModelBusinessTenantMapping(testID, testName, parents)
	tenantMappingModelWithoutParents = newModelBusinessTenantMapping(testID, testName, []string{})
	tenantMappingEntity              = newEntityBusinessTenantMapping(testID, testName)
)

func TestPgRepository_Upsert(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)
				for _, row := range parentRows {
					dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
						WithArgs(fixTenantParentCreateArgs(row.tenantID, row.parentID)...).WillReturnResult(driver.ResultNoRows)
				}
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Success without parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModelWithoutParents).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			Input:                *tenantMappingModelWithoutParents,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while creating parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(fixTenantParentCreateArgs(parentRows[0].tenantID, parentRows[0].parentID)...).WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while creating tenant parent mapping for tenant with id %s", testID),
		},
		{
			Name: "Error while getting tenant by external ID",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while getting business tenant mapping by external id %s", testExternal),
		},
		{
			Name: "Error while upserting tenant",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( external_tenant ) DO UPDATE SET external_name=EXCLUDED.external_name`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while upserting business tenant mapping for tenant with external id %s", testExternal),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			_, err := tenantMappingRepo.Upsert(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_UnsafeCreate(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)
				for _, row := range parentRows {
					dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
						WithArgs(fixTenantParentCreateArgs(row.tenantID, row.parentID)...).WillReturnResult(driver.ResultNoRows)
				}
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Success without parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModelWithoutParents).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			Input:                *tenantMappingModelWithoutParents,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while creating parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(fixTenantParentCreateArgs(parentRows[0].tenantID, parentRows[0].parentID)...).WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while creating tenant parent mapping for tenant with id %s", testID),
		},
		{
			Name: "Error while getting tenant by external ID",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while getting business tenant mapping by external id %s", testExternal),
		},
		{
			Name: "Error while creating tenant",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", tenantMappingModel).Return(tenantMappingEntity).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, type, provider_name, status ) VALUES ( ?, ?, ?, ?, ?, ? )  ON CONFLICT ( external_tenant ) DO NOTHING`)).
					WithArgs(fixTenantMappingCreateArgs(*tenantMappingEntity)...).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                *tenantMappingModel,
			ExpectedErrorMessage: fmt.Sprintf("while creating business tenant mapping for tenant with external id %s", testExternal),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			_, err := tenantMappingRepo.UnsafeCreate(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_Get(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                string
		ExpectedResult       *model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				return db, dbMock
			},
			Input:          testID,
			ExpectedResult: tenantMappingModel,
		},
		{
			Name: "Success without parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			Input:          testID,
			ExpectedResult: tenantMappingModelWithoutParents,
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                testID,
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while getting tenant",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                testID,
			ExpectedErrorMessage: fmt.Sprintf("while getting tenant with id %s", testID),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.Get(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_GetByExternalTenant(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                string
		ExpectedResult       *model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows(parentRows)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				return db, dbMock
			},
			Input:          testExternal,
			ExpectedResult: tenantMappingModel,
		},
		{
			Name: "Success without parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			Input:          testExternal,
			ExpectedResult: tenantMappingModelWithoutParents,
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(newModelBusinessTenantMapping(testID, testName, nil)).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while getting tenant",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnError(testError)
				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while getting tenant with external id %s", testExternal),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.GetByExternalTenant(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Exists(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.Exists(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsByExternalTenant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE external_tenant = $1`)).
			WithArgs(testExternal).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE external_tenant = $1`)).
			WithArgs(testExternal).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_ExistsSubscribed(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))) AND id = $3)`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel, testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsSubscribed(ctx, testID, selfRegDistinguishLabel)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))) AND id = $3)`)).
			WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel, testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.ExistsSubscribed(ctx, testID, selfRegDistinguishLabel)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_List(t *testing.T) {
	initializedVal := true
	notInitializedVal := false

	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				tenantModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
				}

				tenantEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
				}

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
				mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
				mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
					{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
					{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
				})
				parentRowsForTenant1ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: "id1", parentID: testParentID},
					{tenantID: "id1", parentID: testParentID2},
				})
				parentRowsForTenant2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: "id2", parentID: testParentID2},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
					WithArgs(tenantEntity.Active).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id1").
					WillReturnRows(parentRowsForTenant1ToReturn)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id2").
					WillReturnRows(parentRowsForTenant2ToReturn)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id3").
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, []string{testParentID, testParentID2}),
				newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, []string{testParentID2}),
				newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, []string{}),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tenantModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
				}

				tenantEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
				}

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
				mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
				mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
					{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
					{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
					WithArgs(tenantEntity.Active).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id1").
					WillReturnError(testError)
				return db, dbMock
			},
			ExpectedErrorMessage: "while listing parent tenants for tenant with ID id1",
		},
		{
			Name: "Error while listing parents",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT t.id, t.external_name, t.external_tenant, t.type, t.provider_name, t.status, ld.tenant_id IS NOT NULL AS initialized FROM public.business_tenant_mappings t LEFT JOIN public.label_definitions ld ON t.id=ld.tenant_id WHERE t.status = $1 ORDER BY initialized DESC, t.external_name ASC`)).
					WithArgs(tenantEntity.Active).
					WillReturnError(testError)
				return db, dbMock
			},
			ExpectedErrorMessage: "while listing tenants from DB",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.List(ctx)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.List(ctx)

		// THEN
		require.EqualError(t, err, "while fetching persistence from context: Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListPageBySearchTerm(t *testing.T) {
	searchTerm := "name"
	first := 10
	endCursor := ""
	initializedVal := true
	notInitializedVal := false

	resultingTenantModels := []*model.BusinessTenantMapping{
		newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, []string{testParentID, testParentID2}),
		newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, []string{testParentID2}),
		newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, []string{}),
	}

	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       *model.BusinessTenantMappingPage
		ExpectedErrorMessage string
	}{
		{
			Name: "Success with parents",
			ConverterFn: func() *automock.Converter {
				tenantModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
				}

				tenantEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
				}
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
				mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
				mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
					{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
					{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
				})
				parentRowsForTenant1ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: "id1", parentID: testParentID},
					{tenantID: "id1", parentID: testParentID2},
				})
				parentRowsForTenant2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: "id2", parentID: testParentID2},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4)) ORDER BY external_name LIMIT 10 OFFSET 0`)).
					WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4))`)).
					WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id1").
					WillReturnRows(parentRowsForTenant1ToReturn)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id2").
					WillReturnRows(parentRowsForTenant2ToReturn)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id3").
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))
				return db, dbMock
			},
			ExpectedResult: &model.BusinessTenantMappingPage{
				Data: resultingTenantModels,
				PageInfo: &pagination.Page{
					StartCursor: "",
					EndCursor:   "",
					HasNextPage: false,
				},
				TotalCount: len(resultingTenantModels),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tenantModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal, nil),
					newModelBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal, nil),
				}

				tenantEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues("id1", "name1", &initializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id2", "name2", &notInitializedVal),
					newEntityBusinessTenantMappingWithComputedValues("id3", "name3", &notInitializedVal),
				}
				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantEntities[0]).Return(tenantModels[0]).Once()
				mockConverter.On("FromEntity", tenantEntities[1]).Return(tenantModels[1]).Once()
				mockConverter.On("FromEntity", tenantEntities[2]).Return(tenantModels[2]).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: "id1", name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &initializedVal},
					{sqlRow: sqlRow{id: "id2", name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
					{sqlRow: sqlRow{id: "id3", name: "name3", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: &notInitializedVal},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4)) ORDER BY external_name LIMIT 10 OFFSET 0`)).
					WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4))`)).
					WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs("id1").
					WillReturnError(testError)
				return db, dbMock
			},
			ExpectedErrorMessage: "while listing parent tenants for tenant with ID id1",
		},
		{
			Name: "Error while listing parents",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (status = $1 AND (id::text ILIKE $2 OR external_name ILIKE $3 OR external_tenant ILIKE $4)) ORDER BY external_name LIMIT 10 OFFSET 0`)).
					WithArgs(tenantEntity.Active, "%name%", "%name%", "%name%").
					WillReturnError(testError)
				return db, dbMock
			},
			ExpectedErrorMessage: "while listing tenants from DB",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		searchTerm := "name"
		first := 10
		endCursor := ""
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

		// THEN
		require.EqualError(t, err, "while listing tenants from DB: Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListByExternalTenants(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		initializedVal := true
		tntID := id()
		externalTntID := id()
		tntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal}
		tntEntity := &tenantEntity.Entity{ID: tntModel.ID, ExternalTenant: tntModel.ExternalTenant, Initialized: &initializedVal}

		resultingTntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal, Parents: []string{testParentID}}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tntEntity).Return(tntModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{{sqlRow: sqlRow{id: tntModel.ID, externalTenant: tntModel.ExternalTenant}, initialized: &initializedVal}})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: tntID, parentID: testParentID},
		})

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
			WithArgs(tntModel.ExternalTenant).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(tntID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, []string{tntModel.ExternalTenant})

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []*model.BusinessTenantMapping{resultingTntModel}, result)
	})

	t.Run("Success when high load of tenants is requested", func(t *testing.T) {
		// GIVEN
		initializedVal := true
		tntID := id()
		externalTntID := id()
		tntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal}
		tntEntity := &tenantEntity.Entity{ID: tntModel.ID, ExternalTenant: tntModel.ExternalTenant, Initialized: &initializedVal}

		resultingTntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal, Parents: []string{testParentID}}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tntEntity).Return(tntModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{{sqlRow: sqlRow{id: tntModel.ID, externalTenant: tntModel.ExternalTenant}, initialized: &initializedVal}})
		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: tntID, parentID: testParentID},
		})
		// get first chunk of tenant IDs
		firstChunkIDs := chunkSizedTenantIDs(49999)
		firstChunkIDs = append(firstChunkIDs, tntModel.ExternalTenant)
		firstChunkQuery, firstChunkQueryArgs := buildQueryWithTenantIDs(firstChunkIDs)
		dbMock.ExpectQuery(regexp.QuoteMeta(firstChunkQuery)).
			WithArgs(firstChunkQueryArgs...).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(tntID).
			WillReturnRows(parentRowsToReturn)

		// get second chunk of tenant IDs
		secondChunkIDs := chunkSizedTenantIDs(100)
		secondChunkQuery, secondChunkIDsChunkQueryArgs := buildQueryWithTenantIDs(secondChunkIDs)
		dbMock.ExpectQuery(regexp.QuoteMeta(secondChunkQuery)).
			WithArgs(secondChunkIDsChunkQueryArgs...).
			WillReturnRows(fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{}))

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, append(firstChunkIDs, secondChunkIDs...))

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []*model.BusinessTenantMapping{resultingTntModel}, result)
	})

	t.Run("Error when listing parent tenants", func(t *testing.T) {
		// GIVEN
		initializedVal := true
		tntID := id()
		externalTntID := id()
		tntModel := &model.BusinessTenantMapping{ID: tntID, ExternalTenant: externalTntID, Initialized: &initializedVal}
		tntEntity := &tenantEntity.Entity{ID: tntModel.ID, ExternalTenant: tntModel.ExternalTenant, Initialized: &initializedVal}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tntEntity).Return(tntModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{{sqlRow: sqlRow{id: tntModel.ID, externalTenant: tntModel.ExternalTenant}, initialized: &initializedVal}})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
			WithArgs(tntModel.ExternalTenant).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(tntID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, []string{tntModel.ExternalTenant})

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("while listing parent tenants for tenant with ID %s", tntID))
		require.Nil(t, result)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		externalTenantID := id()

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
			WithArgs(externalTenantID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		result, err := tenantMappingRepo.ListByExternalTenants(ctx, []string{externalTenantID})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
		require.Nil(t, result)
	})

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		_, err := repo.ListByExternalTenants(ctx, []string{id()})

		// THEN
		require.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
	})
}

func TestPgRepository_ListByParentAndType(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID2},
				})

				parentRowsForTestID2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT public.business_tenant_mappings.id, public.business_tenant_mappings.external_name, public.business_tenant_mappings.external_tenant, public.business_tenant_mappings.type, public.business_tenant_mappings.provider_name, public.business_tenant_mappings.status from public.business_tenant_mappings join tenant_parents on public.business_tenant_mappings.id = tenant_parents.tenant_id where tenant_parents.parent_id = $1 and public.business_tenant_mappings.type = $2`)).
					WithArgs(testParentID, tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(parentRowsForTestID2ToReturn)

				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account),
				newModelBusinessTenantMappingWithParentAndType(testID2, "name2", []string{testParentID}, nil, tenantEntity.Account),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				tenantByParentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID2, parentID: testParentID},
				})

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testParentID).
					WillReturnRows(tenantByParentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2) AND type = $3`)).
					WithArgs(testID, testID2, tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while listing tenants by id",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				tenantByParentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID2, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testParentID).
					WillReturnRows(tenantByParentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2) AND type = $3`)).
					WithArgs(testID, testID2, tenantEntity.Account).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing tenants of type %s with ids %v", tenantEntity.Account, []string{testID, testID2}),
		},
		{
			Name: "Error while listing tenant parent records",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testParentID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("wlile listing tenant parent records for parent with id %s", testParentID),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListByParentAndType(ctx, testParentID, tenantEntity.Account)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_ListByType(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID2},
				})

				parentRowsForTestID2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE type = $1`)).
					WithArgs(tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(parentRowsForTestID2ToReturn)

				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account),
				newModelBusinessTenantMappingWithParentAndType(testID2, "name2", []string{testParentID}, nil, tenantEntity.Account),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE type = $1`)).
					WithArgs(tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while listing tenants by type",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE type = $1`)).
					WithArgs(tenantEntity.Account).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing tenants of type %s", tenantEntity.Account),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListByType(ctx, tenantEntity.Account)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		result, err := repo.ListByType(ctx, tenantEntity.Account)

		// THEN
		require.Contains(t, err.Error(), "Internal Server Error: unable to fetch database from context")
		assert.Nil(t, result)
	})
}

func TestPgRepository_ListByIDsAndType(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID2},
				})

				parentRowsForTestID2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2) AND type = $3`)).
					WithArgs(testID, testID2, tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(parentRowsForTestID2ToReturn)

				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account),
				newModelBusinessTenantMappingWithParentAndType(testID2, "name2", []string{testParentID}, nil, tenantEntity.Account),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2) AND type = $3`)).
					WithArgs(testID, testID2, tenantEntity.Account).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while listing tenants by type",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2) AND type = $3`)).
					WithArgs(testID, testID2, tenantEntity.Account).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing tenants of type %s with ids %v", tenantEntity.Account, []string{testID, testID2}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListByIdsAndType(ctx, []string{testID, testID2}, tenantEntity.Account)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_ListByIDs(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID2},
				})

				parentRowsForTestID2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2)`)).
					WithArgs(testID, testID2).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(parentRowsForTestID2ToReturn)

				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account),
				newModelBusinessTenantMappingWithParentAndType(testID2, "name2", []string{testParentID}, nil, tenantEntity.Account),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2)`)).
					WithArgs(testID, testID2).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while listing tenants by type",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id IN ($1, $2)`)).
					WithArgs(testID, testID2).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing tenants with ids %v", []string{testID, testID2}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListByIds(ctx, []string{testID, testID2})

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_ListBySubscribedRuntimesAndApplicationTemplates(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		ExpectedResult       []*model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID2},
				})

				parentRowsForTestID2ToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))`)).
					WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(parentRowsForTestID2ToReturn)

				return db, dbMock
			},
			ExpectedResult: []*model.BusinessTenantMapping{
				newModelBusinessTenantMappingWithParentAndType(testID, "name1", []string{testParentID, testParentID2}, nil, tenantEntity.Account),
				newModelBusinessTenantMappingWithParentAndType(testID2, "name2", []string{testParentID}, nil, tenantEntity.Account),
			},
		},
		{
			Name: "Error while listing parents",
			ConverterFn: func() *automock.Converter {
				tntModels := []*model.BusinessTenantMapping{
					newModelBusinessTenantMappingWithParentAndType(testID, "name1", nil, nil, tenantEntity.Account),
					newModelBusinessTenantMappingWithParentAndType(testID2, "name2", nil, nil, tenantEntity.Account),
				}

				tntEntities := []*tenantEntity.Entity{
					newEntityBusinessTenantMappingWithComputedValues(testID, "name1", boolToPtr(true)),
					newEntityBusinessTenantMappingWithComputedValues(testID2, "name2", boolToPtr(true)),
				}

				mockConverter := &automock.Converter{}
				for i := 0; i < len(tntEntities); i++ {
					mockConverter.On("FromEntity", tntEntities[i]).Return(tntModels[i]).Once()
				}
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRowsWithComputedValues([]sqlRowWithComputedValues{
					{sqlRow: sqlRow{id: testID, name: "name1", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
					{sqlRow: sqlRow{id: testID2, name: "name2", externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active}, initialized: boolToPtr(true)},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))`)).
					WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while listing tenants by type",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE (type = $1 AND (id IN (SELECT tenant_id FROM tenant_runtime_contexts ) OR id IN (SELECT tenant_id FROM tenant_applications WHERE id IN (SELECT id FROM applications WHERE app_template_id IN (SELECT app_template_id FROM labels WHERE key = $2 AND app_template_id IS NOT NULL)))`)).
					WithArgs(tenantEntity.Subaccount, selfRegDistinguishLabel).
					WillReturnError(testError)

				return db, dbMock
			},
			ExpectedErrorMessage: "while listing tenants by label",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			result, err := tenantMappingRepo.ListBySubscribedRuntimesAndApplicationTemplates(ctx, selfRegDistinguishLabel)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}

	t.Run("Error when missing persistence context", func(t *testing.T) {
		// GIVEN
		repo := tenant.NewRepository(nil)
		ctx := context.TODO()

		// WHEN
		result, err := repo.ListBySubscribedRuntimesAndApplicationTemplates(ctx, selfRegDistinguishLabel)

		// THEN
		require.Contains(t, err.Error(), "Internal Server Error: unable to fetch database from context")
		assert.Nil(t, result)
	})
}

func buildQueryWithTenantIDs(ids []string) (string, []driver.Value) {
	argumentValues := make([]driver.Value, 0)
	var sb strings.Builder
	for i, id := range ids {
		argumentValues = append(argumentValues, id)
		sb.WriteString(fmt.Sprintf("$%d", i+1))
		if i < len(ids)-1 {
			sb.WriteString(", ")
		}
	}

	queryFormat := `SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN (%s)`
	query := fmt.Sprintf(queryFormat, sb.String())

	return query, argumentValues
}

func chunkSizedTenantIDs(chunkSize int) []string {
	ids := make([]string, chunkSize)
	for i := 0; i < chunkSize; i++ {
		ids[i] = id()
	}
	return ids
}

func TestPgRepository_Update(t *testing.T) {
	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                *model.BusinessTenantMapping
		ExpectedErrorMessage string
	}{
		{
			Name: "Success - add, remove, keep parent",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID2, testParent2External, testParent2Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID2, testParent2External, testParent2Name, nil, nil, tenantEntity.Account)).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1, $2)`)).
					WithArgs(testParent2External, testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID2, name: testParent2Name, externalTenant: testParent2External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs(testID, testParentID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				tenantAccesses := fixTenantAccesses()

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

					dbMock.ExpectExec(fixDeleteTenantAccessesQuery()).
						WithArgs(testID, testParentID, tenantAccesses[0].ResourceID, tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))

					dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM `)+`(.+)`+regexp.QuoteMeta(` WHERE tenant_id = $1 AND source = $2`)).
						WithArgs(testID, testParentID).
						WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(testID, testParent2Name).WillReturnResult(sqlmock.NewResult(-1, 1))

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))
					dbMock.ExpectExec(fixInsertTenantAccessesQuery()).
						WithArgs(testID, testParentID2, tenantAccesses[0].ResourceID, tenantAccesses[0].Owner).WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				return db, dbMock
			},
			Input: newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
		},
		{
			Name: "Fail while deleting tenant access record granted from parent",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
					WithArgs(testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs(testID, testParentID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				tenantAccesses := fixTenantAccesses()
				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

				dbMock.ExpectExec(fixDeleteTenantAccessesQuery()).
					WithArgs(testID, testParentID, tenantAccesses[0].ResourceID, tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM `)+`(.+)`+regexp.QuoteMeta(` WHERE tenant_id = $1 AND source = $2`)).
					WithArgs(testID, testParentID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while deleting tenant accesses for granted from the old parent %s to tenant %s", testParentID, testID),
		},
		{
			Name: "Fail while deleting tenant access records for parent",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
					WithArgs(testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs(testID, testParentID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

				dbMock.ExpectExec(fixDeleteTenantAccessesQuery()).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while deleting tenant accesses for the old parent %s of the tenant %s", testParentID, testID),
		},
		{
			Name: "Fail while listing resources for removing tenant access for them for parent",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
					WithArgs(testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs(testID, testParentID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while listing tenant access records for tenant with id %s", testID),
		},
		{
			Name: "Fail while deleting tenant parent record",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1)`)).
					WithArgs(testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs(testID, testParentID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while deleting tenant parent record for tenant with ID %s and parent tenant with ID %s", testID, testParentID),
		},
		{
			Name: "Fail while creating new tenant access records",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID2, testParent2External, testParent2Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID2, testParent2External, testParent2Name, nil, nil, tenantEntity.Account)).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1, $2)`)).
					WithArgs(testParent2External, testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID2, name: testParent2Name, externalTenant: testParent2External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				tenantAccesses := fixTenantAccesses()

				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(testID, testParent2Name).WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

				dbMock.ExpectExec(fixInsertTenantAccessesQuery()).
					WithArgs(testID, testParentID2, tenantAccesses[0].ResourceID, tenantAccesses[0].Owner).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while creating tenant acccess record for resource %s for parent %s of tenant %s", "resourceID", testParentID2, testID),
		},
		{
			Name: "Fail while listing resources for adding tenant access for them for parent",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID2, testParent2External, testParent2Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID2, testParent2External, testParent2Name, nil, nil, tenantEntity.Account)).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1, $2)`)).
					WithArgs(testParent2External, testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID2, name: testParent2Name, externalTenant: testParent2External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(testID, testParent2Name).WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while listing tenant access records for tenant with id %s", testID),
		},
		{
			Name: "Fail while creating tenant parent record",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID2, testParent2External, testParent2Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID2, testParent2External, testParent2Name, nil, nil, tenantEntity.Account)).Once()
				mockConverter.On("FromEntity", newEntityBusinessTenantMappingWithExternalID(testParentID3, testParent3External, testParent3Name)).Return(newModelBusinessTenantMappingWithTypeAndExternalID(testParentID3, testParent3External, testParent3Name, nil, nil, tenantEntity.Account)).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1, $2)`)).
					WithArgs(testParent2External, testParent3External).
					WillReturnRows(fixSQLRows([]sqlRow{
						{id: testParentID2, name: testParent2Name, externalTenant: testParent2External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
						{id: testParentID3, name: testParent3Name, externalTenant: testParent3External, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
					}))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testParentID3).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs(testID, testParent2Name).WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while adding tenant parent record for tenant with ID %s and parent teannt with ID %s", testID, testParentID2),
		},
		{
			Name: "Fail while listing parent tenants by external ID",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant IN ($1, $2)`)).
					WithArgs(testParent2External, testParent3External).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while listing parent tenants by external IDs %v", []string{testParent2External, testParent3External}),
		},
		{
			Name: "Fail while updating tenant",
			ConverterFn: func() *automock.Converter {
				oldTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, nil, nil, tenantEntity.Account)
				newTenantMappingModel := newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account)
				oldTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)
				newTenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("ToEntity", newTenantMappingModel).Return(newTenantMappingEntity).Once()
				mockConverter.On("FromEntity", oldTenantMappingEntity).Return(oldTenantMappingModel).Once()

				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
					{tenantID: testID, parentID: testParentID3},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.business_tenant_mappings SET external_name = ?, external_tenant = ?, type = ?, provider_name = ?, status = ? WHERE id = ? `)).
					WithArgs(testName, testExternal, "account", "Compass", tenantEntity.Active, testID).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while updating tenant with ID %s", testID),
		},
		{
			Name: "Fail while getting tenant",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE id = $1 AND status != $2 `)).
					WithArgs(testID, tenantEntity.Inactive).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                newModelBusinessTenantMappingWithType(testID, testName, []string{testParent2External, testParent3External}, nil, tenantEntity.Account),
			ExpectedErrorMessage: fmt.Sprintf("while getting tenant with ID %s", testID),
		},
		{
			Name:                 "Fail when model is empty",
			Input:                nil,
			ExpectedErrorMessage: "model can not be empty",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			err := tenantMappingRepo.Update(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_DeleteByExternalTenant(t *testing.T) {
	deleteStatement := regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE external_tenant = $1`)

	testCases := []struct {
		Name                 string
		ConverterFn          func() *automock.Converter
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		Input                string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					tenantAccesses := fixTenantAccesses()

					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

					dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
						WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				parentRowsToReturn = fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID2, parentID: testID},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE id = $1`)).
					WithArgs(testID2).WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectExec(deleteStatement).
					WithArgs(testExternal).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				return db, dbMock
			},
			Input: testExternal,
		},
		{
			Name: "Success when getting tenant by external ID returnes not found",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).WillReturnError(sql.ErrNoRows)

				return db, dbMock
			},
			Input: testExternal,
		},
		{
			Name: "Error while deleting tenant",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					tenantAccesses := fixTenantAccesses()

					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

					dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
						WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				parentRowsToReturn = fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID2, parentID: testID},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE id = $1`)).
					WithArgs(testID2).WillReturnResult(sqlmock.NewResult(-1, 1))

				dbMock.ExpectExec(deleteStatement).
					WithArgs(testExternal).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
		{
			Name: "Error while deleting child tenant",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					tenantAccesses := fixTenantAccesses()

					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

					dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
						WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				parentRowsToReturn = fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID2, parentID: testID},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID2).
					WillReturnRows(sqlmock.NewRows(testTenantParentsTableColumns))

				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.business_tenant_mappings WHERE id = $1`)).
					WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while deleting tenant with ID %s", testID2),
		},
		{
			Name: "Error while listing child tenants",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				for topLvlEntity := range resource.TopLevelEntities {
					if _, ok := topLvlEntity.IgnoredTenantAccessTable(); ok {
						continue
					}
					tenantAccesses := fixTenantAccesses()

					dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
						WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

					dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
						WithArgs(tenantAccesses[0].ResourceID).WillReturnResult(sqlmock.NewResult(-1, 1))
				}

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs(testID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while listing child tenants for tenant with ID %s", testID),
		},
		{
			Name: "Error while deleting owned resources",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				tenantAccesses := fixTenantAccesses()

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID, true).WillReturnRows(sqlmock.NewRows(repo.M2MColumns).AddRow(fixTenantAccessesRow()...))

				dbMock.ExpectExec(`DELETE FROM (.+) WHERE id IN \(\$1\)`).
					WithArgs(tenantAccesses[0].ResourceID).WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while deleting resources owned by tenant %s", testID),
		},
		{
			Name: "Error while listing owned resources",
			ConverterFn: func() *automock.Converter {
				tenantMappingModel := newModelBusinessTenantMapping(testID, testName, nil)
				tenantMappingEntity := newEntityBusinessTenantMapping(testID, testName)

				mockConverter := &automock.Converter{}
				mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
				return mockConverter
			},
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rowsToReturn := fixSQLRows([]sqlRow{
					{id: testID, name: testName, externalTenant: testExternal, typeRow: string(tenantEntity.Account), provider: "Compass", status: tenantEntity.Active},
				})
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).
					WillReturnRows(rowsToReturn)

				parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
					{tenantID: testID, parentID: testParentID},
				})

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs(testID).
					WillReturnRows(parentRowsToReturn)

				dbMock.ExpectQuery(`SELECT tenant_id, id, owner, source FROM (.+) WHERE tenant_id = \$1`).
					WithArgs(testID, true).WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while listing tenant access records for tenant with id %s", testID),
		},
		{
			Name: "Error while listing owned resources",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, external_name, external_tenant, type, provider_name, status FROM public.business_tenant_mappings WHERE external_tenant = $1 AND status != $2 `)).
					WithArgs(testExternal, tenantEntity.Inactive).WillReturnError(testError)

				return db, dbMock
			},
			Input:                testExternal,
			ExpectedErrorMessage: fmt.Sprintf("while getting tenant with external ID %s", testExternal),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := &automock.Converter{}
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantMappingRepo := tenant.NewRepository(converter)

			err := tenantMappingRepo.DeleteByExternalTenant(ctx, testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
			mock.AssertExpectationsForObjects(t, converter)
		})
	}
}

func TestPgRepository_GetLowestOwnerForResource(t *testing.T) {
	runtimeID := "runtimeID"

	t.Run("Success", func(t *testing.T) {
		db, dbMock := mockDBSuccess(t, runtimeID)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.GetLowestOwnerForResource(ctx, resource.Runtime, runtimeID)

		// THEN
		require.NoError(t, err)
		require.Equal(t, testID, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		db, dbMock := mockDBError(t, runtimeID)
		defer dbMock.AssertExpectations(t)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		result, err := tenantMappingRepo.GetLowestOwnerForResource(ctx, resource.Runtime, runtimeID)

		// THEN
		require.Error(t, err)
		require.Empty(t, result)
	})
}

func TestPgRepository_GetParentRecursivelyByExternalTenant(t *testing.T) {

	dbQuery := `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.external_name, t1.external_tenant, t1.provider_name, t1.status, t1.type, tp1.parent_id, 0 AS depth
                    FROM business_tenant_mappings t1 JOIN tenant_parents tp1 on t1.id = tp1.tenant_id
                    WHERE external_tenant = $1
                    UNION ALL
                    SELECT t2.id, t2.external_name, t2.external_tenant, t2.provider_name, t2.status, t2.type, tp2.parent_id, p.depth+ 1
                    FROM business_tenant_mappings t2 LEFT JOIN tenant_parents tp2 on t2.id = tp2.tenant_id
                                                     INNER JOIN parents p on p.parent_id = t2.id)
			SELECT id, external_name, external_tenant, provider_name, status, type FROM parents WHERE parent_id is NULL AND (type != 'cost-object' OR (type = 'cost-object' AND depth = (SELECT MIN(depth) FROM parents WHERE type = 'cost-object')))`

	tenantMappingModel := &model.BusinessTenantMapping{
		ID:             testID,
		Name:           testName,
		ExternalTenant: testExternal,
		Parents:        []string{},
		Type:           tenantEntity.Account,
		Provider:       testProvider,
		Status:         tenantEntity.Active,
	}
	tenantMappingEntity := &tenantEntity.Entity{
		ID:             testID,
		Name:           testName,
		ExternalTenant: testExternal,
		Type:           tenantEntity.Account,
		ProviderName:   testProvider,
		Status:         tenantEntity.Active,
	}

	t.Run("Success when parent and type are returned", func(t *testing.T) {
		// GIVEN

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"id", "external_name", "external_tenant", "provider_name", "status", "type"}).AddRow(testID, testName, testExternal, testProvider, tenantEntity.Active, tenantEntity.TypeToStr(tenantEntity.Account))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testExternal).
			WillReturnRows(rowsToReturn)

		parentRowsToReturn := fixSQLTenantParentsRows([]sqlTenantParentsRow{
			{tenantID: testID, parentID: testParentID},
		})

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnRows(parentRowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		parentTenant, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)

		// THEN
		require.NoError(t, err)
		require.Equal(t, []*model.BusinessTenantMapping{tenantMappingModel}, parentTenant)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when listing parents", func(t *testing.T) {
		// GIVEN

		db, dbMock := testdb.MockDatabase(t)

		rowsToReturn := sqlmock.NewRows([]string{"id", "external_name", "external_tenant", "provider_name", "status", "type"}).AddRow(testID, testName, testExternal, testProvider, tenantEntity.Active, tenantEntity.TypeToStr(tenantEntity.Account))
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testExternal).
			WillReturnRows(rowsToReturn)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantMappingEntity).Return(tenantMappingModel).Once()
		tenantMappingRepo := tenant.NewRepository(mockConverter)

		// WHEN
		parentTenant, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("while listing parent tenants for tenant with ID %s", testID))
		require.Nil(t, parentTenant)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error when executing db query", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(dbQuery)).
			WithArgs(testExternal).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		tenantMappingRepo := tenant.NewRepository(nil)

		// WHEN
		parentTenant, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		require.Empty(t, parentTenant)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error if missing persistence context", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		tenantMappingRepo := tenant.NewRepository(nil)
		// WHEN
		_, err := tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, testExternal)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

const selectTenantsQuery = `(SELECT tenant_id FROM tenant_runtimes ta WHERE ta.id = $1 AND ta.owner = true AND (NOT EXISTS(SELECT 1 FROM public.business_tenant_mappings JOIN tenant_parents ON public.business_tenant_mappings.id = tenant_parents.tenant_id WHERE parent_id = ta.tenant_id) OR (NOT EXISTS(SELECT 1 FROM tenant_runtimes ta2 WHERE ta2.id = $2 AND ta2.owner = true AND ta2.tenant_id IN (SELECT id FROM public.business_tenant_mappings JOIN tenant_parents ON public.business_tenant_mappings.id = tenant_parents.tenant_id WHERE parent_id = ta.tenant_id)))))`

func mockDBSuccess(t *testing.T, runtimeID string) (*sqlx.DB, testdb.DBMock) {
	db, dbMock := testdb.MockDatabase(t)
	rowsToReturn := sqlmock.NewRows([]string{"tenant_id"}).AddRow(testID)
	dbMock.ExpectQuery(regexp.QuoteMeta(selectTenantsQuery)).
		WithArgs(runtimeID, runtimeID).
		WillReturnRows(rowsToReturn)
	return db, dbMock
}

func mockDBError(t *testing.T, runtimeID string) (*sqlx.DB, testdb.DBMock) {
	db, dbMock := testdb.MockDatabase(t)
	dbMock.ExpectQuery(regexp.QuoteMeta(selectTenantsQuery)).
		WithArgs(runtimeID, runtimeID).WillReturnError(testError)
	return db, dbMock
}

func id() string {
	return uuid.New().String()
}
