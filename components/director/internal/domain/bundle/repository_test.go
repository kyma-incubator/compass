package bundle_test

import (
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	name := "foo"
	desc := "bar"

	var nilBundleMode *model.Bundle
	bndlModel := fixBundleModel(name, desc)
	bndlEntity := fixEntityBundle(bundleID, name, desc)

	defAuth, err := json.Marshal(bndlModel.DefaultInstanceAuth)
	require.NoError(t, err)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.bundles \(.+\) VALUES \(.+\)$`,
				Args:        fixBundleCreateArgs(string(defAuth), *bndlModel.InstanceAuthRequestInputSchema, bndlModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndlModel,
		DBEntity:            bndlEntity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.bundles SET name = ?, description = ?, instance_auth_request_json_schema = ?, default_instance_auth = ?, ord_id = ?, short_description = ?, links = ?, labels = ?, credential_exchange_strategies = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, documentation_labels = ? WHERE id = ? AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = ? AND owner = true))`)

	var nilBundleMode *model.Bundle
	bndl := fixBundleModel("foo", "update")
	entity := fixEntityBundle(bundleID, "foo", "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateQuery,
				Args:          []driver.Value{entity.Name, entity.Description, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, entity.OrdID, entity.ShortDescription, entity.Links, entity.Labels, entity.CredentialExchangeStrategies, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels, entity.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndl,
		DBEntity:            entity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Bundle Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{bundleID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Bundle Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
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
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		TargetID:            bundleID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	bndlEntity := fixEntityBundle(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRow(bundleID, "placeholder")...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_GetForApplication(t *testing.T) {
	bndlEntity := fixEntityBundle(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle For Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{bundleID, appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRow(bundleID, "placeholder")...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID, appID},
		MethodName:          "GetForApplication",
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	pageSize := 1
	cursor := ""

	emptyPageAppID := "emptyPageAppID"

	onePageAppID := "onePageAppID"
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstBundleEntity := fixEntityBundle(firstBundleID, "foo", "bar")
	firstBundleEntity.ApplicationID = onePageAppID
	firstBndlModel := fixBundleModelWithID(firstBundleID, "foo", desc)
	firstBndlModel.ApplicationID = onePageAppID

	multiplePagesAppID := "multiplePagesAppID"

	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondBundleEntity := fixEntityBundle(secondBundleID, "foo", "bar")
	secondBundleEntity.ApplicationID = multiplePagesAppID
	secondBndlModel := fixBundleModelWithID(secondBundleID, "foo", desc)
	secondBndlModel.ApplicationID = multiplePagesAppID

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Bundles for multiple Applications with paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`(SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $1)) AND app_id = $2 ORDER BY app_id ASC, id ASC LIMIT $3 OFFSET $4)
												UNION
												(SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $5)) AND app_id = $6 ORDER BY app_id ASC, id ASC LIMIT $7 OFFSET $8)
												UNION
												(SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $9)) AND app_id = $10 ORDER BY app_id ASC, id ASC LIMIT $11 OFFSET $12)`),

				Args:     []driver.Value{tenantID, emptyPageAppID, pageSize, 0, tenantID, onePageAppID, pageSize, 0, tenantID, multiplePagesAppID, pageSize, 0},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).AddRow(fixBundleRowWithAppID(firstBundleID, onePageAppID)...).AddRow(fixBundleRowWithAppID(secondBundleID, multiplePagesAppID)...)}
				},
			},
			{
				Query:    regexp.QuoteMeta(`SELECT app_id AS id, COUNT(*) AS total_count FROM public.bundles WHERE (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $1)) GROUP BY app_id ORDER BY app_id ASC`),
				Args:     []driver.Value{tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id", "total_count"}).AddRow(emptyPageAppID, 0).AddRow(onePageAppID, 1).AddRow(multiplePagesAppID, 2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: nil,
				ExpectedDBEntities:    nil,
				ExpectedPage: &model.BundlePage{
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
				ExpectedModelEntities: []interface{}{firstBndlModel},
				ExpectedDBEntities:    []interface{}{firstBundleEntity},
				ExpectedPage: &model.BundlePage{
					Data: []*model.Bundle{firstBndlModel},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 1,
				},
			},
			{
				ExpectedModelEntities: []interface{}{secondBndlModel},
				ExpectedDBEntities:    []interface{}{secondBundleEntity},
				ExpectedPage: &model.BundlePage{
					Data: []*model.Bundle{secondBndlModel},
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
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		MethodName:          "ListByApplicationIDs",
		MethodArgs:          []interface{}{tenantID, []string{emptyPageAppID, onePageAppID, multiplePagesAppID}, pageSize, cursor},
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationIDNoPaging(t *testing.T) {
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstBundleEntity := fixEntityBundle(firstBundleID, "foo", "bar")
	firstBndlModel := fixBundleModelWithID(firstBundleID, "foo", desc)
	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondBundleEntity := fixEntityBundle(secondBundleID, "foo", "bar")
	secondBndlModel := fixBundleModelWithID(secondBundleID, "foo", desc)

	suite := testdb.RepoListTestSuite{
		Name: "List Bundles No Paging",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.bundles WHERE app_id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns()).
						AddRow(fixBundleRow(firstBundleID, "placeholder")...).AddRow(fixBundleRow(secondBundleID, "placeholder")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixBundleColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundle.NewRepository,
		ExpectedModelEntities: []interface{}{firstBndlModel, secondBndlModel},
		ExpectedDBEntities:    []interface{}{firstBundleEntity, secondBundleEntity},
		MethodArgs:            []interface{}{tenantID, appID},
		MethodName:            "ListByApplicationIDNoPaging",
	}

	suite.Run(t)
}
