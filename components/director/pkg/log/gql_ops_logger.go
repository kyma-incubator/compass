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

// NewGqlLoggingInterceptor creates a new opsInterceptor instance
func NewGqlLoggingInterceptor() *opsInterceptor {
	return &opsInterceptor{}
}

type opsInterceptor struct {
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

				selectionSet += field.Name
			}
		}
		mdc.Set(logKeySelectionSet, selectionSet)
	}

	return next(ctx)
}
