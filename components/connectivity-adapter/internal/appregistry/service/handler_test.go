package service

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const url = "http://doesnt.really/matter"

func TestHandler_Delete(t *testing.T) {
	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req := fixRequest(id, strings.NewReader(""))

	w := httptest.NewRecorder()

	gqlReq := prepareUnregisterApplicationRequest(id)

	cli := &automock.GraphQLClient{}
	cli.On("Run", context.Background(), gqlReq, nil).Return(nil).Once()

	cliProvider := &automock.Provider{}
	cliProvider.On("GQLClient", req).Return(cli).Once()
	defer cli.AssertExpectations(t)
	defer cliProvider.AssertExpectations(t)

	handler := NewHandler(cliProvider)

	handler.Delete(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "", string(bodyBytes))
}

func fixRequest(id string, body io.Reader) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, url, body)
	req = mux.SetURLVars(req, map[string]string{serviceIDVarKey: id})
	return req
}
