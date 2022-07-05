package accessstrategy_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/stretchr/testify/require"
)

func TestOpenAccessStrategy(t *testing.T) {
	testURL := "http://test"

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, req.Method, http.MethodGet)
		require.Equal(t, req.URL.String(), testURL)
		return expectedResp, nil
	})

	cerCache := certloader.NewCertificateCache()
	provider := accessstrategy.NewDefaultExecutorProvider(cerCache)
	executor, err := provider.Provide(accessstrategy.OpenAccessStrategy)
	require.NoError(t, err)

	resp, err := executor.Execute(context.TODO(), client, testURL, "")
	require.NoError(t, err)
	require.Equal(t, expectedResp, resp)
}
