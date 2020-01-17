package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/mock"

	"github.com/gorilla/mux"
	svcautomock "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Delete(t *testing.T) {
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	httpReq := fixServiceDetailsRequest(id, strings.NewReader(""))
	expectedGQLReq := prepareUnregisterApplicationRequest(id)
	notFoundErr := fmt.Errorf("graphql: while getting Application with ID %s: Object was not found", id)
	internalErr := errors.New("Post http://127.0.0.1:3000/graphql: dial tcp 127.0.0.1:3000: connect: connection refused")

	testCases := []struct {
		Name                       string
		GraphQLClientErr           error
		LoggerAssertionsFn         func(t *testing.T, hook *test.Hook)
		ExpectedResponseStatusCode int
		ExpectedResponseBody       string
	}{
		{
			Name:                       "Success",
			GraphQLClientErr:           nil,
			ExpectedResponseBody:       "",
			ExpectedResponseStatusCode: http.StatusNoContent,
		},
		{
			Name:                       "Not found",
			GraphQLClientErr:           notFoundErr,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":2,\"error\":\"entity with ID %s not found\"}\n", id),
			ExpectedResponseStatusCode: http.StatusNotFound,
		},
		{
			Name: "Error",
			LoggerAssertionsFn: func(t *testing.T, hook *test.Hook) {
				assert.Equal(t, 1, len(hook.AllEntries()))
				entry := hook.LastEntry()
				require.NotNil(t, entry)
				assert.Equal(t, log.ErrorLevel, entry.Level)
				assert.Equal(t, id, entry.Data["ID"])
				assert.Equal(t, fmt.Sprintf("while deleting service: %s", internalErr.Error()), entry.Message)
			},
			GraphQLClientErr:           internalErr,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":1,\"error\":\"%s\"}\n", internalErr.Error()),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()

			cli, cliProvider := fixClientAndProviderMocks(httpReq, expectedGQLReq, tc.GraphQLClientErr)
			converter := &svcautomock.Converter{}
			defer mock.AssertExpectationsForObjects(t, cli, cliProvider, converter)

			w := httptest.NewRecorder()

			handler := NewHandler(cliProvider, converter, logger)

			handler.Delete(w, httpReq)

			resp := w.Result()
			defer closeRequestBody(t, resp)

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedResponseStatusCode, resp.StatusCode)
			assert.Equal(t, tc.ExpectedResponseBody, string(bodyBytes))

			if tc.LoggerAssertionsFn != nil {
				tc.LoggerAssertionsFn(t, hook)
			}
		})
	}
}

func fixServiceDetailsRequest(id string, body io.Reader) *http.Request {
	// Method and URL doesn't matter, as we rely on gorilla/mux for routing.
	// In scope of Handler, we don't check them.
	req := httptest.NewRequest("Anything", "http://doesnt.really/matter", body)
	req = mux.SetURLVars(req, map[string]string{serviceIDVarKey: id})
	return req
}

func closeRequestBody(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}

func fixClientAndProviderMocks(httpReq *http.Request, expectedGQLReq *gcli.Request, gqlCliError error) (*automock.GraphQLClient, *automock.Provider) {
	cli := &automock.GraphQLClient{}
	cli.On("Run", context.Background(), expectedGQLReq, nil).Return(gqlCliError).Once()

	cliProvider := &automock.Provider{}
	cliProvider.On("GQLClient", httpReq).Return(cli).Once()

	return cli, cliProvider
}
