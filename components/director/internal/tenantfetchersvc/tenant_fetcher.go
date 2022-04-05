package tenantfetchersvc

import (
	"context"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type config struct {
	Database                    persistence.DatabaseConfig
	OAuthConfig                 tenantfetcher.OAuth2Config
	APIConfig                   tenantfetcher.APIConfig
	AuthMode                    oauth.AuthMode `envconfig:"APP_OAUTH_AUTH_MODE,default=standard"`
	ClientTimeout               time.Duration  `envconfig:"default=60s"`
	DirectorGraphQLEndpoint     string         `envconfig:"APP_DIRECTOR_GRAPHQL_ENDPOINT"`
	HTTPClientSkipSslValidation bool           `envconfig:"default=false"`
	QueryConfig                 tenantfetcher.QueryConfig
	TenantFieldMapping          tenantfetcher.TenantFieldMapping
	MovedSubaccountFieldMapping tenantfetcher.MovedSubaccountsFieldMapping
	TenantProvider              string   `envconfig:"APP_TENANT_PROVIDER"`
	SubaccountRegions           []string `envconfig:"default=central,APP_SUBACCOUNT_REGIONS"`
	TenantInsertChunkSize       int      `envconfig:"default=500,APP_TENANT_INSERT_CHUNK_SIZE"`
}

type fetcher struct {
}

// NewTenantFetcher creates new fetcher
func NewTenantFetcher() *fetcher {
	return &fetcher{}
}

func (f *fetcher) FetchTenantOnDemand(ctx context.Context, tenantID string) error {
	cfg := config{}
	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	tenantFetcherOnDemandSvc, err := createTenantFetcherOnDemandSvc(cfg, transact)
	exitOnError(err, "failed to create tenant fetcher on-demand service")

	return tenantFetcherOnDemandSvc.SyncTenant(tenantID)
}

func createTenantFetcherOnDemandSvc(cfg config, transact persistence.Transactioner) (*tenantfetcher.SubaccountOnDemandService, error) {
	eventAPIClient, err := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.AuthMode, cfg.APIConfig, cfg.ClientTimeout)
	if nil != err {
		return nil, err
	}

	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	labelDefRepository := labeldef.NewRepository(labelDefConverter)

	labelConverter := label.NewConverter()
	labelRepository := label.NewRepository(labelConverter)
	labelService := label.NewLabelService(labelRepository, labelDefRepository, uidSvc)

	tenantStorageConv := tenant.NewConverter()
	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepository, labelService)

	gqlClient := newInternalGraphQLClient(cfg.DirectorGraphQLEndpoint, cfg.ClientTimeout, cfg.HTTPClientSkipSslValidation)
	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}
	directorClient := graphqlclient.NewDirector(gqlClient)

	return tenantfetcher.NewSubaccountOnDemandService(cfg.SubaccountRegions, cfg.QueryConfig, cfg.TenantFieldMapping, eventAPIClient, transact, tenantStorageSvc, directorClient, cfg.TenantProvider, cfg.TenantInsertChunkSize, tenantStorageConv), nil
}

func newInternalGraphQLClient(url string, timeout time.Duration, skipSSLValidation bool) *gcli.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(tr, "Authorization")),
		Timeout:   timeout,
	}

	return gcli.NewClient(url, gcli.WithHTTPClient(client))
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
