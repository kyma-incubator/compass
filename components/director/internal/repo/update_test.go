package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdaterGlobal(t *testing.T) {
	t.Run("Global Update", func(t *testing.T) {
		sut := repo.NewUpdaterGlobal(UserType, "users", []string{"first_name", "last_name", "age"}, []string{"id"})
		givenUser := User{
			ID:        "given_id",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		t.Run("success", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id = ?")).
				WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 1))
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.NoError(t, err)
		})

		t.Run("success when no id column", func(t *testing.T) {
			// GIVEN
			sut := repo.NewUpdaterGlobal(UserType, "users", []string{"first_name", "last_name", "age"}, []string{})
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?")).
				WithArgs("given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(0, 1))
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.NoError(t, err)
		})

		t.Run("returns error when operation on db failed", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec("UPDATE users .*").
				WillReturnError(someError())
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run("returns non unique error", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec("UPDATE users .*").
				WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.True(t, apperrors.IsNotUniqueError(err))
		})

		t.Run("returns error if modified more than one row", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id = ?")).
				WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 157))
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "should update single row, but updated 157 rows")
		})

		t.Run("returns error if does not modified any row", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id = ?")).
				WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 0))
			// WHEN
			err := sut.UpdateSingleGlobal(ctx, givenUser)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "should update single row, but updated 0 rows")
		})

		t.Run("returns error if missing persistence context", func(t *testing.T) {
			// WHEN
			err := sut.UpdateSingleGlobal(context.TODO(), User{})
			// THEN
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})

		t.Run("returns error if entity is nil", func(t *testing.T) {
			// WHEN
			err := sut.UpdateSingleGlobal(context.TODO(), nil)
			// THEN
			require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
		})
	})

	t.Run("Update with embedded tenant", func(t *testing.T) {
		tests := []struct {
			Name                string
			AdditionalCondition string
			Method              func(updater repo.UpdaterGlobal) func(ctx context.Context, dbEntity interface{}) error
		}{
			{
				Name: "UpdateSingleGlobal",
				Method: func(updater repo.UpdaterGlobal) func(ctx context.Context, dbEntity interface{}) error {
					return updater.UpdateSingleGlobal
				},
			},
			{
				Name:                "UpdateSingleWithVersionGlobal",
				AdditionalCondition: ", version = version+1",
				Method: func(updater repo.UpdaterGlobal) func(ctx context.Context, dbEntity interface{}) error {
					return updater.UpdateSingleWithVersionGlobal
				},
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				sut := repo.NewUpdaterWithEmbeddedTenant(UserType, "users", []string{"first_name", "last_name", "age"}, "tenant_id", []string{"id"})
				givenUser := User{
					ID:        "given_id",
					Tenant:    "given_tenant",
					FirstName: "given_first_name",
					LastName:  "given_last_name",
					Age:       55,
				}

				t.Run("success", func(t *testing.T) {
					// GIVEN
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)

					mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?"+test.AdditionalCondition+" WHERE id = ? AND tenant_id = ?")).
						WithArgs("given_first_name", "given_last_name", 55, "given_id", "given_tenant").WillReturnResult(sqlmock.NewResult(0, 1))
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.NoError(t, err)
				})

				t.Run("success when no id column", func(t *testing.T) {
					// GIVEN
					sut := repo.NewUpdaterWithEmbeddedTenant(UserType, "users", []string{"first_name", "last_name", "age"}, "tenant_id", []string{})
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)

					mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?"+test.AdditionalCondition+" WHERE tenant_id = ?")).
						WithArgs("given_first_name", "given_last_name", 55, "given_tenant").WillReturnResult(sqlmock.NewResult(0, 1))
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.NoError(t, err)
				})

				t.Run("returns error when operation on db failed", func(t *testing.T) {
					// GIVEN
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)
					mock.ExpectExec("UPDATE users .*").
						WillReturnError(someError())
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
				})

				t.Run("context properly canceled", func(t *testing.T) {
					db, mock := testdb.MockDatabase(t)
					defer mock.AssertExpectations(t)

					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
					defer cancel()

					ctx = persistence.SaveToContext(ctx, db)

					err := test.Method(sut)(ctx, givenUser)

					require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
				})

				t.Run("returns non unique error", func(t *testing.T) {
					// GIVEN
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)
					mock.ExpectExec("UPDATE users .*").
						WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.True(t, apperrors.IsNotUniqueError(err))
				})

				t.Run("returns error if modified more than one row", func(t *testing.T) {
					// GIVEN
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)

					mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?"+test.AdditionalCondition+" WHERE id = ? AND tenant_id = ?")).
						WithArgs("given_first_name", "given_last_name", 55, "given_id", "given_tenant").WillReturnResult(sqlmock.NewResult(0, 157))
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.Error(t, err)
					require.Contains(t, err.Error(), "should update single row, but updated 157 rows")
				})

				t.Run("returns error if does not modified any row", func(t *testing.T) {
					// GIVEN
					db, mock := testdb.MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), db)
					defer mock.AssertExpectations(t)

					mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?"+test.AdditionalCondition+" WHERE id = ? AND tenant_id = ?")).
						WithArgs("given_first_name", "given_last_name", 55, "given_id", "given_tenant").WillReturnResult(sqlmock.NewResult(0, 0))
					// WHEN
					err := test.Method(sut)(ctx, givenUser)
					// THEN
					require.Error(t, err)
					if !strings.Contains(err.Error(), apperrors.ShouldBeOwnerMsg) && !strings.Contains(err.Error(), apperrors.ConcurrentUpdateMsg) {
						t.Errorf("unexpected error: %s", err)
					}
				})

				t.Run("returns error if missing persistence context", func(t *testing.T) {
					// WHEN
					err := test.Method(sut)(context.TODO(), User{})
					// THEN
					require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
				})

				t.Run("returns error if entity is nil", func(t *testing.T) {
					// WHEN
					err := test.Method(sut)(context.TODO(), nil)
					// THEN
					require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
				})
			})
		}
	})
}

