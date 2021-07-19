package ownertenant_test

import (
	"context"
	"testing"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/ownertenant"
	"github.com/kyma-incubator/compass/components/director/internal/ownertenant/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestInterceptField(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	parentTenantID := "56384145-8b6c-4394-855a-c6e55619448b"
	ownerTenantID := "68182a53-ba3a-4cef-8932-055d5c7e0e91"
	ownerTenantExternalID := "916aa40c-fe7b-4edb-bc34-158c728326b6"

	ownerTenant := &model.BusinessTenantMapping{
		ID:             ownerTenantID,
		Name:           "ownerTenantName",
		ExternalTenant: ownerTenantExternalID,
		Parent:         parentTenantID,
		Type:           "account",
		Provider:       "Compass",
		Status:         "Active",
	}

	entityID := "dc9964c8-4e81-4a58-bc7a-44e788ee1fdd"

	fixFieldCtx := &gqlgen.FieldContext{
		Object: ownertenant.MutationObject,
		Field: gqlgen.CollectedField{
			Field: &ast.Field{
				Arguments: ast.ArgumentList{
					&ast.Argument{
						Value: &ast.Value{
							Raw: entityID,
							Definition: &ast.Definition{
								Name:    "ID",
								BuiltIn: true,
							},
						},
					},
				},
			},
		},
	}

	successContextProviderFn := func() context.Context {
		ctx := context.TODO()
		ctx = gqlgen.WithFieldContext(ctx, fixFieldCtx)
		ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
		return ctx
	}

	assertTenantNotModifiedResolverFn := func() gqlgen.Resolver {
		return func(ctx context.Context) (res interface{}, err error) {
			tnt, err := tenant.LoadFromContext(ctx)
			require.NoError(t, err)
			require.Equal(t, parentTenantID, tnt)
			return nil, nil
		}
	}

	assertNextNotCalledResolverFn := func() gqlgen.Resolver {
		return func(ctx context.Context) (res interface{}, err error) {
			t.FailNow() // Assert never called
			return nil, nil
		}
	}

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TenantIndexRepoFn func() *automock.TenantIndexRepository
		TenantRepoFn      func() *automock.TenantRepository
		ContextProviderFn func() context.Context
		NextResolverFn    func() gqlgen.Resolver
		ExpectedErr       string
	}{
		{
			Name:            "Mutation with parent ID executed by parent tenant should impersonate owner tenant",
			TransactionerFn: txGen.ThatSucceeds,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return(ownerTenantID, nil).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), ownerTenantID).Return(ownerTenant, nil).Once()
				return repo
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn: func() gqlgen.Resolver {
				return func(ctx context.Context) (res interface{}, err error) {
					tnt, err := tenant.LoadFromContext(ctx)
					require.NoError(t, err)
					require.Equal(t, ownerTenantID, tnt)
					return nil, nil
				}
			},
		},
		{
			Name:            "Mutation with parent ID in graphQL variable executed by parent tenant should impersonate owner tenant",
			TransactionerFn: txGen.ThatSucceeds,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return(ownerTenantID, nil).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), ownerTenantID).Return(ownerTenant, nil).Once()
				return repo
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: ownertenant.MutationObject,
					Args: map[string]interface{}{
						"id": entityID,
					},
					Field: gqlgen.CollectedField{
						Field: &ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Value: &ast.Value{
										Raw: "id",
										Definition: &ast.Definition{
											Name:    "ID",
											BuiltIn: true,
										},
									},
								},
							},
						},
					},
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: func() gqlgen.Resolver {
				return func(ctx context.Context) (res interface{}, err error) {
					tnt, err := tenant.LoadFromContext(ctx)
					require.NoError(t, err)
					require.Equal(t, ownerTenantID, tnt)
					return nil, nil
				}
			},
		},
		{
			Name:            "Mutation with no GUID or GraphQL variable provided for ID should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: ownertenant.MutationObject,
					Args: map[string]interface{}{
						"non-existing": "non-existing",
					},
					Field: gqlgen.CollectedField{
						Field: &ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Value: &ast.Value{
										Raw: "non-guid-or-variable",
										Definition: &ast.Definition{
											Name:    "ID",
											BuiltIn: true,
										},
									},
								},
							},
						},
					},
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Nil GraphQL Field context should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "GraphQL Object is not Mutation should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: "Query",
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Mutation without ID argument should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: ownertenant.MutationObject,
					Field: gqlgen.CollectedField{
						Field: &ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Value: &ast.Value{
										Raw: entityID,
										Definition: &ast.Definition{
											Name:    "in",
											BuiltIn: false,
										},
									},
								},
							},
						},
					},
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Mutation with empty ID argument should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: ownertenant.MutationObject,
					Field: gqlgen.CollectedField{
						Field: &ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Value: &ast.Value{
										Raw: "",
										Definition: &ast.Definition{
											Name:    "ID",
											BuiltIn: true,
										},
									},
								},
							},
						},
					},
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Mutation with two ID arguments should return error",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, &gqlgen.FieldContext{
					Object: ownertenant.MutationObject,
					Field: gqlgen.CollectedField{
						Field: &ast.Field{
							Arguments: ast.ArgumentList{
								&ast.Argument{
									Value: &ast.Value{
										Raw: entityID,
										Definition: &ast.Definition{
											Name:    "ID",
											BuiltIn: true,
										},
									},
								},
								&ast.Argument{
									Value: &ast.Value{
										Raw: entityID + "2",
										Definition: &ast.Definition{
											Name:    "ID",
											BuiltIn: true,
										},
									},
								},
							},
						},
					},
				})
				ctx = tenant.SaveToContext(ctx, parentTenantID, parentTenantID)
				return ctx
			},
			NextResolverFn: assertNextNotCalledResolverFn,
			ExpectedErr:    "More than one argument with type ID is provided for the mutation",
		},
		{
			Name:            "No calling tenant in context should proceed without any modification",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: func() context.Context {
				ctx := context.TODO()
				ctx = gqlgen.WithFieldContext(ctx, fixFieldCtx)
				return ctx
			},
			NextResolverFn: func() gqlgen.Resolver {
				return func(ctx context.Context) (res interface{}, err error) {
					tnt, err := tenant.LoadFromContext(ctx)
					require.Error(t, err)
					require.Empty(t, tnt)
					return nil, nil
				}
			},
		},
		{
			Name:            "Transaction Begin failure should return error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				return &automock.TenantIndexRepository{}
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertNextNotCalledResolverFn,
			ExpectedErr:       testErr.Error(),
		},
		{
			Name:            "Getting owner tenant from index error should fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return("", testErr).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertNextNotCalledResolverFn,
			ExpectedErr:       testErr.Error(),
		},
		{
			Name:            "Owner tenant Not Found in index should proceed without any modifications",
			TransactionerFn: txGen.ThatSucceeds,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return("", apperrors.NewNotFoundError(resource.TenantIndex, ownerTenantID)).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Owner tenant Not Found in index and tx commit fails should return error",
			TransactionerFn: txGen.ThatFailsOnCommit,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return("", apperrors.NewNotFoundError(resource.TenantIndex, ownerTenantID)).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertNextNotCalledResolverFn,
			ExpectedErr:       testErr.Error(),
		},
		{
			Name:            "Owner tenant is the same as caller tenant should proceed without any modifications",
			TransactionerFn: txGen.ThatSucceeds,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return(parentTenantID, nil).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				return &automock.TenantRepository{}
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertTenantNotModifiedResolverFn,
		},
		{
			Name:            "Getting owner tenant failure should return error",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return(ownerTenantID, nil).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), ownerTenantID).Return(nil, testErr).Once()
				return repo
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertNextNotCalledResolverFn,
			ExpectedErr:       testErr.Error(),
		},
		{
			Name:            "Tx commit failure should return error",
			TransactionerFn: txGen.ThatFailsOnCommit,
			TenantIndexRepoFn: func() *automock.TenantIndexRepository {
				repo := &automock.TenantIndexRepository{}
				repo.On("GetOwnerTenantByResourceID", txtest.CtxWithDBMatcher(), parentTenantID, entityID).Return(ownerTenantID, nil).Once()
				return repo
			},
			TenantRepoFn: func() *automock.TenantRepository {
				repo := &automock.TenantRepository{}
				repo.On("Get", txtest.CtxWithDBMatcher(), ownerTenantID).Return(ownerTenant, nil).Once()
				return repo
			},
			ContextProviderFn: successContextProviderFn,
			NextResolverFn:    assertNextNotCalledResolverFn,
			ExpectedErr:       testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			tenantIndexRepo := testCase.TenantIndexRepoFn()
			tenantRepo := testCase.TenantRepoFn()

			middleware := ownertenant.NewMiddleware(transact, tenantIndexRepo, tenantRepo)

			// when
			_, err := middleware.InterceptField(testCase.ContextProviderFn(), testCase.NextResolverFn())

			// then
			if len(testCase.ExpectedErr) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr)
			}

			mock.AssertExpectationsForObjects(t, persistTx, transact, tenantIndexRepo, tenantRepo)
		})
	}
}
