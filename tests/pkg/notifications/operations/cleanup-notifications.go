package operations

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type CleanupNotificationsOperation struct {
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
	asserters                          []asserters.Asserter
}

func NewCleanupNotificationsOperation() *CleanupNotificationsOperation {
	return &CleanupNotificationsOperation{}
}

func (o *CleanupNotificationsOperation) WithExternalServicesMockMtlsSecuredURL(externalServicesMockMtlsSecuredURL string) *CleanupNotificationsOperation {
	o.externalServicesMockMtlsSecuredURL = externalServicesMockMtlsSecuredURL
	return o
}

func (o *CleanupNotificationsOperation) WithHTTPClient(client *http.Client) *CleanupNotificationsOperation {
	o.client = client
	return o
}

func (o *CleanupNotificationsOperation) WithAsserters(asserters ...asserters.Asserter) *CleanupNotificationsOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *CleanupNotificationsOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	req, err := http.NewRequest(http.MethodDelete, o.externalServicesMockMtlsSecuredURL+"/formation-callback/cleanup", nil)
	require.NoError(t, err)
	resp, err := o.client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *CleanupNotificationsOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
}

func (o *CleanupNotificationsOperation) Operation() Operation {
	return o
}
