package appdetails_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var value graphql.ApplicationExt = graphql.ApplicationExt{Application: graphql.Application{Name: "foo"}}

func TestLoadFromContext(t *testing.T) {
	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult     *graphql.ApplicationExt
		ExpectedErrMessage string
	}{
		{
			Name:               "Success",
			Context:            context.WithValue(context.TODO(), appdetails.AppDetailsContextKey{}, value),
			ExpectedResult:     &value,
			ExpectedErrMessage: "",
		},
		{
			Name:               "Error",
			Context:            context.TODO(),
			ExpectedResult:     nil,
			ExpectedErrMessage: "cannot read Application details from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result, err := appdetails.LoadFromContext(testCase.Context)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			if testCase.ExpectedResult != nil {
				assert.Equal(t, *testCase.ExpectedResult, result)
			}
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// given
	ctx := context.TODO()

	// when
	result := appdetails.SaveToContext(ctx, value)

	// then
	assert.Equal(t, value, result.Value(appdetails.AppDetailsContextKey{}))
}
