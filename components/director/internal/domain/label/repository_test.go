package label_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	var nilLabelModel *model.Label
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create Application Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
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
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{appLabelEntity.ID, appLabelEntity.TenantID, appLabelEntity.AppID, appLabelEntity.RuntimeID, appLabelEntity.RuntimeContextID, appLabelEntity.AppTemplateID, appLabelEntity.Key, appLabelEntity.Value, appLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         applabelModel,
		DBEntity:            appLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
	}

	runtimeLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create Runtime Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_runtimes WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
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
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{runtimeLabelEntity.ID, runtimeLabelEntity.TenantID, runtimeLabelEntity.AppID, runtimeLabelEntity.RuntimeID, runtimeLabelEntity.RuntimeContextID, runtimeLabelEntity.AppTemplateID, runtimeLabelEntity.Key, runtimeLabelEntity.Value, runtimeLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimelabelModel,
		DBEntity:            runtimeLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
	}

	runtimeCtxLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create RuntimeCtx Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_runtime_contexts WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
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
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{runtimeCtxLabelEntity.ID, runtimeCtxLabelEntity.TenantID, runtimeCtxLabelEntity.AppID, runtimeCtxLabelEntity.RuntimeID, runtimeCtxLabelEntity.RuntimeContextID, runtimeCtxLabelEntity.AppTemplateID, runtimeCtxLabelEntity.Key, runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimeCtxlabelModel,
		DBEntity:            runtimeCtxLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		IsTopLevelEntity:    true,
	}

	appLabelSuite.Run(t)
	runtimeLabelSuite.Run(t)
	runtimeCtxLabelSuite.Run(t)

	// Additional tests - tenant labels are created globally as the tenant is embedded in the entity.
	t.Run("Success create - Label for Tenant", func(t *testing.T) {
		// GIVEN
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Create(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Application Template", func(t *testing.T) {
		// GIVEN
		appTemplateLabelModel := fixModelLabel(model.AppTemplateLabelableObject)
		appTemplateLabelEntity := fixEntityLabel(model.AppTemplateLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", appTemplateLabelModel).Return(appTemplateLabelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(appTemplateLabelEntity.ID, appTemplateLabelEntity.TenantID, appTemplateLabelEntity.AppID, appTemplateLabelEntity.RuntimeID, appTemplateLabelEntity.RuntimeContextID, appTemplateLabelEntity.AppTemplateID, appTemplateLabelEntity.Key, appTemplateLabelEntity.Value, appTemplateLabelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Create(ctx, tenantID, appTemplateLabelModel)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_Upsert(t *testing.T) {
	testErr := errors.New("Test error")

	t.Run("Success update - Label for Runtime", func(t *testing.T) {
		labelModel := fixModelLabel(model.RuntimeLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`)
		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = ? AND owner = true))`)
		escapedExsistsQuery := regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2 AND owner = true))")

		mockedRows := sqlmock.NewRows(fixColumns).AddRow(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedExsistsQuery).WithArgs(labelID, tenantID).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Runtime Context", func(t *testing.T) {
		// GIVEN
		labelModel := fixModelLabel(model.RuntimeContextLabelableObject)
		labelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3))`)
		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = ? AND owner = true))`)
		escapedExsistsQuery := regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2 AND owner = true))")

		mockedRows := sqlmock.NewRows(fixColumns).AddRow(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedExsistsQuery).WithArgs(labelID, tenantID).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Application", func(t *testing.T) {
		labelModel := fixModelLabel(model.ApplicationLabelableObject)
		labelEntity := fixEntityLabel(model.ApplicationLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3))`)
		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = ? AND owner = true))`)
		escapedExsistsQuery := regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2 AND owner = true))")

		mockedRows := sqlmock.NewRows(fixColumns).AddRow(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedExsistsQuery).WithArgs(labelID, tenantID).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Tenant", func(t *testing.T) {
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND tenant_id = ?`)

		mockedRows := sqlmock.NewRows(fixColumns).AddRow(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tenantID, key).WillReturnRows(mockedRows)
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, labelEntity.TenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Application Template", func(t *testing.T) {
		appTemplateLabelModel := fixModelLabel(model.AppTemplateLabelableObject)
		appTemplateLabelEntity := fixEntityLabel(model.AppTemplateLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", appTemplateLabelModel).Return(appTemplateLabelEntity, nil).Once()
		mockConverter.On("FromEntity", appTemplateLabelEntity).Return(appTemplateLabelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_template_id = $2`)
		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ?`)

		mockedRows := sqlmock.NewRows(fixColumns).AddRow(appTemplateLabelEntity.ID, appTemplateLabelEntity.TenantID, appTemplateLabelEntity.AppID, appTemplateLabelEntity.RuntimeID, appTemplateLabelEntity.RuntimeContextID, appTemplateLabelEntity.AppTemplateID, appTemplateLabelEntity.Key, appTemplateLabelEntity.Value, appTemplateLabelEntity.Version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID).WillReturnRows(mockedRows)
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(appTemplateLabelEntity.Value, appTemplateLabelEntity.ID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, appTemplateLabelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Runtime", func(t *testing.T) {
		labelModel := fixModelLabel(model.RuntimeLabelableObject)
		labelEntity := fixEntityLabel(model.RuntimeLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`)
		escapedCheckParentAccessQuery := regexp.QuoteMeta("SELECT 1 FROM tenant_runtimes WHERE tenant_id = $1 AND id = $2 AND owner = $3")
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedCheckParentAccessQuery).WithArgs(tenantID, refID, true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Runtime Context", func(t *testing.T) {
		labelModel := fixModelLabel(model.RuntimeContextLabelableObject)
		labelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3))`)
		escapedCheckParentAccessQuery := regexp.QuoteMeta("SELECT 1 FROM tenant_runtime_contexts WHERE tenant_id = $1 AND id = $2 AND owner = $3")
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedCheckParentAccessQuery).WithArgs(tenantID, refID, true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Application", func(t *testing.T) {
		labelModel := fixModelLabel(model.ApplicationLabelableObject)
		labelEntity := fixEntityLabel(model.ApplicationLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3))`)
		escapedCheckParentAccessQuery := regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3")
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedCheckParentAccessQuery).WithArgs(tenantID, refID, true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Tenant", func(t *testing.T) {
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tenantID, key).WillReturnRows(mockedRows)
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Tenant", func(t *testing.T) {
		labelModel := fixModelLabel(model.AppTemplateLabelableObject)
		labelEntity := fixEntityLabel(model.AppTemplateLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_template_id = $2`)
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID).WillReturnRows(mockedRows)
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - GetByKey", func(t *testing.T) {
		labelModel := fixModelLabel(model.RuntimeLabelableObject)

		labelRepo := label.NewRepository(&automock.Converter{})

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error - Create", func(t *testing.T) {
		labelModel := fixModelLabel(model.RuntimeLabelableObject)
		labelEntity := fixEntityLabel(model.RuntimeLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`)
		escapedCheckParentAccessQuery := regexp.QuoteMeta("SELECT 1 FROM tenant_runtimes WHERE tenant_id = $1 AND id = $2 AND owner = $3")
		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )")

		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(key, refID, tenantID).WillReturnRows(mockedRows)
		dbMock.ExpectQuery(escapedCheckParentAccessQuery).WithArgs(tenantID, refID, true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.AppTemplateID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, tenantID, labelModel)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_UpdateWithVersion(t *testing.T) {
	version := 42

	var nilLabelModel *model.Label
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Application Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{appLabelEntity.Value, appLabelEntity.ID, version, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         applabelModel,
		DBEntity:            appLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	runtimeLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Runtime Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{runtimeLabelEntity.Value, runtimeLabelEntity.ID, version, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimelabelModel,
		DBEntity:            runtimeLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	runtimeCtxLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update RuntimeCtx Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.ID, version, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimeCtxlabelModel,
		DBEntity:            runtimeCtxLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	appLabelSuite.Run(t)
	runtimeLabelSuite.Run(t)
	runtimeCtxLabelSuite.Run(t)

	// Additional tests - tenant labels are updated globally as the tenant is embedded in the entity.
	t.Run("Success update - Label for Tenant", func(t *testing.T) {
		// GIVEN
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, version, labelEntity.TenantID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.UpdateWithVersion(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_GetByKey(t *testing.T) {
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Application Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(appLabelEntity.ID, appLabelEntity.TenantID, appLabelEntity.AppID, appLabelEntity.RuntimeID, appLabelEntity.RuntimeContextID, appLabelEntity.AppTemplateID, appLabelEntity.Key, appLabelEntity.Value, appLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: applabelModel,
		ExpectedDBEntity:    appLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	rtLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(runtimeLabelEntity.ID, runtimeLabelEntity.TenantID, runtimeLabelEntity.AppID, runtimeLabelEntity.RuntimeID, runtimeLabelEntity.RuntimeContextID, runtimeCtxLabelEntity.AppTemplateID, runtimeLabelEntity.Key, runtimeLabelEntity.Value, runtimeLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: runtimelabelModel,
		ExpectedDBEntity:    runtimeLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	rtCtxLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Context Label",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(runtimeCtxLabelEntity.ID, runtimeCtxLabelEntity.TenantID, runtimeCtxLabelEntity.AppID, runtimeCtxLabelEntity.RuntimeID, runtimeCtxLabelEntity.RuntimeContextID, runtimeCtxLabelEntity.AppTemplateID, runtimeCtxLabelEntity.Key, runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: runtimeCtxlabelModel,
		ExpectedDBEntity:    runtimeCtxLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		tenantLabelModel := fixModelLabel(model.TenantLabelableObject)
		tenantLabelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantLabelEntity).Return(tenantLabelModel, nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		mockedRows := sqlmock.NewRows(fixColumns).AddRow(tenantLabelEntity.ID, tenantLabelEntity.TenantID, tenantLabelEntity.AppID, tenantLabelEntity.RuntimeID, tenantLabelEntity.RuntimeContextID, tenantLabelEntity.AppTemplateID, tenantLabelEntity.Key, tenantLabelEntity.Value, tenantLabelEntity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tenantID, key).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.GetByKey(ctx, tenantID, model.TenantLabelableObject, tenantID, key)
		// THEN
		require.NoError(t, err)
		require.Equal(t, tenantLabelModel, actual)
		require.Equal(t, value, actual.Value)
	})

	t.Run("Success - Label for Application Template", func(t *testing.T) {
		appTemplatelabelModel := fixModelLabel(model.AppTemplateLabelableObject)
		appTemplateLabelEntity := fixEntityLabel(model.AppTemplateLabelableObject)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", appTemplateLabelEntity).Return(appTemplatelabelModel, nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		mockedRows := sqlmock.NewRows(fixColumns).AddRow(appTemplateLabelEntity.ID, appTemplateLabelEntity.TenantID, appTemplateLabelEntity.AppID, appTemplateLabelEntity.RuntimeID, appTemplateLabelEntity.RuntimeContextID, appTemplateLabelEntity.AppTemplateID, appTemplateLabelEntity.Key, appTemplateLabelEntity.Value, appTemplateLabelEntity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tenantID, key).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.GetByKey(ctx, tenantID, model.TenantLabelableObject, tenantID, key)
		// THEN
		require.NoError(t, err)
		require.Equal(t, appTemplatelabelModel, actual)
		require.Equal(t, value, actual.Value)
	})
}

func TestRepository_ListForObject(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		label1Model := fixModelLabelWithID("1", "foo", model.RuntimeLabelableObject)
		label2Model := fixModelLabelWithID("2", "bar", model.RuntimeLabelableObject)

		label1Entity := fixEntityLabelWithID("1", "foo", model.RuntimeLabelableObject)
		label2Entity := fixEntityLabelWithID("2", "bar", model.RuntimeLabelableObject)

		inputItems := []*label.Entity{label1Entity, label2Entity}
		expected := map[string]*model.Label{
			"foo": label1Model,
			"bar": label2Model,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE runtime_id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2))`)
		mockedRows := sqlmock.NewRows(fixColumns).
			AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
			AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: refID}, tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.RuntimeLabelableObject, refID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Runtime Context", func(t *testing.T) {
		// GIVEN
		label1Model := fixModelLabelWithID("1", "foo", model.RuntimeContextLabelableObject)
		label2Model := fixModelLabelWithID("2", "bar", model.RuntimeContextLabelableObject)

		label1Entity := fixEntityLabelWithID("1", "foo", model.RuntimeContextLabelableObject)
		label2Entity := fixEntityLabelWithID("2", "bar", model.RuntimeContextLabelableObject)

		inputItems := []*label.Entity{label1Entity, label2Entity}
		expected := map[string]*model.Label{
			"foo": label1Model,
			"bar": label2Model,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE runtime_context_id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2))`)
		mockedRows := sqlmock.NewRows(fixColumns).
			AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
			AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: refID}, tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.RuntimeContextLabelableObject, refID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		label1Model := fixModelLabelWithID("1", "foo", model.ApplicationLabelableObject)
		label2Model := fixModelLabelWithID("2", "bar", model.ApplicationLabelableObject)

		label1Entity := fixEntityLabelWithID("1", "foo", model.ApplicationLabelableObject)
		label2Entity := fixEntityLabelWithID("2", "bar", model.ApplicationLabelableObject)

		inputItems := []*label.Entity{label1Entity, label2Entity}
		expected := map[string]*model.Label{
			"foo": label1Model,
			"bar": label2Model,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE app_id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2))`)
		mockedRows := sqlmock.NewRows(fixColumns).
			AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
			AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label1Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: refID}, tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.ApplicationLabelableObject, refID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		// GIVEN
		label1Model := fixModelLabelWithID("1", "foo", model.TenantLabelableObject)
		label2Model := fixModelLabelWithID("2", "bar", model.TenantLabelableObject)

		label1Entity := fixEntityLabelWithID("1", "foo", model.TenantLabelableObject)
		label2Entity := fixEntityLabelWithID("2", "bar", model.TenantLabelableObject)

		inputItems := []*label.Entity{label1Entity, label2Entity}
		expected := map[string]*model.Label{
			"foo": label1Model,
			"bar": label2Model,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND app_id IS NULL AND runtime_context_id IS NULL AND runtime_id IS NULL`)
		mockedRows := sqlmock.NewRows(fixColumns).
			AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
			AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.TenantLabelableObject, tenantID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Application Template", func(t *testing.T) {
		// GIVEN
		label1Model := fixModelLabelWithID("1", "foo", model.AppTemplateLabelableObject)
		label2Model := fixModelLabelWithID("2", "bar", model.AppTemplateLabelableObject)

		label1Entity := fixEntityLabelWithID("1", "foo", model.AppTemplateLabelableObject)
		label2Entity := fixEntityLabelWithID("2", "bar", model.AppTemplateLabelableObject)

		inputItems := []*label.Entity{label1Entity, label2Entity}
		expected := map[string]*model.Label{
			"foo": label1Model,
			"bar": label2Model,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND app_id IS NULL AND runtime_context_id IS NULL AND runtime_id IS NULL`)
		mockedRows := sqlmock.NewRows(fixColumns).
			AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
			AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.TenantLabelableObject, tenantID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error - Doesn't exist", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE app_id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2))`)
		mockedRows := sqlmock.NewRows(fixColumns)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: refID}, tenantID).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tenantID, model.ApplicationLabelableObject, refID)
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		labelRepo := label.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE app_id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2))`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: refID}, tenantID).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.ListForObject(ctx, tenantID, model.ApplicationLabelableObject, refID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)

		// WHEN
		_, err := labelRepo.ListForObject(context.TODO(), tenantID, model.ApplicationLabelableObject, refID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_ListByKey(t *testing.T) {
	label1Model := fixModelLabelWithID("1", "foo", model.RuntimeLabelableObject)
	label2Model := fixModelLabelWithID("2", "bar", model.ApplicationLabelableObject)
	label3Model := fixModelLabelWithID("3", "bar", model.RuntimeContextLabelableObject)

	label1Entity := fixEntityLabelWithID("1", "foo", model.RuntimeLabelableObject)
	label2Entity := fixEntityLabelWithID("2", "bar", model.ApplicationLabelableObject)
	label3Entity := fixEntityLabelWithID("3", "bar", model.RuntimeContextLabelableObject)

	suite := testdb.RepoListTestSuite{
		Name: "List Labels by key",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND (id IN (SELECT id FROM labels_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{key, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
						AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version).
						AddRow(label3Entity.ID, label3Entity.TenantID, label3Entity.AppID, label3Entity.RuntimeID, label3Entity.RuntimeContextID, label3Entity.AppTemplateID, label3Entity.Key, label3Entity.Value, label3Entity.Version),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   label.NewRepository,
		ExpectedModelEntities: []interface{}{label1Model, label2Model, label3Model},
		ExpectedDBEntities:    []interface{}{label1Entity, label2Entity, label3Entity},
		MethodArgs:            []interface{}{tenantID, key},
		MethodName:            "ListByKey",
	}

	suite.Run(t)
}

func TestRepository_ListGlobalByKey(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"
		version := 42

		inputItems := []*label.Entity{
			{ID: "1", Key: labelKey, Value: "test1", AppID: sql.NullString{Valid: true, String: objIDs[0]}, Version: version},
			{ID: "2", Key: labelKey, Value: "test2", AppID: sql.NullString{Valid: true, String: objIDs[1]}, Version: version},
			{ID: "3", Key: labelKey, Value: "test3", AppID: sql.NullString{Valid: true, String: objIDs[2]}, Version: version},
		}
		expected := []*model.Label{
			{ID: "1", Key: labelKey, Value: "test1", ObjectType: objType, ObjectID: objIDs[0], Version: version},
			{ID: "2", Key: labelKey, Value: "test2", ObjectType: objType, ObjectID: objIDs[1], Version: version},
			{ID: "3", Key: labelKey, Value: "test3", ObjectType: objType, ObjectID: objIDs[2], Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			for i := range expected {
				if expected[i].ObjectID == entity.AppID.String {
					mockConverter.On("FromEntity", entity).Return(expected[i], nil).Once()
				}
			}
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_template_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", nil, nil, labelKey, "test1", objIDs[0], nil, nil, version).
			AddRow("2", nil, nil, labelKey, "test2", objIDs[1], nil, nil, version).
			AddRow("3", nil, nil, labelKey, "test3", objIDs[2], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListGlobalByKey(ctx, labelKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		labelKey := "key"

		labelRepo := label.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKey(ctx, labelKey)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
	})

	t.Run("Error - Converting entity", func(t *testing.T) {
		// GIVEN
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"
		testErr := errors.New("test error")
		version := 42

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", mock.Anything).Return(&model.Label{}, testErr).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_template_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", nil, nil, labelKey, "test1", objIDs[0], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKey(ctx, labelKey)
		// THEN
		require.Error(t, err)
	})
}

func TestRepository_ListGlobalByKeyAndObjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"
		version := 42

		inputItems := []*label.Entity{
			{ID: "1", Key: labelKey, Value: "test1", AppID: sql.NullString{Valid: true, String: objIDs[0]}, Version: version},
			{ID: "2", Key: labelKey, Value: "test2", AppID: sql.NullString{Valid: true, String: objIDs[1]}, Version: version},
			{ID: "3", Key: labelKey, Value: "test3", AppID: sql.NullString{Valid: true, String: objIDs[2]}, Version: version},
		}
		expected := []*model.Label{
			{ID: "1", Key: labelKey, Value: "test1", ObjectType: objType, ObjectID: objIDs[0], Version: version},
			{ID: "2", Key: labelKey, Value: "test2", ObjectType: objType, ObjectID: objIDs[1], Version: version},
			{ID: "3", Key: labelKey, Value: "test3", ObjectType: objType, ObjectID: objIDs[2], Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			for i := range expected {
				if expected[i].ObjectID == entity.AppID.String {
					mockConverter.On("FromEntity", entity).Return(expected[i], nil).Once()
				}
			}
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4) FOR UPDATE`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_template_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", nil, nil, labelKey, "test1", objIDs[0], nil, nil, version).
			AddRow("2", nil, nil, labelKey, "test2", objIDs[1], nil, nil, version).
			AddRow("3", nil, nil, labelKey, "test3", objIDs[2], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"

		labelRepo := label.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4) FOR UPDATE`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).
			WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
	})

	t.Run("Error - Converting entity", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"
		testErr := errors.New("test error")
		version := 42

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", mock.Anything).Return(&model.Label{}, testErr).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4) FOR UPDATE`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_template_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", nil, nil, labelKey, "test1", objIDs[0], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.Error(t, err)
	})
}

