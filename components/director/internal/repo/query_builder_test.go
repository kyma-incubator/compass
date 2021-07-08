package repo_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder(t *testing.T) {
	givenTenant := uuidA()
	sut := repo.NewQueryBuilder(UserType, "users", "tenant_id", []string{"id_col", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success with only default tenant condition", func(t *testing.T) {
		// GIVEN
		expectedQuery, expectedArgumentsNumber := getExpectedQueryWithTenantCondition()

		// WHEN
		query, args, err := sut.BuildQuery(givenTenant, true, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedArgumentsNumber, len(args))
		assert.Equal(t, givenTenant, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with only default tenant condition and no argument rebinding", func(t *testing.T) {
		// GIVEN
		expectedQuery, expectedArgumentsNumber := getExpectedQueryWithTenantConditionNoRebinding()

		// WHEN
		query, args, err := sut.BuildQuery(givenTenant, false, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedArgumentsNumber, len(args))
		assert.Equal(t, givenTenant, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery, expectedArgumentsNumber := getExpectedQueryWithMoreConditions()
		expectedFirstName := "foo"

		// WHEN
		query, args, err := sut.BuildQuery(givenTenant, true, repo.Conditions{repo.NewEqualCondition("first_name", expectedFirstName)}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedArgumentsNumber, len(args))
		assert.Equal(t, givenTenant, args[0])
		assert.Equal(t, expectedFirstName, args[1])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with values", func(t *testing.T) {
		// GIVEN
		expectedQuery, expectedArgumentsNumber := getExpectedQueryWithINValues()
		expectedFirstINValue := "foo"
		expectedSecondINValue := "bar"

		// WHEN
		query, args, err := sut.BuildQuery(givenTenant, true, repo.Conditions{repo.NewInConditionForStringValues("first_name", []string{"foo", "bar"})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedArgumentsNumber, len(args))
		assert.Equal(t, givenTenant, args[0])
		assert.Equal(t, expectedFirstINValue, args[1])
		assert.Equal(t, expectedSecondINValue, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with subquery", func(t *testing.T) {
		// GIVEN
		expectedQuery, expectedArgumentsNumber := getExpectedQueryWithINSubquery()
		expectedFirstArgument := "foo"
		expectedSecondArgument := 3

		// WHEN
		query, args, err := sut.BuildQuery(givenTenant, true, repo.Conditions{repo.NewInConditionForSubQuery("first_name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedArgumentsNumber, len(args))
		assert.Equal(t, givenTenant, args[0])
		assert.Equal(t, expectedFirstArgument, args[1])
		assert.Equal(t, expectedSecondArgument, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))

	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		// WHEN
		query, args, err := sut.BuildQuery("", true, repo.Conditions{}...)

		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsTenantRequired(err))
		assert.Equal(t, "", removeWhitespace(query))
		assert.Equal(t, []interface{}(nil), args)
	})

}

func getExpectedQueryWithTenantCondition() (string, int) {
	expectedQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s", fixTenantIsolationSubquery())
	return expectedQuery, 1
}

func getExpectedQueryWithTenantConditionNoRebinding() (string, int) {
	expectedQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s", fixTenantIsolationSubqueryNoRebind())
	return expectedQuery, 1
}

func getExpectedQueryWithMoreConditions() (string, int) {
	expectedQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name = $2", fixTenantIsolationSubquery())
	return expectedQuery, 2
}

func getExpectedQueryWithINValues() (string, int) {
	expectedQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name IN ($2, $3)", fixTenantIsolationSubquery())
	return expectedQuery, 3
}

func getExpectedQueryWithINSubquery() (string, int) {
	expectedQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name IN (SELECT name from names WHERE description = $2 AND id = $3)", fixTenantIsolationSubquery())
	return expectedQuery, 3
}

func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
