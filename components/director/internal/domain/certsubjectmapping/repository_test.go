package certsubjectmapping_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

func TestRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name:       "Create certificate subject mapping",
		MethodName: "Create",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.cert_subject_mapping \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{CertSubjectMappingEntity.ID, CertSubjectMappingEntity.Subject, CertSubjectMappingEntity.ConsumerType, CertSubjectMappingEntity.InternalConsumerID, CertSubjectMappingEntity.TenantAccessLevels},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       certsubjectmapping.NewRepository,
		ModelEntity:               CertSubjectMappingModel,
		DBEntity:                  CertSubjectMappingEntity,
		NilModelEntity:            nilModelEntity,
		IsGlobal:                  true,
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Get(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get certificate subject mapping by ID",
		MethodName: "Get",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, subject, consumer_type, internal_consumer_id, tenant_access_levels FROM public.cert_subject_mapping WHERE id = $1`),
				Args:     []driver.Value{TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(CertSubjectMappingEntity.ID, CertSubjectMappingEntity.Subject, CertSubjectMappingEntity.ConsumerType, CertSubjectMappingEntity.InternalConsumerID, CertSubjectMappingEntity.TenantAccessLevels)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       certsubjectmapping.NewRepository,
		ExpectedModelEntity:       CertSubjectMappingModel,
		ExpectedDBEntity:          CertSubjectMappingEntity,
		MethodArgs:                []interface{}{TestID},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_GetBySubject(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name:       "Get certificate subject mapping by ID",
		MethodName: "GetBySubject",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, subject, consumer_type, internal_consumer_id, tenant_access_levels FROM public.cert_subject_mapping WHERE subject ILIKE $1`),
				Args:     []driver.Value{"C=DE_ L=test_ O=SAP SE_ OU=TestRegion_ OU=SAP Cloud Platform Clients_ OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10_ CN=test-compass"},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(CertSubjectMappingEntity.ID, CertSubjectMappingEntity.Subject, CertSubjectMappingEntity.ConsumerType, CertSubjectMappingEntity.InternalConsumerID, CertSubjectMappingEntity.TenantAccessLevels)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       certsubjectmapping.NewRepository,
		ExpectedModelEntity:       CertSubjectMappingModel,
		ExpectedDBEntity:          CertSubjectMappingEntity,
		MethodArgs:                []interface{}{TestSubject},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.cert_subject_mapping SET subject = ?, consumer_type = ?, internal_consumer_id = ?, tenant_access_levels = ? WHERE id = ?`)
	suite := testdb.RepoUpdateTestSuite{
		Name: "Update certificate subject mapping by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{CertSubjectMappingEntity.Subject, CertSubjectMappingEntity.ConsumerType, CertSubjectMappingEntity.InternalConsumerID, CertSubjectMappingEntity.TenantAccessLevels, CertSubjectMappingEntity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		RepoConstructorFunc: certsubjectmapping.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		ModelEntity:    CertSubjectMappingModel,
		DBEntity:       CertSubjectMappingEntity,
		NilModelEntity: nilModelEntity,
		IsGlobal:       true,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete certificate subject mapping by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.cert_subject_mapping WHERE id = $1`),
				Args:          []driver.Value{TestID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		RepoConstructorFunc: certsubjectmapping.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		IsGlobal:   true,
		MethodArgs: []interface{}{TestID},
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists certificate subject mapping by ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.cert_subject_mapping WHERE id = $1`),
				Args:     []driver.Value{TestID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: certsubjectmapping.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   TestID,
		IsGlobal:   true,
		MethodName: "Exists",
		MethodArgs: []interface{}{TestID},
	}

	suite.Run(t)
}

func TestRepository_ExistsBySubject(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Exists certificate subject mapping by Subject",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.cert_subject_mapping WHERE subject ILIKE $1`),
				Args:     []driver.Value{"C=DE_ L=test_ O=SAP SE_ OU=TestRegion_ OU=SAP Cloud Platform Clients_ OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10_ CN=test-compass"},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		RepoConstructorFunc: certsubjectmapping.NewRepository,
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		TargetID:   TestID,
		IsGlobal:   true,
		MethodName: "ExistsBySubject",
		MethodArgs: []interface{}{TestSubject},
	}

	suite.Run(t)
}

func TestRepository_List(t *testing.T) {
	suite := testdb.RepoListPageableTestSuite{
		Name:       "List certificate subject mappings with paging",
		MethodName: "List",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, subject, consumer_type, internal_consumer_id, tenant_access_levels FROM public.cert_subject_mapping ORDER BY id LIMIT 3 OFFSET 0`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns()).AddRow(CertSubjectMappingEntity.ID, CertSubjectMappingEntity.Subject, CertSubjectMappingEntity.ConsumerType, CertSubjectMappingEntity.InternalConsumerID, CertSubjectMappingEntity.TenantAccessLevels)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns())}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT COUNT(*) FROM public.cert_subject_mapping`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(1)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{CertSubjectMappingModel},
				ExpectedDBEntities:    []interface{}{CertSubjectMappingEntity},
				ExpectedPage: &model.CertSubjectMappingPage{
					Data: []*model.CertSubjectMapping{CertSubjectMappingModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       certsubjectmapping.NewRepository,
		MethodArgs:                []interface{}{3, ""},
		DisableConverterErrorTest: false,
	}

	suite.Run(t)
}
