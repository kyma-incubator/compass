package metrics

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphqlRequestInstrumenter collects metrics for GraphQL requests usage per enpoint.
type GraphqlRequestInstrumenter interface {
	InstrumentGraphqlRequest(queryType, queryOperation string)
}

// NewInstrumentGraphqlRequestInterceptor creates a new InstrumentGraphqlRequestInterceptor instance
func NewInstrumentGraphqlRequestInterceptor(graphqlRequestInstrumenter GraphqlRequestInstrumenter) *instrumentGraphqlRequestInterceptor {
	return &instrumentGraphqlRequestInterceptor{graphqlRequestInstrumenter: graphqlRequestInstrumenter}
}

type instrumentGraphqlRequestInterceptor struct {
	graphqlRequestInstrumenter GraphqlRequestInstrumenter
}

func (m *instrumentGraphqlRequestInterceptor) ExtensionName() string {
	return "GraphQL Metrics Request Interceptor Interceptor"
}

func (m *instrumentGraphqlRequestInterceptor) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (m *instrumentGraphqlRequestInterceptor) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	opsCtx := graphql.GetOperationContext(ctx)

	for _, selection := range opsCtx.Operation.SelectionSet {
		if field, ok := selection.(*ast.Field); ok {
			m.graphqlRequestInstrumenter.InstrumentGraphqlRequest(string(opsCtx.Operation.Operation), field.Name)
		}
	}

	return next(ctx)
}
