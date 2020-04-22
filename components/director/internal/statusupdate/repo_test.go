package statusupdate_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

const testID = "foo"

func TestRepository_IsConnected(t *testing.T) {
	testError := errors.New("test")

	t.Run("Success for applications", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.applications WHERE id = $1 AND status_condition = 'CONNECTED'`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())
		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := statusupdate.NewRepository()

		//WHEN
		res, err := repo.IsConnected(ctx, testID, "applications")

		//THEN
		require.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("Success for runtimes", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.runtimes WHERE id = $1 AND status_condition = 'CONNECTED'`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())
		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := statusupdate.NewRepository()

		//WHEN
		res, err := repo.IsConnected(ctx, testID, "runtimes")

		//THEN
		require.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("Error for applications", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.applications WHERE id = $1 AND status_condition = 'CONNECTED'`)).
			WithArgs(testID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := statusupdate.NewRepository()

		//WHEN
		res, err := repo.IsConnected(ctx, testID, "applications")

		//THEN
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", testError))
		assert.False(t, res)
	})

	t.Run("Error for runtimes", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.runtimes WHERE id = $1 AND status_condition = 'CONNECTED'`)).
			WithArgs(testID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := statusupdate.NewRepository()

		//WHEN
		res, err := repo.IsConnected(ctx, testID, "runtimes")

		//THEN
		require.EqualError(t, err, fmt.Sprintf("while getting object from DB: %s", testError))
		assert.False(t, res)
	})

}

func TestRepository_UpdateStatus(t *testing.T) {
	timestamp := time.Now()
	repo := statusupdate.NewRepository()
	repo.SetTimestampGen(func() time.Time { return timestamp })
	testError := errors.New("test")

	t.Run("Success for applications", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.applications SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2`)).
			WithArgs(timestamp, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)

		//WHEN
		err := repo.UpdateStatus(ctx, testID, "applications")

		//THEN
		require.NoError(t, err)
	})
	t.Run("Success for runtimes", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.runtimes SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2`)).
			WithArgs(timestamp, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)

		//WHEN
		err := repo.UpdateStatus(ctx, testID, "runtimes")

		//THEN
		require.NoError(t, err)
	})
	t.Run("Error for applications", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.applications SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2`)).
			WithArgs(timestamp, testID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), db)

		//WHEN
		err := repo.UpdateStatus(ctx, testID, "applications")

		//THEN
		require.EqualError(t, err, fmt.Sprintf("while updating applications status: %s", testError.Error()))
	})

	t.Run("Error for runtimes", func(t *testing.T) {

		//GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.runtimes SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2`)).
			WithArgs(timestamp, testID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), db)

		//WHEN
		err := repo.UpdateStatus(ctx, testID, "runtimes")

		//THEN
		require.EqualError(t, err, fmt.Sprintf("while updating runtimes status: %s", testError.Error()))
	})
}
