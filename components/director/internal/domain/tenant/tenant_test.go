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
	tenants := tenant.TenantCtx{InternalID: value, ExternalID: value}

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     string
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.TenantContextKey, tenants),
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
			// WHEN
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

func TestLoadTenantPairFromContext(t *testing.T) {
	value := "foo"
	tenants := tenant.TenantCtx{InternalID: value, ExternalID: value}

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     tenant.TenantCtx
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.TenantContextKey, tenants),
			ExpectedResult:     tenants,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error",
			Context:            context.TODO(),
			ExpectedResult:     tenant.TenantCtx{},
			ExpectedErrMessage: "cannot read tenant from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result, err := tenant.LoadTenantPairFromContext(testCase.Context)

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
	// GIVEN
	value := "foo"
	externalValue := "bar"
	ctx := context.TODO()

	tenants := tenant.TenantCtx{InternalID: value, ExternalID: externalValue}
	// WHEN
	result := tenant.SaveToContext(ctx, value, externalValue)

	// then
	assert.Equal(t, tenants, result.Value(tenant.TenantContextKey))
}

func TestLoadLocalTenantIDFromContext(t *testing.T) {
	localTenantID := "foo"

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     string
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), tenant.LocalTenantIDContextKey, localTenantID),
			ExpectedResult:     localTenantID,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error loading context",
			Context:            context.TODO(),
			ExpectedResult:     "",
			ExpectedErrMessage: "Internal Server Error: cannot read local tenant id from context",
		},
		{
			Name:               "Error in case no local tenant id is provided",
			Context:            context.WithValue(context.TODO(), tenant.LocalTenantIDContextKey, ""),
			ExpectedResult:     "",
			ExpectedErrMessage: "Internal Server Error: local tenant id is required",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result, err := tenant.LoadLocalTenantIDFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveLocalTenantIDToLoadFromContext(t *testing.T) {
	// GIVEN
	localTenantID := "foo"
	ctx := context.TODO()

	// WHEN
	result := tenant.SaveLocalTenantIDToContext(ctx, localTenantID)

	// then
	assert.Equal(t, localTenantID, result.Value(tenant.LocalTenantIDContextKey))
}
