package repo_test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder(t *testing.T) {
	sut := repo.NewQueryBuilder(appTableName, appColumns)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("success with only default tenant condition", func(t *testing.T) {
		// GIVEN
		expectedQuery := fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1"))

		// WHEN
		query, args, err := sut.BuildQuery(resourceType, tenantID, true, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with only default tenant condition and no argument rebinding", func(t *testing.T) {
		// GIVEN
		expectedQuery := fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "?"))

		// WHEN
		query, args, err := sut.BuildQuery(resourceType, tenantID, false, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := fmt.Sprintf("SELECT id, name, description FROM %s WHERE first_name = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2"))
		expectedFirstName := "foo"

		// WHEN
		query, args, err := sut.BuildQuery(resourceType, tenantID, true, repo.Conditions{repo.NewEqualCondition("first_name", expectedFirstName)}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 2, len(args))
		assert.Equal(t, expectedFirstName, args[0])
		assert.Equal(t, tenantID, args[1])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with values", func(t *testing.T) {
		// GIVEN
		expectedQuery := fmt.Sprintf("SELECT id, name, description FROM %s WHERE first_name IN ($1, $2) AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3"))
		expectedFirstINValue := "foo"
		expectedSecondINValue := "bar"

		// WHEN
		query, args, err := sut.BuildQuery(resourceType, tenantID, true, repo.Conditions{repo.NewInConditionForStringValues("first_name", []string{"foo", "bar"})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 3, len(args))
		assert.Equal(t, expectedFirstINValue, args[0])
		assert.Equal(t, expectedSecondINValue, args[1])
		assert.Equal(t, tenantID, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with subquery", func(t *testing.T) {
		// GIVEN
		expectedQuery := fmt.Sprintf("SELECT id, name, description FROM %s WHERE first_name IN (SELECT name from names WHERE description = $1 AND id = $2) AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3"))
		expectedFirstArgument := "foo"
		expectedSecondArgument := 3

		// WHEN
		query, args, err := sut.BuildQuery(resourceType, tenantID, true, repo.Conditions{repo.NewInConditionForSubQuery("first_name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 3, len(args))
		assert.Equal(t, expectedFirstArgument, args[0])
		assert.Equal(t, expectedSecondArgument, args[1])
		assert.Equal(t, tenantID, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		// WHEN
		query, args, err := sut.BuildQuery(resourceType, "", true, repo.Conditions{}...)

		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsTenantRequired(err))
		assert.Equal(t, "", removeWhitespace(query))
		assert.Equal(t, []interface{}(nil), args)
	})
}

func TestQueryBuilderWithEmbeddedTenant(t *testing.T) {
	sut := repo.NewQueryBuilderWithEmbeddedTenant(userTableName, "tenant_id", []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success with only default tenant condition", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1"

		// WHEN
		query, args, err := sut.BuildQuery(UserType, tenantID, true, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with only default tenant condition and no argument rebinding", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = ?"

		// WHEN
		query, args, err := sut.BuildQuery(UserType, tenantID, false, repo.Conditions{}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2"
		expectedFirstName := "foo"

		// WHEN
		query, args, err := sut.BuildQuery(UserType, tenantID, true, repo.Conditions{repo.NewEqualCondition("first_name", expectedFirstName)}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 2, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedFirstName, args[1])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with values", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name IN ($2, $3)"
		expectedFirstINValue := "foo"
		expectedSecondINValue := "bar"

		// WHEN
		query, args, err := sut.BuildQuery(UserType, tenantID, true, repo.Conditions{repo.NewInConditionForStringValues("first_name", []string{"foo", "bar"})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 3, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedFirstINValue, args[1])
		assert.Equal(t, expectedSecondINValue, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with subquery", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name IN (SELECT name from names WHERE description = $2 AND id = $3)"
		expectedFirstArgument := "foo"
		expectedSecondArgument := 3

		// WHEN
		query, args, err := sut.BuildQuery(UserType, tenantID, true, repo.Conditions{repo.NewInConditionForSubQuery("first_name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 3, len(args))
		assert.Equal(t, tenantID, args[0])
		assert.Equal(t, expectedFirstArgument, args[1])
		assert.Equal(t, expectedSecondArgument, args[2])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		// WHEN
		query, args, err := sut.BuildQuery(UserType, "", true, repo.Conditions{}...)

		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsTenantRequired(err))
		assert.Equal(t, "", removeWhitespace(query))
		assert.Equal(t, []interface{}(nil), args)
	})
}

func TestQueryBuilderGlobal(t *testing.T) {
	sut := repo.NewQueryBuilderGlobal(UserType, userTableName, []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success without conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users"

		// WHEN
		query, args, err := sut.BuildQueryGlobal(false, repo.Conditions{}...)
		var (
			tenantRuntimeContextTable           = "tenant_runtime_contexts"
			tenantRuntimeContextSelectedColumns = []string{"tenant_id"}
			labelsTable                         = "labels"
			labelsSelectedColumns               = []string{"app_template_id"}
			applicationTable                    = "applications"
			applicationsSelectedColumns         = []string{"id"}
			tenantApplicationsTable             = "tenant_applications"
			tenantApplicationsSelectedColumns   = []string{"tenant_id"}

			appTemplateIDColumn = "app_template_id"
			keyColumn           = "key"
		)

		tenantRuntimeContextQueryBuilder := repo.NewQueryBuilderGlobal(resource.RuntimeContext, tenantRuntimeContextTable, tenantRuntimeContextSelectedColumns)
		labelsQueryBuilder := repo.NewQueryBuilderGlobal(resource.Label, labelsTable, labelsSelectedColumns)
		applicationQueryBuilder := repo.NewQueryBuilderGlobal(resource.Application, applicationTable, applicationsSelectedColumns)
		tenantApplicationsQueryBuilder := repo.NewQueryBuilderGlobal(resource.Application, tenantApplicationsTable, tenantApplicationsSelectedColumns)

		subaccountConditions := repo.Conditions{repo.NewEqualCondition("type", "subaccount")}

		tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs, err := tenantRuntimeContextQueryBuilder.BuildQueryGlobal(false, repo.Conditions{}...)
		applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs, err := labelsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewEqualCondition(keyColumn, "selfRegDistinguishLabel"), repo.NewNotNullCondition(appTemplateIDColumn)}...)
		applicationSubquery, applicationArgs, err := applicationQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(appTemplateIDColumn, applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs)}...)
		tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs, err := tenantApplicationsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery("id", applicationSubquery, applicationArgs)}...)

		subscriptionConditions := repo.Conditions{
			repo.NewInConditionForSubQuery("id", tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs),
			repo.NewInConditionForSubQuery("id", tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs),
		}

		conditions := repo.And(
			append(
				append(
					repo.ConditionTreesFromConditions(subaccountConditions),
					repo.Or(repo.ConditionTreesFromConditions(subscriptionConditions)...),
				),
				&repo.ConditionTree{Operand: repo.NewEqualCondition("id", "given-id")})...,
		)

		spew.Dump(conditions)
		spew.Dump(conditions.BuildSubquery())

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 0, len(args))
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1"
		expectedFirstName := "foo"

		// WHEN
		query, args, err := sut.BuildQueryGlobal(true, repo.Conditions{repo.NewEqualCondition("first_name", expectedFirstName)}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, expectedFirstName, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with values", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name IN ($1, $2)"
		expectedFirstINValue := "foo"
		expectedSecondINValue := "bar"

		// WHEN
		query, args, err := sut.BuildQueryGlobal(true, repo.Conditions{repo.NewInConditionForStringValues("first_name", []string{"foo", "bar"})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 2, len(args))
		assert.Equal(t, expectedFirstINValue, args[0])
		assert.Equal(t, expectedSecondINValue, args[1])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})

	t.Run("success with IN condition with subquery", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name IN (SELECT name from names WHERE description = $1 AND id = $2)"
		expectedFirstArgument := "foo"
		expectedSecondArgument := 3

		// WHEN
		query, args, err := sut.BuildQueryGlobal(true, repo.Conditions{repo.NewInConditionForSubQuery("first_name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}...)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 2, len(args))
		assert.Equal(t, expectedFirstArgument, args[0])
		assert.Equal(t, expectedSecondArgument, args[1])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})
}

func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
