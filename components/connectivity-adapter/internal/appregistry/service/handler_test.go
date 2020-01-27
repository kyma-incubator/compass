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

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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

var (
	testErr = errors.New("Test err")
)

func TestHandler_Create(t *testing.T) {
	// given

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	name := "foo"
	validBody := fmt.Sprintf("{\"name\":\"%s\"}", name)
	svcDetailsModel := model.ServiceDetails{Name: name}
	gqlAppInput := graphql.ApplicationRegisterInput{Name: name}
	expectedGQLReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		result: registerApplication(in: { name: "%s"}) {
			id
		}	
	}`, name))
	successGraphQLResponse := gqlCreateApplicationResponse{}

	testCases := []struct {
		Name                       string
		InputBody                  string
		GraphQLClientFn            func() *automock.GraphQLClient
		ConverterFn                func() *svcautomock.Converter
		ValidatorFn                func() *svcautomock.Validator
		GraphQLRequestBuilderFn    func() *svcautomock.GraphQLRequestBuilder
		LoggerAssertionsFn         func(t *testing.T, hook *test.Hook)
		ExpectedResponseStatusCode int
		ExpectedResponseBody       string
	}{
		{
			Name:      "Success",
			InputBody: validBody,
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*gqlCreateApplicationResponse)
					if !ok {
						t.Logf("Invalid type %T, expected *gqlCreateApplicationResponse", args.Get(2))
						t.FailNow()
					}
					arg.Result.ID = id
				}).Return(nil).Once()
				return cli
			},
			ConverterFn:                SuccessfulDetailsToGQLInputConverterFn(svcDetailsModel, gqlAppInput),
			ValidatorFn:                SuccessfulValidatorFn(svcDetailsModel),
			GraphQLRequestBuilderFn:    SuccessfulRegisterAppGraphQLRequestBuilderFn(gqlAppInput, expectedGQLReq),
			ExpectedResponseStatusCode: http.StatusOK,
			ExpectedResponseBody:       fmt.Sprintf("{\"id\":\"%s\"}\n", id),
		},
		{
			Name:                       "Error - Decoding input",
			InputBody:                  "test",
			GraphQLClientFn:            EmptyGraphQLClientFn(),
			ConverterFn:                EmptyConverterFn(),
			ValidatorFn:                EmptyValidatorFn(),
			GraphQLRequestBuilderFn:    EmptyGraphQLRequestBuilderFn(),
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while unmarshalling service: invalid character 'e' in literal true (expecting 'r')"),
			ExpectedResponseStatusCode: http.StatusBadRequest,
			ExpectedResponseBody:       "{\"code\":4,\"error\":\"while unmarshalling service: invalid character 'e' in literal true (expecting 'r')\"}\n",
		},
		{
			Name:            "Error - Validation",
			InputBody:       validBody,
			GraphQLClientFn: EmptyGraphQLClientFn(),
			ConverterFn:     EmptyConverterFn(),
			ValidatorFn: func() *svcautomock.Validator {
				validator := &svcautomock.Validator{}
				validator.On("Validate", svcDetailsModel).Return(apperrors.WrongInput(testErr.Error())).Once()
				return validator
			},
			GraphQLRequestBuilderFn:    EmptyGraphQLRequestBuilderFn(),
			ExpectedResponseStatusCode: http.StatusBadRequest,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":4,\"error\":\"while validating input: %s\"}\n", testErr.Error()),
		},
		{
			Name:            "Error - Converter",
			InputBody:       validBody,
			GraphQLClientFn: EmptyGraphQLClientFn(),
			ConverterFn: func() *svcautomock.Converter {
				converter := &svcautomock.Converter{}
				converter.On("DetailsToGraphQLInput", svcDetailsModel).Return(graphql.ApplicationRegisterInput{}, testErr).Once()
				return converter
			},
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while converting service input: Test err"),
			ValidatorFn:                SuccessfulValidatorFn(svcDetailsModel),
			GraphQLRequestBuilderFn:    EmptyGraphQLRequestBuilderFn(),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":1,\"error\":\"while converting service input: %s\"}\n", testErr.Error()),
		},
		{
			Name:            "Error - Request Builder",
			InputBody:       validBody,
			GraphQLClientFn: EmptyGraphQLClientFn(),
			ConverterFn:     SuccessfulDetailsToGQLInputConverterFn(svcDetailsModel, gqlAppInput),
			ValidatorFn:     SuccessfulValidatorFn(svcDetailsModel),
			GraphQLRequestBuilderFn: func() *svcautomock.GraphQLRequestBuilder {
				gqlRequestBuilder := &svcautomock.GraphQLRequestBuilder{}
				gqlRequestBuilder.On("RegisterApplicationRequest", gqlAppInput).Return(nil, testErr).Once()
				return gqlRequestBuilder
			},
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while building Application Register input: Test err"),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":1,\"error\":\"while building Application Register input: %s\"}\n", testErr.Error()),
		},
		{
			Name:      "Error - GraphQL Request",
			InputBody: validBody,
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Return(testErr).Once()
				return cli
			},
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while creating service: Test err"),
			ConverterFn:                SuccessfulDetailsToGQLInputConverterFn(svcDetailsModel, gqlAppInput),
			ValidatorFn:                SuccessfulValidatorFn(svcDetailsModel),
			GraphQLRequestBuilderFn:    SuccessfulRegisterAppGraphQLRequestBuilderFn(gqlAppInput, expectedGQLReq),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":1,\"error\":\"while creating service: %s\"}\n", testErr.Error()),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()

			httpReq := fixServiceDetailsRequest(id, strings.NewReader(tc.InputBody))

			cli := tc.GraphQLClientFn()
			cliProvider := fixProviderMock(httpReq, cli)
			gqlBuilder := tc.GraphQLRequestBuilderFn()
			converter := tc.ConverterFn()
			validator := tc.ValidatorFn()

			defer mock.AssertExpectationsForObjects(t, cli, cliProvider, gqlBuilder, converter, validator)

			w := httptest.NewRecorder()

			handler := NewHandler(cliProvider, converter, validator, gqlBuilder, logger)

			// when

			handler.Create(w, httpReq)

			resp, bodyStr, closeBody := readBody(t, w)
			defer closeBody(t)

			// then
			assert.Equal(t, tc.ExpectedResponseStatusCode, resp.StatusCode)
			assert.Equal(t, tc.ExpectedResponseBody, bodyStr)

			if tc.LoggerAssertionsFn != nil {
				tc.LoggerAssertionsFn(t, hook)
			}
		})
	}
}

func TestHandler_Get(t *testing.T) {
	// given

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	name := "foo"
	svcDetailsModel := model.ServiceDetails{Identifier: id, Name: name}
	gqlApp := graphql.ApplicationExt{Application: graphql.Application{ID: id, Name: name}}
	expectedGQLReq := gcli.NewRequest(fmt.Sprintf(`query {
		result: application(id: "%s") {
			id
		}	
	}`, id))
	successGraphQLResponse := gqlGetApplicationResponse{}

	testCases := []struct {
		Name                       string
		GraphQLClientFn            func() *automock.GraphQLClient
		ConverterFn                func() *svcautomock.Converter
		LoggerAssertionsFn         func(t *testing.T, hook *test.Hook)
		ExpectedResponseStatusCode int
		ExpectedResponseBody       string
	}{
		{
			Name: "Success",
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*gqlGetApplicationResponse)
					if !ok {
						t.Logf("Invalid type %T, expected *gqlGetApplicationResponse", args.Get(2))
						t.FailNow()
					}
					arg.Result = &gqlApp
				}).Return(nil).Once()
				return cli
			},
			ConverterFn: func() *svcautomock.Converter {
				converter := &svcautomock.Converter{}
				converter.On("GraphQLToDetailsModel", gqlApp).Return(svcDetailsModel, nil).Once()
				return converter
			},
			ExpectedResponseStatusCode: http.StatusOK,
			ExpectedResponseBody:       fmt.Sprintf("{\"provider\":\"\",\"name\":\"%s\",\"description\":\"\",\"identifier\":\"%s\"}\n", name, id),
		},
		{
			Name: "Error - Not Found",
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Return(nil).Once()
				return cli
			},
			ConverterFn:                EmptyConverterFn(),
			ExpectedResponseStatusCode: http.StatusNotFound,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":2,\"error\":\"entity with ID %s not found\"}\n", id),
		},
		{
			Name: "Error - Converter",
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Run(func(args mock.Arguments) {
					arg, ok := args.Get(2).(*gqlGetApplicationResponse)
					if !ok {
						t.Logf("Invalid type %T, expected *gqlGetApplicationResponse", args.Get(2))
						t.FailNow()
					}
					arg.Result = &gqlApp
				}).Return(nil).Once()
				return cli
			},
			ConverterFn: func() *svcautomock.Converter {
				converter := &svcautomock.Converter{}
				converter.On("GraphQLToDetailsModel", gqlApp).Return(model.ServiceDetails{}, testErr).Once()
				return converter
			},
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while converting model: Test err"),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
			ExpectedResponseBody:       "{\"code\":1,\"error\":\"while converting model: Test err\"}\n",
		},
		{
			Name: "Error - GraphQL Request",
			GraphQLClientFn: func() *automock.GraphQLClient {
				cli := &automock.GraphQLClient{}
				cli.On("Run", context.Background(), expectedGQLReq, &successGraphQLResponse).Return(testErr).Once()
				return cli
			},
			ConverterFn:                EmptyConverterFn(),
			LoggerAssertionsFn:         SingleErrorLoggerAssertions("while getting service: Test err"),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
			ExpectedResponseBody:       "{\"code\":1,\"error\":\"while getting service: Test err\"}\n",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()

			httpReq := fixServiceDetailsRequest(id, strings.NewReader(""))

			cli := tc.GraphQLClientFn()
			cliProvider := fixProviderMock(httpReq, cli)
			gqlRequestBuilder := &svcautomock.GraphQLRequestBuilder{}
			gqlRequestBuilder.On("GetApplicationRequest", id).Return(expectedGQLReq, nil).Once()
			converter := tc.ConverterFn()

			defer mock.AssertExpectationsForObjects(t, cli, cliProvider, gqlRequestBuilder, converter)

			w := httptest.NewRecorder()

			handler := NewHandler(cliProvider, converter, nil, gqlRequestBuilder, logger)

			// when

			handler.Get(w, httpReq)

			resp, bodyStr, closeBody := readBody(t, w)
			defer closeBody(t)

			// then
			assert.Equal(t, tc.ExpectedResponseStatusCode, resp.StatusCode)
			assert.Equal(t, tc.ExpectedResponseBody, bodyStr)

			if tc.LoggerAssertionsFn != nil {
				tc.LoggerAssertionsFn(t, hook)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	// given

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	httpReq := fixServiceDetailsRequest(id, strings.NewReader(""))
	expectedGQLReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			id
		}	
	}`, id))
	notFoundErr := fmt.Errorf("graphql: while getting Application with ID %s: Object was not found", id)

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
			Name:                       "Error - Not Found",
			GraphQLClientErr:           notFoundErr,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":2,\"error\":\"entity with ID %s not found\"}\n", id),
			ExpectedResponseStatusCode: http.StatusNotFound,
		},
		{
			Name: "Error - Internal",
			LoggerAssertionsFn: func(t *testing.T, hook *test.Hook) {
				assert.Equal(t, 1, len(hook.AllEntries()))
				entry := hook.LastEntry()
				require.NotNil(t, entry)
				assert.Equal(t, log.ErrorLevel, entry.Level)
				assert.Equal(t, id, entry.Data["ID"])
				assert.Equal(t, fmt.Sprintf("while deleting service: %s", testErr.Error()), entry.Message)
			},
			GraphQLClientErr:           testErr,
			ExpectedResponseBody:       fmt.Sprintf("{\"code\":1,\"error\":\"%s\"}\n", testErr.Error()),
			ExpectedResponseStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()

			cli := &automock.GraphQLClient{}
			cli.On("Run", context.Background(), expectedGQLReq, nil).Return(tc.GraphQLClientErr).Once()
			cliProvider := fixProviderMock(httpReq, cli)
			gqlBuilder := &svcautomock.GraphQLRequestBuilder{}
			gqlBuilder.On("UnregisterApplicationRequest", id).Return(expectedGQLReq).Once()

			defer mock.AssertExpectationsForObjects(t, cli, cliProvider, gqlBuilder)

			w := httptest.NewRecorder()

			handler := NewHandler(cliProvider, nil, nil, gqlBuilder, logger)

			// when
			handler.Delete(w, httpReq)

			resp, bodyStr, closeBody := readBody(t, w)
			defer closeBody(t)

			// then

			assert.Equal(t, tc.ExpectedResponseStatusCode, resp.StatusCode)
			assert.Equal(t, tc.ExpectedResponseBody, bodyStr)

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

func fixProviderMock(httpReq *http.Request, gqlClient *automock.GraphQLClient) *automock.Provider {
	cliProvider := &automock.Provider{}
	cliProvider.On("GQLClient", httpReq).Return(gqlClient).Maybe() // In not all cases it will be fired

	return cliProvider
}

func readBody(t *testing.T, w *httptest.ResponseRecorder) (*http.Response, string, func(t *testing.T)) {
	resp := w.Result()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(bodyBytes), func(t *testing.T) {
		err := resp.Body.Close()
		require.NoError(t, err)
	}
}
