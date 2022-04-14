package log

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log/automock"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

const testOperationName = "myOperation"

func TestGraphQlRequestDetailsLogging(t *testing.T) {
	t.Run("log GQL query details", func(t *testing.T) {
		ctx := ContextWithMdc(context.Background())
		ctx = graphql.WithOperationContext(ctx, &graphql.OperationContext{
			Operation: &ast.OperationDefinition{
				Operation: ast.Query,
				SelectionSet: []ast.Selection{&ast.Field{
					Name: testOperationName,
				}},
			},
		})
		mutationInstrumenter := &automock.GraphqlQueryRequestInstrumenter{}
		mutationInstrumenter.On("InstrumentGraphqlQueryRequest", mock.Anything, mock.Anything).Once()

		middleware := NewGqlLoggingInterceptor(mutationInstrumenter)
		middleware.InterceptOperation(ctx, func(ctx context.Context) graphql.ResponseHandler {
			return nil
		})

		mdc := MdcFromContext(ctx)
		opType, ok := mdc.mdc[logKeyOperationType]
		if !ok || opType.(string) != string(ast.Query) {
			t.Errorf("The GraphQL operation type is missing or incorrect. Expected=%v; Actual=%v", ast.Query, opType)
		}

		opName, ok := mdc.mdc[logKeySelectionSet]
		if !ok || opName.(string) != testOperationName {
			t.Errorf("The GraphQL operation name is missing or incorrect: %v", opName)
		}
	})
}
