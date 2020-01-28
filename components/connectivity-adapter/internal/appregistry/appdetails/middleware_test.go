package appdetails_test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddleware(t *testing.T) {
	//GIVEN
	logger, _ := test.NewNullLogger()
	appName := "app-name"
	req := httptest.NewRequest("GET", fmt.Sprintf("/%s", appName), strings.NewReader(""))
	app := fixApplicationExt(appName)
	cli := automock.GraphQLClient{}
	expectedQuery := appdetails.FixAppByNameQuery(gql.GqlFieldsProvider{}, appName)
	var apps appdetails.GqlSuccessfulAppPage
	cli.On("Run", mock.Anything, expectedQuery, &apps).Run(injectDirectorResponse(t, app)).Return(nil)
	cliProvider := fixProviderMock(req, &cli)

	router := mux.NewRouter()
	router.HandleFunc("/{app-name}", assertContextHandler(t, app))

	appMidlleware := appdetails.NewApplicationMiddleware(cliProvider, logger, gql.GqlFieldsProvider{})
	router.Use(appMidlleware.Middleware)
	rw := httptest.NewRecorder()

	//WHEN
	router.ServeHTTP(rw, req)

	//THEN
	assert.Equal(t, http.StatusOK, rw.Code)
}

func assertContextHandler(t *testing.T, expectedApp graphql.ApplicationExt) func(w http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		app, err := appdetails.LoadFromContext(r.Context())
		require.NoError(t, err)
		assert.Equal(t, expectedApp, app)
		writer.WriteHeader(http.StatusOK)
	}
}
func fixProviderMock(httpReq *http.Request, gqlClient *automock.GraphQLClient) *automock.Provider {
	cliProvider := &automock.Provider{}
	cliProvider.On("GQLClient", mock.AnythingOfType("*http.Request")).Return(gqlClient)
	return cliProvider
}

func injectDirectorResponse(t *testing.T, app graphql.ApplicationExt) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*appdetails.GqlSuccessfulAppPage)
		if !ok {
			t.FailNow()
		}
		arg.Result = graphql.ApplicationPageExt{
			ApplicationPage: graphql.ApplicationPage{},
			Data:            []*graphql.ApplicationExt{&app},
		}
	}
}

func fixApplicationExt(name string) graphql.ApplicationExt {
	return graphql.ApplicationExt{
		Application: graphql.Application{
			ID:   uuid.New().String(),
			Name: name,
		},
		Labels: map[string]interface{}{"name": name},
	}
}
