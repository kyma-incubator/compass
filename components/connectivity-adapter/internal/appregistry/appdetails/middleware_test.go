package appdetails_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	appName := "app-name"
	req := httptest.NewRequest("GET", fmt.Sprintf("/%s", appName), strings.NewReader(""))
	app := fixApplicationExt(appName)

	testErr := errors.New("test error")
	t.Run("success", func(t *testing.T) {
		//GIVEN
		appPage := graphql.ApplicationPageExt{ApplicationPage: graphql.ApplicationPage{},
			Data: []*graphql.ApplicationExt{&app}}

		gqlQueryBuilder := director.NewClient(nil, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		expectedQuery := gqlQueryBuilder.GetApplicationsByNameRequest(appName)

		var apps appdetails.GqlSuccessfulAppPage
		gqlClient := &automock.GraphQLClient{}
		gqlClient.On("Run", mock.Anything, expectedQuery, &apps).Run(injectDirectorResponse(t, appPage)).Return(nil)

		cliProvider := &automock.Provider{}
		cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)

		appMiddleware := appdetails.NewApplicationMiddleware(cliProvider)

		router := mux.NewRouter()
		router.HandleFunc("/{app-name}", assertAppInContext(t, app))
		router.Use(appMiddleware.Middleware)
		rw := httptest.NewRecorder()

		//WHEN
		router.ServeHTTP(rw, req)

		//THEN
		assert.Equal(t, http.StatusOK, rw.Code)
		mock.AssertExpectationsForObjects(t, gqlClient, cliProvider)
	})

	t.Run("application not found", func(t *testing.T) {
		emptyResponse := graphql.ApplicationPageExt{}

		gqlQueryBuilder := director.NewClient(nil, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		expectedQuery := gqlQueryBuilder.GetApplicationsByNameRequest(appName)

		var apps appdetails.GqlSuccessfulAppPage
		gqlClient := &automock.GraphQLClient{}
		gqlClient.On("Run", mock.Anything, expectedQuery, &apps).Run(injectDirectorResponse(t, emptyResponse)).Return(nil)

		cliProvider := &automock.Provider{}
		cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)
		appMiddleware := appdetails.NewApplicationMiddleware(cliProvider)

		router := mux.NewRouter()
		router.Use(appMiddleware.Middleware)
		router.HandleFunc("/{app-name}", fixDummyHandler())
		rw := httptest.NewRecorder()

		//WHEN
		router.ServeHTTP(rw, req)

		//THEN
		assert.Equal(t, http.StatusNotFound, rw.Code)
		mock.AssertExpectationsForObjects(t, gqlClient, cliProvider)

		var response map[string]interface{}
		err := json.Unmarshal(rw.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`application with name %s not found`, appName), response["error"])
	})

	t.Run("found more than one application", func(t *testing.T) {
		appPage := graphql.ApplicationPageExt{ApplicationPage: graphql.ApplicationPage{},
			Data: []*graphql.ApplicationExt{&app, &app}}
		gqlQueryBuilder := director.NewClient(nil, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		expectedQuery := gqlQueryBuilder.GetApplicationsByNameRequest(appName)

		var apps appdetails.GqlSuccessfulAppPage
		gqlClient := &automock.GraphQLClient{}
		gqlClient.On("Run", mock.Anything, expectedQuery, &apps).Run(injectDirectorResponse(t, appPage)).Return(nil)

		cliProvider := &automock.Provider{}
		cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)
		appMiddleware := appdetails.NewApplicationMiddleware(cliProvider)

		router := mux.NewRouter()
		router.Use(appMiddleware.Middleware)
		router.HandleFunc("/{app-name}", fixDummyHandler())
		rw := httptest.NewRecorder()

		//WHEN
		router.ServeHTTP(rw, req)

		//THEN
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		mock.AssertExpectationsForObjects(t, gqlClient, cliProvider)

		var response map[string]interface{}
		err := json.Unmarshal(rw.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf(`found more than 1 application with name %s`, appName), response["error"])
	})

	t.Run("director returns error", func(t *testing.T) {
		logger, hook := test.NewNullLogger()

		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		request := req.WithContext(ctx)

		gqlQueryBuilder := director.NewClient(nil, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		expectedQuery := gqlQueryBuilder.GetApplicationsByNameRequest(appName)

		var apps appdetails.GqlSuccessfulAppPage
		gqlClient := &automock.GraphQLClient{}
		gqlClient.On("Run", mock.Anything, expectedQuery, &apps).Return(testErr)

		cliProvider := &automock.Provider{}
		cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)
		appMiddleware := appdetails.NewApplicationMiddleware(cliProvider)

		router := mux.NewRouter()
		router.Use(appMiddleware.Middleware)
		router.HandleFunc("/{app-name}", fixDummyHandler())
		rw := httptest.NewRecorder()

		//WHEN
		router.ServeHTTP(rw, request)

		//THEN
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		mock.AssertExpectationsForObjects(t, gqlClient, cliProvider)
		assertLastLogEntryError(t, "while getting service: test error", hook)
	})

	t.Run("director returns unauthorized error", func(t *testing.T) {
		logger, hook := test.NewNullLogger()

		ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
		request := req.WithContext(ctx)

		gqlQueryBuilder := director.NewClient(nil, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		expectedQuery := gqlQueryBuilder.GetApplicationsByNameRequest(appName)

		var apps appdetails.GqlSuccessfulAppPage
		gqlClient := &automock.GraphQLClient{}
		gqlClient.On("Run", mock.Anything, expectedQuery, &apps).Return(errors.New("insufficient scopes provided"))

		cliProvider := &automock.Provider{}
		cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)
		appMiddleware := appdetails.NewApplicationMiddleware(cliProvider)

		router := mux.NewRouter()
		router.Use(appMiddleware.Middleware)
		router.HandleFunc("/{app-name}", fixDummyHandler())
		rw := httptest.NewRecorder()

		//WHEN
		router.ServeHTTP(rw, request)

		//THEN
		assert.Equal(t, http.StatusUnauthorized, rw.Code)
		mock.AssertExpectationsForObjects(t, gqlClient, cliProvider)
		assertLastLogEntryError(t, "while getting service: insufficient scopes provided", hook)
	})
}

func assertAppInContext(t *testing.T, expectedApp graphql.ApplicationExt) func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		app, err := appdetails.LoadFromContext(r.Context())
		require.NoError(t, err)
		assert.Equal(t, expectedApp, app)
		writer.WriteHeader(http.StatusOK)
	}
}

func injectDirectorResponse(t *testing.T, result graphql.ApplicationPageExt) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*appdetails.GqlSuccessfulAppPage)
		if !ok {
			t.FailNow()
		}
		arg.Result = result
	}
}

func fixApplicationExt(name string) graphql.ApplicationExt {
	return graphql.ApplicationExt{
		Application: graphql.Application{
			BaseEntity: &graphql.BaseEntity{ID: uuid.New().String()},
			Name:       name,
		},
		Labels: map[string]interface{}{"name": name},
	}
}

func fixDummyHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
	}
}

func assertLastLogEntryError(t *testing.T, errMessage string, hook *test.Hook) {
	entry := hook.LastEntry()
	require.NotNil(t, entry)
	assert.Equal(t, logrus.ErrorLevel, entry.Level)
	assert.Equal(t, errMessage, entry.Message)
}
