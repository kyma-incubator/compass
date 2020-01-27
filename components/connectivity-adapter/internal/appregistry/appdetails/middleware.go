package appdetails

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"net/http"
)

const appNameVar = "app-name"
const nameKey = "name"

type applicationMiddleware struct {
	cliProvider gqlcli.Provider
}

func NewApplicationMiddleware(cliProvider gqlcli.Provider) *applicationMiddleware {
	return &applicationMiddleware{cliProvider: cliProvider}
}

func (mw *applicationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		variables := mux.Vars(r)
		gqlProvider := gql.GqlFieldsProvider{}
		appName := variables[appNameVar]
		query := fixAppByNameQuery(gqlProvider, appName)

		client := mw.cliProvider.GQLClient(r)
		var apps gqlSuccessfulAppPage
		err := client.Run(r.Context(), query, &apps)
		if err != nil {
			panic(err)
		}
		app := apps.Result.Data[0]
		if app == nil {
			panic("wyjebalo sie")
		}
		ctx := SaveToContext(r.Context(), *app)
		requestWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, requestWithCtx)
	})
}

func fixAppByNameQuery(gql gql.GqlFieldsProvider, appName string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(`query {
			result: applications(filter: {key:"%s", query: "\"%s\""}) {
					%s
			}
	}`, nameKey, appName, gql.Page(gql.ForApplication())))
}

type gqlSuccessfulAppPage struct {
	Result graphql.ApplicationPageExt `json:"result"`
}
