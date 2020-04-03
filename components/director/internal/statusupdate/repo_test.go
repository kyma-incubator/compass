package statusupdate_test

import (
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

const testID = "foo"

func TestUpdate_IsConnected(t *testing.T) {
	db, dbMock := testdb.MockDatabase(t)
	defer dbMock.AssertExpectations(t)
	dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.applications WHERE status_condition = 'CONNECTED' AND id = $1`)).
		WithArgs(testID).
		WillReturnRows(testdb.RowWhenObjectExist())

	statusupdate.NewUpdate()
}
