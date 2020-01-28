package appdetails

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const appNameVar = "app-name"
const nameKey = "name"

//go:generate mockery -name=GraphQLRequestBuilder -output=automock -outpkg=automock -case=underscore
type GraphQLRequestBuilder interface {
	GetApplicationsByName(appName string) *gcli.Request
}

type applicationMiddleware struct {
	cliProvider     gqlcli.Provider
	logger          *log.Logger
	gqlProvider     gql.GqlFieldsProvider
	gqlQueryBuilder GraphQLRequestBuilder
}

func NewApplicationMiddleware(cliProvider gqlcli.Provider, logger *log.Logger, gqlQueryBuilder GraphQLRequestBuilder) *applicationMiddleware {
	return &applicationMiddleware{cliProvider: cliProvider, logger: logger, gqlQueryBuilder: gqlQueryBuilder}
}

func (mw *applicationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		variables := mux.Vars(r)
		appName := variables[appNameVar]
		query := mw.gqlQueryBuilder.GetApplicationsByName(appName)

		client := mw.cliProvider.GQLClient(r)
		var apps GqlSuccessfulAppPage
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
		ctx := SaveToContext(r.Context(), *app)
		requestWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, requestWithCtx)
	})
}

type GqlSuccessfulAppPage struct {
	Result graphql.ApplicationPageExt `json:"result"`
}
