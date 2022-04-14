package log

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	logKeyOperationType string = "gql-op-type"
	logKeySelectionSet  string = "gql-operation"
)

// GraphqlQueryRequestInstrumenter collects metrics for different client and auth flows.
//go:generate mockery --name=GraphqlQueryRequestInstrumenter --output=automock --outpkg=automock --case=underscore
type GraphqlQueryRequestInstrumenter interface {
	InstrumentGraphqlQueryRequest(queryType, queryOperation string)
}

// NewGqlLoggingInterceptor creates a new opsInterceptor instance
func NewGqlLoggingInterceptor(graphqlQueryRequestInstrumenter GraphqlQueryRequestInstrumenter) *opsInterceptor {
	return &opsInterceptor{graphqlQueryRequestInstrumenter: graphqlQueryRequestInstrumenter}
}

type opsInterceptor struct {
	graphqlQueryRequestInstrumenter GraphqlQueryRequestInstrumenter
}

func (m *opsInterceptor) ExtensionName() string {
	return "Logging Interceptor"
}

func (m *opsInterceptor) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (m *opsInterceptor) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	if mdc := MdcFromContext(ctx); nil != mdc {
		opsCtx := graphql.GetOperationContext(ctx)
		mdc.Set(logKeyOperationType, string(opsCtx.Operation.Operation))

		selectionSet := ""

		for _, selection := range opsCtx.Operation.SelectionSet {
			if field, ok := selection.(*ast.Field); ok {
				if len(selectionSet) != 0 {
					selectionSet += ","
				}

				m.graphqlQueryRequestInstrumenter.InstrumentGraphqlQueryRequest(string(opsCtx.Operation.Operation), field.Name)

				selectionSet += field.Name
			}
		}
		mdc.Set(logKeySelectionSet, selectionSet)
	}

	return next(ctx)
}
