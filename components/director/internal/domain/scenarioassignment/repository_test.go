package scenarioassignment_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", fixModel()).Return(fixEntity()).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.automatic_scenario_assignments ( scenario, tenant_id, selector_key, selector_value ) VALUES ( ?, ?, ?, ? )`)).
			WithArgs("scenario-A", "tenant", "key", "value").
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, fixModel())

		// then
		assert.NoError(t, err)
	})

	t.Run("DB error", func(t *testing.T) {
		// given

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", fixModel()).Return(fixEntity()).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(fixError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := scenarioassignment.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, fixModel())

		// then
		require.EqualError(t, err, "while inserting row to 'public.automatic_scenario_assignments' table: some error")
	})
}