func TestUpdater(t *testing.T) {
	tests := []struct {
		Name                    string
		AdditionalCondition     string
		AttachAdditionalQueries func(mock testdb.DBMock, m2mTable, tenant string)
		Method                  func(updater repo.Updater) func(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
	}{
		{
			Name:                    "UpdateSingle",
			AttachAdditionalQueries: func(mock testdb.DBMock, m2mTable, tenant string) {},
			Method: func(updater repo.Updater) func(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
				return updater.UpdateSingle
			},
		},
		{
			Name: "UpdateSingleWithVersion",
			AttachAdditionalQueries: func(mock testdb.DBMock, m2mTable, tenant string) {
				mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$2")))).
					WithArgs(appID, tenant).WillReturnRows(testdb.RowWhenObjectExist())
			},
			AdditionalCondition: ", version = version+1",
			Method: func(updater repo.Updater) func(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
				return updater.UpdateSingleWithVersion
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			updater := repo.NewUpdater(appTableName, []string{"name", "description"}, []string{"id"})
			resourceType := resource.Application
			m2mTable, ok := resourceType.TenantAccessTable()
			require.True(t, ok)

			t.Run("success", func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true))", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, appID, tenantID).WillReturnResult(sqlmock.NewResult(0, 1))
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.NoError(t, err)
			})

			t.Run("success when no id column", func(t *testing.T) {
				updater := repo.NewUpdater(appTableName, []string{"name", "description"}, []string{})
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true)", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, tenantID).WillReturnResult(sqlmock.NewResult(0, 1))
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.NoError(t, err)
			})

			t.Run("returns error when operation on db failed", func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true)", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, appID, tenantID).WillReturnError(someError())
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
			})

			t.Run("context properly canceled", func(t *testing.T) {
				db, mock := testdb.MockDatabase(t)
				defer mock.AssertExpectations(t)

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()

				ctx = persistence.SaveToContext(ctx, db)

				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)

				require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
			})

			t.Run("returns non unique error", func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true)", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, appID, tenantID).WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.True(t, apperrors.IsNotUniqueError(err))
			})

			t.Run("returns error if modified more than one row", func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true)", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, appID, tenantID).WillReturnResult(sqlmock.NewResult(0, 157))
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.Error(t, err)
				require.Contains(t, err.Error(), "should update single row, but updated 157 rows")
			})

			t.Run("returns error if does not modified any row", func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				test.AttachAdditionalQueries(mock, m2mTable, tenantID)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ?%s WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true)", appTableName, test.AdditionalCondition, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
					WithArgs(appName, appDescription, appID, tenantID).WillReturnResult(sqlmock.NewResult(0, 0))
				// WHEN
				err := test.Method(updater)(ctx, resourceType, tenantID, fixApp)
				// THEN
				require.Error(t, err)
				if !strings.Contains(err.Error(), apperrors.ShouldBeOwnerMsg) && !strings.Contains(err.Error(), apperrors.ConcurrentUpdateMsg) {
					t.Errorf("unexpected error: %s", err)
				}
			})

			t.Run("returns error if entity does not have tenant access table", func(t *testing.T) {
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)
				// WHEN
				err := test.Method(updater)(ctx, resource.Type("unknown"), tenantID, fixApp)
				// THEN
				assert.Contains(t, err.Error(), "entity unknown does not have access table")
			})

			t.Run("returns error if missing persistence context", func(t *testing.T) {
				// WHEN
				err := test.Method(updater)(context.TODO(), resourceType, tenantID, fixApp)
				// THEN
				require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
			})

			t.Run("returns error if entity is nil", func(t *testing.T) {
				// WHEN
				err := test.Method(updater)(context.TODO(), resourceType, tenantID, nil)
				// THEN
				require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
			})
		})
	}

	t.Run("success for BundleInstanceAuth", func(t *testing.T) {
		updater := repo.NewUpdater(biaTableName, []string{"name", "description"}, []string{"id"})
		resourceType := resource.BundleInstanceAuth
		m2mTable, ok := resourceType.TenantAccessTable()
		require.True(t, ok)
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE %s SET name = ?, description = ? WHERE id = ? AND (id IN (SELECT %s FROM %s WHERE %s = ? AND %s = true) OR owner_id = ?)", biaTableName, repo.M2MResourceIDColumn, m2mTable, repo.M2MTenantIDColumn, repo.M2MOwnerColumn))).
			WithArgs(biaName, biaDescription, biaID, tenantID, tenantID).WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := updater.UpdateSingle(ctx, resourceType, tenantID, fixBIA)
		// THEN
		require.NoError(t, err)
	})

	t.Run("UpdateSingleWithVersion", func(t *testing.T) {
		updater := repo.NewUpdater(appTableName, []string{"name", "description"}, []string{"id"})
		resourceType := resource.Application
		m2mTable, ok := resourceType.TenantAccessTable()
		require.True(t, ok)

		t.Run("tenant does not have owner access should return error", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$2")))).
				WithArgs(appID, tenantID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())

			// WHEN
			err := updater.UpdateSingleWithVersion(ctx, resourceType, tenantID, fixApp)
			// THEN
			require.Error(t, err)
			require.Contains(t, err.Error(), "entity does not exist or caller tenant does not have owner access")
		})

		t.Run("entity is not identifiable should return error", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			// WHEN
			err := updater.UpdateSingleWithVersion(ctx, resourceType, tenantID, struct{}{})
			// THEN
			require.Error(t, err)
			require.Contains(t, err.Error(), "id cannot be empty, check if the entity implements Identifiable")
		})
	})
}
