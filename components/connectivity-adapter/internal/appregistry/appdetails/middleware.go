package appdetails

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/retry"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const appNamePathVariable = "app-name"

//go:generate mockery --name=GraphQLRequestBuilder --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLRequestBuilder interface {
	GetApplicationsByName(appName string) *gcli.Request
}

type applicationMiddleware struct {
	cliProvider gqlcli.Provider
	gqlProvider graphqlizer.GqlFieldsProvider
}

func NewApplicationMiddleware(cliProvider gqlcli.Provider) *applicationMiddleware {
	return &applicationMiddleware{cliProvider: cliProvider}
}

func (mw *applicationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		variables := mux.Vars(r)
		appName := variables[appNamePathVariable]

		ctx := r.Context()
		logger := log.C(ctx)

		logger.Infof("resolving application with name '%s'...", appName)

		client := mw.cliProvider.GQLClient(r)
		directorCli := director.NewClient(client, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		query := directorCli.GetApplicationsByNameRequest(appName)

		var apps GqlSuccessfulAppPage
		err := retry.GQLRun(client.Run, r.Context(), query, &apps)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while getting service")
			logger.Error(wrappedErr)
			res.WriteError(w, wrappedErr, apperrors.CodeInternal)
			return
		}

		if len(apps.Result.Data) == 0 {
			message := fmt.Sprintf("application with name %s not found", appName)
			logger.Warn(message)
			res.WriteErrorMessage(w, message, apperrors.CodeNotFound)
			return
		}

		if len(apps.Result.Data) != 1 {
			message := fmt.Sprintf("found more than 1 application with name %s", appName)
			logger.Warn(message)
			res.WriteErrorMessage(w, message, apperrors.CodeInternal)
			return
		}

		app := apps.Result.Data[0]

		logger.Infof("app '%s' details fetched successfully", appName)

		ctx = SaveToContext(ctx, *app)
		ctxWithCli := gqlcli.SaveToContext(ctx, client)
		requestWithCtx := r.WithContext(ctxWithCli)
		next.ServeHTTP(w, requestWithCtx)
	})
}

type GqlSuccessfulAppPage struct {
	Result graphql.ApplicationPageExt `json:"result"`
}
