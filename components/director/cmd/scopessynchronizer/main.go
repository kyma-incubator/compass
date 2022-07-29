package main

import (
	"context"
	"net/http"
	"net/url"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	scopes "github.com/kyma-incubator/compass/components/director/internal/scopes_sync"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/ory/hydra-client-go/client"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Database          persistence.DatabaseConfig
	ConfigurationFile string
	OAuth20           oauth20.Config
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uidSvc := uid.NewService()
	correlationID := uidSvc.Generate()
	ctx = withCorrelationID(ctx, correlationID)

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(ctx, err, "Error while loading app config")

	oAuth20HTTPClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)))),
		Timeout:   cfg.OAuth20.HTTPClientTimeout,
	}
	adminURL, err := url.Parse(cfg.OAuth20.URL)
	exitOnError(ctx, err, "Error while parsing OAuth client endpoint")

	transport := httptransport.NewWithClient(adminURL.Host, adminURL.Path, []string{adminURL.Scheme}, oAuth20HTTPClient)
	hydra := client.New(transport, nil)

	cfgProvider := configProvider(ctx, cfg)
	oAuth20Svc := oauth20.NewService(cfgProvider, uidSvc, cfg.OAuth20.PublicAccessTokenEndpoint, hydra.Admin)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(ctx, err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(ctx, err, "Error while closing the connection to the database")
	}()

	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	syncService := scopes.NewService(oAuth20Svc, transact, systemauth.NewRepository(systemAuthConverter))
	err = syncService.SynchronizeClientScopes(ctx)
	exitOnError(ctx, err, "Error while updating client scopes")
}

func exitOnError(ctx context.Context, err error, context string) {
	if err != nil {
		log.C(ctx).WithError(err).Errorf("%s: %v", context, err)
		os.Exit(1)
	}
}

func withCorrelationID(ctx context.Context, id string) context.Context {
	correlationIDKey := correlation.RequestIDHeaderKey
	return correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &id)
}

func configProvider(ctx context.Context, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	exitOnError(ctx, provider.Load(), "Error on loading configuration file")

	return provider
}
