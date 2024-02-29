package accessstrategy_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/stretchr/testify/require"
)

const (
	externalClientCertSecretName = "resource-name1"
)

func TestOpenAccessStrategy(t *testing.T) {
	testURL := "http://test"
	headerKey := "key"
	headerValue := "value"

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, req.Method, http.MethodGet)
		require.Equal(t, req.URL.String(), testURL)
		require.Equal(t, req.Header.Get(headerKey), headerValue)
		return expectedResp, nil
	})

	cerCache := credloader.NewCertificateCache()
	provider := accessstrategy.NewDefaultExecutorProvider(cerCache, externalClientCertSecretName)
	executor, err := provider.Provide(accessstrategy.OpenAccessStrategy)
	require.NoError(t, err)
	headers := &sync.Map{}
	headers.Store(headerKey, headerValue)

	resp, err := executor.Execute(context.TODO(), client, testURL, "", headers)

	require.NoError(t, err)
	require.Equal(t, expectedResp, resp)
}
