package document_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const insertQuery = "INSERT INTO public.documents ( id, tenant_id, package_id, title, display_name, description, format, kind, data ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )"

var columns = []string{"id", "tenant_id", "package_id", "title", "display_name", "description", "format", "kind", "data"}

func TestRepository_Create(t *testing.T) {
	refID := pkgID()
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", *docModel).Return(*docEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, docModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", *docModel).Return(*docEntity, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(ctx, docModel)
		// THEN
		require.EqualError(t, err, "while inserting row to 'public.documents' table: some error")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", *docModel).Return(document.Entity{}, givenError())

		repo := document.NewRepository(mockConverter)
		// WHEN
		err := repo.Create(context.TODO(), docModel)
		// THEN
		require.EqualError(t, err, "while creating Document entity from model: some error")
	})
}

func TestRepository_CreateMany(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", pkgID()),
			fixModelDocument("2", pkgID()),
			fixModelDocument("3", pkgID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", pkgID()),
			fixEntityDocument("2", pkgID()),
			fixEntityDocument("3", pkgID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		for i, givenModel := range given {
			expectedEntity := expected[i]
			conv.On("ToEntity", *givenModel).Return(*expectedEntity, nil).Once()
			dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
				expectedEntity.ID, expectedEntity.TenantID, expectedEntity.PkgID, expectedEntity.Title, expectedEntity.DisplayName, expectedEntity.Description, expectedEntity.Format, expectedEntity.Kind, expectedEntity.Data).WillReturnResult(sqlmock.NewResult(-1, 1))
		}

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)
		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", pkgID()),
			fixModelDocument("2", pkgID()),
			fixModelDocument("3", pkgID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", pkgID()),
			fixEntityDocument("2", pkgID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", *given[0]).Return(*expected[0], nil).Once()
		conv.On("ToEntity", *given[1]).Return(*expected[1], nil).Once()
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[0].ID, expected[0].TenantID, expected[0].PkgID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[1].ID, expected[1].TenantID, expected[1].PkgID, expected[1].Title, expected[1].DisplayName, expected[1].Description, expected[1].Format, expected[1].Kind, expected[1].Data).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: while inserting row to 'public.documents' table: some error")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		given := []*model.Document{
			fixModelDocument("1", pkgID()),
			fixModelDocument("2", pkgID()),
			fixModelDocument("3", pkgID()),
		}
		expected := []*document.Entity{
			fixEntityDocument("1", pkgID()),
			fixEntityDocument("2", pkgID()),
		}

		db, dbMock := testdb.MockDatabase(t)
		//defer dbMock.AssertExpectations(t)

		conv.On("ToEntity", *given[0]).Return(*expected[0], nil).Once()
		conv.On("ToEntity", *given[1]).Return(document.Entity{}, givenError()).Once()
		dbMock.ExpectExec(regexp.QuoteMeta(insertQuery)).WithArgs(
			expected[0].ID, expected[0].TenantID, expected[0].PkgID, expected[0].Title, expected[0].DisplayName, expected[0].Description, expected[0].Format, expected[0].Kind, expected[0].Data).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(conv)

		// WHEN
		err := repo.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating Document with ID 2: while creating Document entity from model: some error")
	})
}

