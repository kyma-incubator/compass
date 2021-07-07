package tenantindex_test

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantindex"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRepository_GetOwnerTenantByResourceID(t *testing.T) {
	// given
	appID := "testAppID"
	callingTenantID := "testCallingTenantID"
	appOwnerTenantID := "appOwnerTenantID"
	selectQuery := fmt.Sprintf(`^SELECT tenant_id FROM "public"."id_tenant_id_index" WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows([]string{"tenant_id"}).
			AddRow(appOwnerTenantID)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(callingTenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		repository := tenantindex.NewRepository()
		// WHEN
		resultOwnerTenantID, err := repository.GetOwnerTenantByResourceID(ctx, callingTenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Equal(t, appOwnerTenantID, resultOwnerTenantID)
		sqlMock.AssertExpectations(t)
	})
}

func fixTenantIsolationSubquery() string {
	return `tenant_id IN \( with recursive children AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \$1 UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN children t on t\.id = t2\.parent\) SELECT id from children \)`
}