package appdetails

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const appNameVar = "app-name"
const nameKey = "name"

type applicationMiddleware struct {
	cliProvider gqlcli.Provider
	logger      *log.Logger
	gqlProvider gql.GqlFieldsProvider
}

func NewApplicationMiddleware(cliProvider gqlcli.Provider, logger *log.Logger, fieldProvider gql.GqlFieldsProvider) *applicationMiddleware {
	return &applicationMiddleware{cliProvider: cliProvider, logger: logger, gqlProvider: fieldProvider}
}

func (mw *applicationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		variables := mux.Vars(r)
		appName := variables[appNameVar]
		query := fixAppByNameQuery(mw.gqlProvider, appName)

		client := mw.cliProvider.GQLClient(r)
		var apps gqlSuccessfulAppPage
		err := client.Run(r.Context(), query, &apps)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while getting service")
			mw.logger.Error(wrappedErr)
			reqerror.WriteError(w, wrappedErr, apperrors.CodeInternal)
			return
		}

		if len(apps.Result.Data) == 0 {
			message := fmt.Sprintf("service with name %s not found", appName)
			reqerror.WriteErrorMessage(w, message, apperrors.CodeNotFound)
			return
		}

		if len(apps.Result.Data) != 1 {
			message := fmt.Sprintf("found more than 1 service with name %s", appName)
			reqerror.WriteErrorMessage(w, message, apperrors.CodeInternal)
			return
		}

		app := apps.Result.Data[0]
		if app == nil {
			message := fmt.Sprintf("service with name %s not found", appName)
			reqerror.WriteErrorMessage(w, message, apperrors.CodeNotFound)
			return
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
