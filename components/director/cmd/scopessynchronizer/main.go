package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
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
		Timeout:   cfg.OAuth20.HTTPClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}
	cfgProvider := configProvider(ctx, cfg)

	oAuth20Svc := oauth20.NewService(cfgProvider, uidSvc, cfg.OAuth20, oAuth20HTTPClient)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(ctx, err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(ctx, err, "Error while closing the connection to the database")
	}()

	clientsFromHydra, err := oAuth20Svc.ListClients(ctx)
	if err != nil {
		exitOnError(ctx, err, "Error while listing clients from hydra")
	}
	clientScopes := convertScopesToMap(clientsFromHydra)

	auths := getAuthsForUpdate(ctx, transact)
	for _, auth := range auths {
		clientID := auth.Value.Credential.Oauth.ClientID

		objType, err := auth.GetReferenceObjectType()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
			continue
		}

		expectedScopes, err := oAuth20Svc.GetClientCredentialScopes(objType)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error while getting client credentials scopes for client with ID %s: %v", clientID, err)
			continue
		}

		scopesFromHydra := clientScopes[clientID]
		if str.Matches(scopesFromHydra, expectedScopes) {
			err = oAuth20Svc.UpdateClientCredentials(ctx, clientID, objType)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
			}
		}
	}

	log.C(ctx).Info("Finished synchronization of Hydra scopes")
}

func convertScopesToMap(clientsFromHydra []oauth20.Client) map[string][]string {
	clientScopes := make(map[string][]string)
	for _, s := range clientsFromHydra {
		clientScopes[s.ClientID] = strings.Split(s.Scopes, " ")
	}
	return clientScopes
}
func exitOnError(ctx context.Context, err error, context string) {
	if err != nil {
		log.C(ctx).WithError(err).Error(context)
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

func getAuthsForUpdate(ctx context.Context, transact persistence.Transactioner) []model.SystemAuth {
	tx, err := transact.Begin()
	if err != nil {
		exitOnError(ctx, err, "An error occurred while opening database transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)

	auths := listOauthAuths(ctx, tx)

	err = tx.Commit()
	exitOnError(ctx, err, fmt.Sprintf("An error occurred while closing database transaction: %v", err))

	return auths
}

func listOauthAuths(ctx context.Context, persist persistence.PersistenceOp) []model.SystemAuth {
	var dest systemauth.Collection

	query := "select * from system_auths where (value -> 'Credential' -> 'Oauth') is not null"
	log.D().Debugf("Executing DB query: %s", query)
	exitOnError(ctx, persist.Select(&dest, query), "Error while getting Oauth system auths")

	auths, err := multipleFromEntities(dest)
	exitOnError(ctx, err, "Error while converting entities")
	return auths
}

func multipleFromEntities(entities systemauth.Collection) ([]model.SystemAuth, error) {
	conv := systemauth.NewConverter(nil)
	var items []model.SystemAuth

	for _, ent := range entities {
		m, err := conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating system auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}
