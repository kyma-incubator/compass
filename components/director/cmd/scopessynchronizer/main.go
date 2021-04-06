package main

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/kyma-incubator/compass/components/director/internal/config/scopesynchronizer"
	"github.com/kyma-incubator/compass/components/director/pkg/env"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/scopes_sync"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/ory/hydra-client-go/client"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	environment, err := env.Default(ctx, scopesynchronizer.AddPFlags)
	exitOnError(ctx, err, "Error while creating environment")

	cfg, err := scopesynchronizer.New(environment)
	exitOnError(ctx, err, "Error while creating config")

	err = cfg.Validate()
	exitOnError(ctx, err, "Error while validating config")

	uidSvc := uid.NewService()
	correlationID := uidSvc.Generate()
	ctx = withCorrelationID(ctx, correlationID)

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	oAuth20HTTPClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
		Timeout:   cfg.OAuth20.HTTPClientTimeout,
	}
	adminURL, err := url.Parse(cfg.OAuth20.URL)
	exitOnError(ctx, err, "Error while parsing OAuth client endpoint")

	transport := httptransport.NewWithClient(adminURL.Host, adminURL.Path, []string{adminURL.Scheme}, oAuth20HTTPClient)
	hydra := client.New(transport, nil)

	cfgProvider := configProvider(ctx, *cfg)
	oAuth20Svc := oauth20.NewService(cfgProvider, uidSvc, cfg.OAuth20.PublicAccessTokenEndpoint, hydra.Admin)

	transact, closeFunc, err := persistence.Configure(ctx, *cfg.DB)
	exitOnError(ctx, err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(ctx, err, "Error while closing the connection to the database")
	}()

	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	syncService := scopes_sync.NewService(oAuth20Svc, transact, systemauth.NewRepository(systemAuthConverter))
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

func configProvider(ctx context.Context, cfg scopesynchronizer.Config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	exitOnError(ctx, provider.Load(), "Error on loading configuration file")

	return provider
}