func TestRepository_ListForPackage(t *testing.T) {
	// GIVEN
	tenantID := "tnt"
	ExpectedLimit := 3
	ExpectedOffset := 0
	testErr := errors.New("Test error")

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	docEntity1 := fixEntityDocument("1", pkgID())
	docEntity2 := fixEntityDocument("2", pkgID())

	selectQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, package_id, title, display_name, description, format, kind, data
		FROM public.documents WHERE tenant_id=$1 AND package_id = '%s' ORDER BY id LIMIT %d OFFSET %d`, pkgID(), ExpectedLimit, ExpectedOffset))

	rawCountQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.documents WHERE tenant_id=$1 AND package_id = '%s'`, pkgID())
	countQuery := regexp.QuoteMeta(rawCountQuery)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(docEntity1.ID, docEntity1.TenantID, docEntity1.PkgID, docEntity1.Title, docEntity1.DisplayName, docEntity1.Description, docEntity1.Format, docEntity1.Kind, docEntity1.Data).
			AddRow(docEntity2.ID, docEntity2.TenantID, docEntity2.PkgID, docEntity2.Title, docEntity2.DisplayName, docEntity2.Description, docEntity2.Format, docEntity2.Kind, docEntity2.Data)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		conv.On("FromEntity", *docEntity1).Return(model.Document{ID: docEntity1.ID}, nil).Once()
		conv.On("FromEntity", *docEntity2).Return(model.Document{ID: docEntity2.ID}, nil).Once()

		pgRepository := document.NewRepository(conv)
		// WHEN
		modelAPIDef, err := pgRepository.ListForPackage(ctx, tenantID, pkgID(), inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef.Data, 2)
		assert.Equal(t, docEntity1.ID, modelAPIDef.Data[0].ID)
		assert.Equal(t, docEntity2.ID, modelAPIDef.Data[1].ID)
		assert.Equal(t, "", modelAPIDef.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelAPIDef.TotalCount)
	})

	t.Run("DB Error", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnError(testErr)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.Converter{}
		defer conv.AssertExpectations(t)

		pgRepository := document.NewRepository(conv)
		// WHEN
		_, err := pgRepository.ListForPackage(ctx, tenantID, pkgID(), 3, "")
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Converter Error", func(t *testing.T) {
		testErr := errors.New("test error")
		rows := sqlmock.NewRows(columns).
			AddRow(docEntity1.ID, docEntity1.TenantID, docEntity1.PkgID, docEntity1.Title, docEntity1.DisplayName, docEntity1.Description, docEntity1.Format, docEntity1.Kind, docEntity1.Data).
			AddRow(docEntity2.ID, docEntity2.TenantID, docEntity2.PkgID, docEntity2.Title, docEntity2.DisplayName, docEntity2.Description, docEntity2.Format, docEntity2.Kind, docEntity2.Data)

		conv := &automock.Converter{}
		conv.On("FromEntity", *docEntity1).Return(model.Document{}, testErr).Once()
		defer conv.AssertExpectations(t)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		repo := document.NewRepository(conv)
		//WHEN
		_, err := repo.ListForPackage(ctx, tenantID, pkgID(), inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
	})
}

func TestRepository_Exists(t *testing.T) {
	// given
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM public.documents WHERE tenant_id = $1 AND id = $2")).WithArgs(
		givenTenant(), givenID()).
		WillReturnRows(testdb.RowWhenObjectExist())

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	repo := document.NewRepository(nil)

	// when
	ex, err := repo.Exists(ctx, givenTenant(), givenID())

	// then
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRepository_GetByID(t *testing.T) {
	refID := pkgID()

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", *docEntity).Return(*docModel, nil).Once()

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data)

		query := "SELECT id, tenant_id, package_id, title, display_name, description, format, kind, data FROM public.documents WHERE tenant_id = $1 AND id = $2"
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, docModel, actual)

		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", *docEntity).Return(model.Document{}, givenError())

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while converting Document entity to model: some error")
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		repo := document.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while getting object from DB: some error")
	})
}

func TestRepository_GetForPackage(t *testing.T) {
	refID := pkgID()

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		docModel := fixModelDocument(givenID(), refID)
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", *docEntity).Return(*docModel, nil).Once()

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data)

		query := "SELECT id, tenant_id, package_id, title, display_name, description, format, kind, data FROM public.documents WHERE tenant_id = $1 AND id = $2 AND package_id = $3"
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(givenTenant(), givenID(), pkgID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := repo.GetForPackage(ctx, givenTenant(), givenID(), pkgID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, docModel, actual)

		mockConverter.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		docEntity := fixEntityDocument(givenID(), refID)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", *docEntity).Return(model.Document{}, givenError())

		repo := document.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(columns).
			AddRow(givenID(), givenTenant(), refID, docEntity.Title, docEntity.DisplayName, docEntity.Description, docEntity.Format, docEntity.Kind, docEntity.Data)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID(), pkgID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetForPackage(ctx, givenTenant(), givenID(), pkgID())
		// THEN
		require.EqualError(t, err, "while converting Document entity to model: some error")
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
		repo := document.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID(), pkgID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := repo.GetForPackage(ctx, givenTenant(), givenID(), pkgID())
		// THEN
		require.EqualError(t, err, "while getting object from DB: some error")
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.documents WHERE tenant_id = $1 AND id = $2")).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := document.NewRepository(nil)
		// WHEN
		err := repo.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while deleting from database: some error")
	})
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func appID() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}

func pkgID() string {
	return "ppppppppp-pppp-pppp-pppp-pppppppppppp"
}

func givenTenant() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func givenError() error {
	return errors.New("some error")
}
