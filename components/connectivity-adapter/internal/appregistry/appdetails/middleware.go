package appdetails

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/gorilla/mux"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const appNamePathVariable = "app-name"

//go:generate mockery -name=GraphQLRequestBuilder -output=automock -outpkg=automock -case=underscore
type GraphQLRequestBuilder interface {
	GetApplicationsByName(appName string) *gcli.Request
}

type applicationMiddleware struct {
	cliProvider gqlcli.Provider
	logger      *log.Logger
	gqlProvider graphqlizer.GqlFieldsProvider
}

func NewApplicationMiddleware(cliProvider gqlcli.Provider, logger *log.Logger) *applicationMiddleware {
	return &applicationMiddleware{cliProvider: cliProvider, logger: logger}
}

func (mw *applicationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		variables := mux.Vars(r)
		appName := variables[appNamePathVariable]

		client := mw.cliProvider.GQLClient(r)
		directorCli := director.NewClient(client, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{})
		query := directorCli.GetApplicationsByNameRequest(appName)

		var apps GqlSuccessfulAppPage
		err := client.Run(r.Context(), query, &apps)
		if err != nil {
			wrappedErr := errors.Wrap(err, "while getting service")
			mw.logger.Error(wrappedErr)
			res.WriteError(w, wrappedErr, apperrors.CodeInternal)
			return
		}

		if len(apps.Result.Data) == 0 {
			message := fmt.Sprintf("application with name %s not found", appName)
			mw.logger.Warn(message)
			res.WriteErrorMessage(w, message, apperrors.CodeNotFound)
			return
		}

		if len(apps.Result.Data) != 1 {
			message := fmt.Sprintf("found more than 1 application with name %s", appName)
			mw.logger.Warn(message)
			res.WriteErrorMessage(w, message, apperrors.CodeInternal)
			return
		}

		app := apps.Result.Data[0]
		ctx := SaveToContext(r.Context(), *app)
		ctxWithCli := gqlcli.SaveToContext(ctx, client)
		requestWithCtx := r.WithContext(ctxWithCli)
		next.ServeHTTP(w, requestWithCtx)
	})
}

type GqlSuccessfulAppPage struct {
	Result graphql.ApplicationPageExt `json:"result"`
}