func TestRepository_Delete(t *testing.T) {
	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID, key).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tenantID, model.TenantLabelableObject, tenantID, key)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND app_template_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, refID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tenantID, model.AppTemplateLabelableObject, refID, key)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_DeleteAll(t *testing.T) {
	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label Delete All",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE app_id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label Delete All",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE runtime_id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label Delete All",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE runtime_context_id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1`)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tenantID, model.TenantLabelableObject, tenantID)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Application Template", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE app_template_id = $1`)
		dbMock.ExpectExec(escapedQuery).WithArgs(refID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tenantID, model.AppTemplateLabelableObject, refID)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_DeleteByKey(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Label Delete By Key",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND (id IN (SELECT id FROM labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{key, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, key},
		MethodName:          "DeleteByKey",
		IsDeleteMany:        true,
	}

	suite.Run(t)
}

func TestRepository_DeleteByKeyNegationPattern(t *testing.T) {
	pattern := "pattern"

	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label DeleteByKeyNegationPattern",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label DeleteByKeyNegationPattern",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label DeleteByKeyNegationPattern",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1 AND NOT key ~ $2 `)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID, pattern).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKeyNegationPattern(ctx, tenantID, model.TenantLabelableObject, tenantID, pattern)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Application Template", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND app_template_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(pattern, refID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKeyNegationPattern(ctx, tenantID, model.AppTemplateLabelableObject, refID, pattern)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_GetScenarioLabelsForRuntimes(t *testing.T) {
	rt1ID := "rt1"
	rt2ID := "rt2"

	label1Model := fixModelLabelWithRefID("1", model.ScenariosKey, model.RuntimeLabelableObject, rt1ID)
	label2Model := fixModelLabelWithRefID("2", model.ScenariosKey, model.RuntimeLabelableObject, rt2ID)

	label1Entity := fixEntityLabelWithRefID("1", model.ScenariosKey, model.RuntimeLabelableObject, rt1ID)
	label2Entity := fixEntityLabelWithRefID("2", model.ScenariosKey, model.RuntimeLabelableObject, rt2ID)

	suite := testdb.RepoListTestSuite{
		Name: "List Runtime Scenarios Matching Selector",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, app_template_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id IN ($2, $3) AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $4))`),
				Args:     []driver.Value{model.ScenariosKey, rt1ID, rt2ID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(label1Entity.ID, label1Entity.TenantID, label1Entity.AppID, label1Entity.RuntimeID, label1Entity.RuntimeContextID, label1Entity.AppTemplateID, label1Entity.Key, label1Entity.Value, label1Entity.Version).
						AddRow(label2Entity.ID, label2Entity.TenantID, label2Entity.AppID, label2Entity.RuntimeID, label2Entity.RuntimeContextID, label2Entity.AppTemplateID, label2Entity.Key, label2Entity.Value, label2Entity.Version),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc:   label.NewRepository,
		ExpectedModelEntities: []interface{}{label1Model, label2Model},
		ExpectedDBEntities:    []interface{}{label1Entity, label2Entity},
		MethodArgs:            []interface{}{tenantID, []string{rt1ID, rt2ID}},
		MethodName:            "GetScenarioLabelsForRuntimes",
	}

	suite.Run(t)
}
