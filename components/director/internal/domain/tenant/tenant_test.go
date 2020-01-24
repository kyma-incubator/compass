package tenant_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromContext(t *testing.T) {
	value := "foo"

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     string
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.TenantContextKey, value),
			ExpectedResult:     value,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error",
			Context:            context.TODO(),
			ExpectedResult:     "",
			ExpectedErrMessage: "cannot read tenant from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result, err := tenant.LoadFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// given
	value := "foo"
	ctx := context.TODO()

	// when
	result := tenant.SaveToContext(ctx, value)

	// then
	assert.Equal(t, value, result.Value(tenant.TenantContextKey))
}
