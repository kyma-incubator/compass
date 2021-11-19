package accessstrategy_test

import (
	"context"
	"net/http"
	"testing"

	accessstrategy2 "github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/stretchr/testify/require"
)

func TestOpenAccessStrategy(t *testing.T) {
	testURL := "http://test"

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, req.Method, http.MethodGet)
		require.Equal(t, req.URL.String(), testURL)
		return expectedResp, nil
	})

	provider := accessstrategy2.NewDefaultExecutorProvider()
	executor, err := provider.Provide(accessstrategy2.OpenAccessStrategy)
	require.NoError(t, err)

	resp, err := executor.Execute(context.Background(), client, testURL)
	require.NoError(t, err)
	require.Equal(t, expectedResp, resp)
}
