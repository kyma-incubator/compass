package adapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter"
	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Run("returns token", func(t *testing.T) {
		// GIVEN
		givenRequestData := adapter.RequestData{
			Tenant: fixTenant(),
			Application: graphql.Application{
				BaseEntity: &graphql.BaseEntity{
					ID: fixAppID(),
				},
				Name: fixAppName(),
			},
		}
		mockSrv := httptest.NewServer(&correctHandler{t: t})
		defer mockSrv.Close()

		cli := adapter.NewClient(http.DefaultClient, adapter.Mapping{
			TemplateExternalURL:       fmt.Sprintf("%s/{{.Tenant}}", mockSrv.URL),
			TemplateHeaders:           `{"Content-Type":["application/json"], "application-id":["{{.Application.ID}} "] }`,
			TemplateJSONBody:          `{"applicationName":"{{.Application.Name}}"}`,
			TemplateTokenFromResponse: `{{ .integrationToken }}`,
		})
		// WHEN
		actualToken, err := cli.Do(context.TODO(), givenRequestData)
		// THEN
		require.NoError(t, err)
		assert.NotNil(t, actualToken)
		assert.Equal(t, fixIntegrationToken(), actualToken.Token)
	})

	t.Run("fails when cannot access URL", func(t *testing.T) {
		// GIVEN
		mockDoer := &automock.HTTPDoer{}
		defer mockDoer.AssertExpectations(t)
		mockDoer.On("Do", mock.Anything).Return(nil, fixError())
		cli := adapter.NewClient(mockDoer, adapter.Mapping{TemplateExternalURL: "http://compass.io"})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "while performing request: fix error")
	})

	t.Run("fails when template URL has incorrect syntax", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateExternalURL: "{{.wrongSyntax"})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: url:1: unclosed action")
	})

	t.Run("fails when template URL access not existing field", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateExternalURL: "https://kyma-project.io/{{.missingParameter}}"})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: url:1:26: executing \"url\" at <.missingParameter>: can't evaluate field missingParameter in type adapter.RequestData")
	})
	t.Run("fails when headers template has incorrect syntax", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateHeaders: `{"aaa":"{{.wrongSyntax"}`})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: header:1: unexpected bad character U+0022 '\"' in command")
	})

	t.Run("fails when headers template access not existing field", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateHeaders: `{"aaa":"{{.missing}}"}`})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: header:1:10: executing \"header\" at <.missing>: can't evaluate field missing in type adapter.RequestData")
	})

	t.Run("fails when headers template has wrong syntax type", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateHeaders: `["a","b","c"]`})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "while unmarshalling headers from JSON to map: json: cannot unmarshal array into Go value of type map[string][]string")
	})

	t.Run("fails when header is not an JSON", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateHeaders: `<xml></xml>`})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "while unmarshalling headers from JSON to map: invalid character '<' looking for beginning of value")
	})

	t.Run("fails when body template has incorrect syntax", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateJSONBody: `{"aaa":"{{.wrongSyntax"}`})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: body:1: unexpected bad character U+0022 '\"' in command")
	})

	t.Run("fails when body template access not existing field", func(t *testing.T) {
		// GIVEN
		cli := adapter.NewClient(nil, adapter.Mapping{TemplateJSONBody: `{"aaa":"{{.missing}}"}`})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "template: body:1:10: executing \"body\" at <.missing>: can't evaluate field missing in type adapter.RequestData")
	})

	t.Run("fails when response template has incorrect syntax", func(t *testing.T) {
		// GIVEN
		mockSrv := httptest.NewServer(&correctHandler{t: t})
		defer mockSrv.Close()

		cli := adapter.NewClient(http.DefaultClient, adapter.Mapping{
			TemplateExternalURL:       fmt.Sprintf("%s/{{.Tenant}}", mockSrv.URL),
			TemplateHeaders:           `{"Content-Type":["application/json"], "application-id":["{{.Application.ID}} "] }`,
			TemplateJSONBody:          `{"applicationName":"{{.Application.Name}}"}`,
			TemplateTokenFromResponse: `{{ .wrongSyntax`,
		})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{
			Tenant: fixTenant(),
			Application: graphql.Application{
				BaseEntity: &graphql.BaseEntity{
					ID: fixAppID(),
				},
				Name: fixAppName(),
			},
		})
		// THEN
		assert.EqualError(t, err, "template: response:1: unclosed action")
	})

	t.Run("fails when response template access not existing fields", func(t *testing.T) {
		// GIVEN
		mockSrv := httptest.NewServer(&correctHandler{t: t})
		defer mockSrv.Close()

		cli := adapter.NewClient(http.DefaultClient, adapter.Mapping{
			TemplateExternalURL:       fmt.Sprintf("%s/{{.Tenant}}", mockSrv.URL),
			TemplateHeaders:           `{"Content-Type":["application/json"], "application-id":["{{.Application.ID}} "] }`,
			TemplateJSONBody:          `{"applicationName":"{{.Application.Name}}"}`,
			TemplateTokenFromResponse: `{{ .missingField }}`,
		})
		// WHEN
		_, err := cli.Do(nil, adapter.RequestData{
			Tenant: fixTenant(),
			Application: graphql.Application{
				BaseEntity: &graphql.BaseEntity{
					ID: fixAppID(),
				},
				Name: fixAppName(),
			},
		})
		// THEN
		assert.EqualError(t, err, "template: response:1:3: executing \"response\" at <.missingField>: map has no entry for key \"missingField\"")
	})

	t.Run("fails when server return non-JSON", func(t *testing.T) {
		// GIVEN
		givenResponse := &http.Response{}
		bodyBuf := new(bytes.Buffer)
		bodyBuf.WriteString("<xml></xml>")
		givenResponse.Body = ioutil.NopCloser(bodyBuf)
		givenResponse.StatusCode = http.StatusOK
		mockDoer := &automock.HTTPDoer{}
		defer mockDoer.AssertExpectations(t)

		mockDoer.On("Do", mock.Anything).Return(givenResponse, nil)
		cli := adapter.NewClient(mockDoer, adapter.Mapping{})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "while unmarshalling response body: invalid character '<' looking for beginning of value")
	})

	t.Run("fails when server returns status code other than HTTP 200", func(t *testing.T) {
		// GIVEN
		buf := bytes.NewBufferString("detailed message")
		givenResponse := &http.Response{StatusCode: http.StatusTeapot, Body: ioutil.NopCloser(buf)}
		mockDoer := &automock.HTTPDoer{}
		defer mockDoer.AssertExpectations(t)

		mockDoer.On("Do", mock.Anything).Return(givenResponse, nil)
		cli := adapter.NewClient(mockDoer, adapter.Mapping{})
		// WHEN
		_, err := cli.Do(context.TODO(), adapter.RequestData{})
		// THEN
		assert.EqualError(t, err, "wrong status code, got: [418], body: [detailed message]")
	})
}

func fixError() error {
	return errors.New("fix error")
}

func fixTenant() string {
	return "my-tenant"
}

func fixAppID() string {
	return "applicationID"
}

func fixAppName() string {
	return "applicationName"
}

func fixIntegrationToken() string {
	return "integration-token"
}

type correctHandler struct {
	t *testing.T
}

func (th *correctHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	assert.Equal(th.t, http.MethodPost, req.Method)
	assert.Equal(th.t, fmt.Sprintf("/%s", fixTenant()), req.URL.String())
	assert.Equal(th.t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(th.t, fixAppID(), req.Header.Get("application-id"))

	defer func() {
		require.NoError(th.t, req.Body.Close())
	}()
	b, err := ioutil.ReadAll(req.Body)
	require.NoError(th.t, err)
	actualBody := map[string]interface{}{}
	err = json.Unmarshal(b, &actualBody)
	require.NoError(th.t, err)

	assert.Len(th.t, actualBody, 1)
	assert.Equal(th.t, fixAppName(), actualBody["applicationName"])
	_, err = rw.Write([]byte(fmt.Sprintf(`{"provider":"test","integrationToken":"%s"}`, fixIntegrationToken())))
	rw.Header().Set("Content-Type", "application/json")
	require.NoError(th.t, err)

}
