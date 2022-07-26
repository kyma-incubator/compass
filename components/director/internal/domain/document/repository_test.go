package document_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

var columns = []string{"id", "bundle_id", "app_id", "title", "display_name", "description", "format", "kind", "data", "ready", "created_at", "updated_at", "deleted_at", "error"}

func TestRepository_Create(t *testing.T) {
	refID := bndlID()
	var nilDocModel *model.Document
	docModel := fixModelDocument(givenID(), refID)
	docEntity := fixEntityDocument(givenID(), refID)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Document",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM bundles_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{givenTenant(), refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query: regexp.QuoteMeta("INSERT INTO public.documents ( id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args: []driver.Value{givenID(), refID, appID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data,
					docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		ModelEntity:         docModel,
		DBEntity:            docEntity,
		NilModelEntity:      nilDocModel,
		TenantID:            givenTenant(),
	}

	suite.Run(t)
}

func TestRepository_CreateMany(t *testing.T) {
	parentAccessQuery := regexp.QuoteMeta("SELECT 1 FROM bundles_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3")
	insertQuery := regexp.QuoteMeta("INSERT INTO public.documents ( id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )")

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
			fixEntityDocument("3", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		for i, givenModel := range given {
			expectedEntity := expected[i]
			conv.On("ToEntity", givenModel).Return(expectedEntity, nil).Once()
			dbMock.ExpectQuery(parentAccessQuery).WithArgs(givenTenant(), bndlID(), true).WillReturnRows(testdb.RowWhenObjectExist())
			dbMock.ExpectExec(insertQuery).
				WithArgs(givenModel.ID, bndlID(), appID, givenModel.Title, givenModel.DisplayName, givenModel.Description, givenModel.Format, givenModel.Kind, givenModel.Data,
					givenModel.Ready, givenModel.CreatedAt, givenModel.UpdatedAt, givenModel.DeletedAt, givenModel.Error).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)
		// WHEN
		err := repo.CreateMany(ctx, givenTenant(), given)
		// THEN
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", given[0]).Return(expected[0], nil).Once()
		conv.On("ToEntity", given[1]).Return(expected[1], nil).Once()

		dbMock.ExpectQuery(parentAccessQuery).WithArgs(givenTenant(), bndlID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(insertQuery).
			WithArgs(expected[0].ID, bndlID(), appID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data,
				expected[0].Ready, expected[0].CreatedAt, expected[0].UpdatedAt, expected[0].DeletedAt, expected[0].Error).WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectQuery(parentAccessQuery).WithArgs(givenTenant(), bndlID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(insertQuery).
			WithArgs(expected[1].ID, bndlID(), appID, expected[1].Title, expected[1].DisplayName, expected[1].Description, expected[1].Format, expected[1].Kind, expected[1].Data,
				expected[1].Ready, expected[1].CreatedAt, expected[1].UpdatedAt, expected[1].DeletedAt, expected[1].Error).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, givenTenant(), given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", bndlID()),
			fixModelDocument("2", bndlID()),
			fixModelDocument("3", bndlID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", bndlID()),
			fixEntityDocument("2", bndlID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		//defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", given[0]).Return(expected[0], nil).Once()
		conv.On("ToEntity", given[1]).Return(nil, givenError()).Once()
		dbMock.ExpectQuery(parentAccessQuery).WithArgs(givenTenant(), bndlID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(insertQuery).
			WithArgs(expected[0].ID, bndlID(), appID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data,
				expected[0].Ready, expected[0].CreatedAt, expected[0].UpdatedAt, expected[0].DeletedAt, expected[0].Error).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, givenTenant(), given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: while creating Document entity from model: some error")
	})
}

func TestRepository_ListAllForBundle(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageBundleID := "emptyPageBundleID"

	onePageBundleID := "onePageBundleID"
	firstDocID := "111111111-1111-1111-1111-111111111111"
	firstDocEntity := fixEntityDocument(firstDocID, onePageBundleID)
	firstDocModel := fixModelDocument(firstDocID, onePageBundleID)

	multiplePagesBundleID := "multiplePagesBundleID"

	secondDocID := "222222222-2222-2222-2222-222222222222"
	secondDocEntity := fixEntityDocument(secondDocID, multiplePagesBundleID)
	secondDocModel := fixModelDocument(secondDocID, multiplePagesBundleID)

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Documents for multiple bundles with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $1)) AND bundle_id = $2 ORDER BY bundle_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $5)) AND bundle_id = $6 ORDER BY bundle_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $9)) AND bundle_id = $10 ORDER BY bundle_id ASC, id ASC LIMIT $11 OFFSET $12)`),

				Args:     []driver.Value{givenTenant(), emptyPageBundleID, pageSize, 0, givenTenant(), onePageBundleID, pageSize, 0, givenTenant(), multiplePagesBundleID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(columns).
						AddRow(firstDocID, firstDocEntity.BndlID, firstDocEntity.AppID, firstDocEntity.Title, firstDocEntity.DisplayName, firstDocEntity.Description, firstDocEntity.Format, firstDocEntity.Kind, firstDocEntity.Data, firstDocEntity.Ready, firstDocEntity.CreatedAt, firstDocEntity.UpdatedAt, firstDocEntity.DeletedAt, firstDocEntity.Error).
						AddRow(secondDocID, secondDocEntity.BndlID, secondDocEntity.AppID, secondDocEntity.Title, secondDocEntity.DisplayName, secondDocEntity.Description, secondDocEntity.Format, secondDocEntity.Kind, secondDocEntity.Data, secondDocEntity.Ready, secondDocEntity.CreatedAt, secondDocEntity.UpdatedAt, secondDocEntity.DeletedAt, secondDocEntity.Error),
					}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT bundle_id AS id, COUNT(*) AS total_count FROM public.documents WHERE (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $1)) GROUP BY bundle_id ORDER BY bundle_id ASC`),
				Args:     []driver.Value{givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(emptyPageBundleID, 0).AddRow(onePageBundleID, 1).AddRow(multiplePagesBundleID, 2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.DocumentPage{
					Data: nil,
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 0,
				},
			},
			{
				ExpectedModelEntities: []interface{}{firstDocModel},
				ExpectedDBEntities:    []interface{}{firstDocEntity},
				ExpectedPage: &model.DocumentPage{
					Data: []*model.Document{firstDocModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{secondDocModel},
				ExpectedDBEntities:    []interface{}{secondDocEntity},
				ExpectedPage: &model.DocumentPage{
					Data: []*model.Document{secondDocModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   pagination.EncodeNextOffsetCursor(0, pageSize),
						HasNextPage: true,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		MethodName:          "ListByBundleIDs",
		MethodArgs:          []interface{}{givenTenant(), []string{emptyPageBundleID, onePageBundleID, multiplePagesBundleID}, pageSize, cursor},
	}

	suite.Run(t)
}

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Document Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.documents WHERE id = $1 AND (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		TargetID:            givenID(),
		TenantID:            givenTenant(),
		MethodName:          "Exists",
		MethodArgs:          []interface{}{givenTenant(), givenID()},
	}

	suite.Run(t)
}

func TestRepository_GetByID(t *testing.T) {
	docModel := fixModelDocument(givenID(), bndlID())
	docEntity := fixEntityDocument(givenID(), bndlID())

	suite := testdb.RepoGetTestSuite{
		Name: "Get Document",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE id = $1 AND (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(columns).
							AddRow(givenID(), docEntity.BndlID, docEntity.AppID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data, docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(columns),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		ExpectedModelEntity: docModel,
		ExpectedDBEntity:    docEntity,
		MethodArgs:          []interface{}{givenTenant(), givenID()},
	}

	suite.Run(t)
}

func TestRepository_GetForBundle(t *testing.T) {
	docModel := fixModelDocument(givenID(), bndlID())
	docEntity := fixEntityDocument(givenID(), bndlID())

	suite := testdb.RepoGetTestSuite{
		Name: "Get Document For Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, bundle_id, app_id, title, display_name, description, format, kind, data, ready, created_at, updated_at, deleted_at, error FROM public.documents WHERE id = $1 AND bundle_id = $2 AND (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{givenID(), bndlID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(columns).
							AddRow(givenID(), docEntity.BndlID, docEntity.AppID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data, docEntity.Ready, docEntity.CreatedAt, docEntity.UpdatedAt, docEntity.DeletedAt, docEntity.Error),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(columns),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		ExpectedModelEntity: docModel,
		ExpectedDBEntity:    docEntity,
		MethodArgs:          []interface{}{givenTenant(), givenID(), bndlID()},
		MethodName:          "GetForBundle",
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Document Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.documents WHERE id = $1 AND (id IN (SELECT id FROM documents_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenID(), givenTenant()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: document.NewRepository,
		MethodArgs:          []interface{}{givenTenant(), givenID()},
	}

	suite.Run(t)
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func bndlID() string {
	return "ppppppppp-pppp-pppp-pppp-pppppppppppp"
}

func givenTenant() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func givenError() error {
	return errors.New("some error")
}
