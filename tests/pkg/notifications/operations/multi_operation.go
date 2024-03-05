package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

type MultiOperation struct {
	operations []Operation
	asserters  []asserters.Asserter
}

func NewMultiOperation() *MultiOperation {
	return &MultiOperation{}
}

func (o *MultiOperation) WithOperation(operation Operation) *MultiOperation {
	o.operations = append(o.operations, operation)
	return o
}

func (o *MultiOperation) WithAsserters(asserters ...asserters.Asserter) *MultiOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *MultiOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	for _, operation := range o.operations {
		operation.Execute(t, ctx, gqlClient)
	}

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *MultiOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	// execute the cleanups in reverse order to mimic defer execution
	for i := 0; i < len(o.operations); i++ {
		o.operations[i].Cleanup(t, ctx, gqlClient)
	}
}

func (o *MultiOperation) Operation() Operation {
	return o
}
