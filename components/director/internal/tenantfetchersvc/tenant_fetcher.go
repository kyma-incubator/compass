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
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type fetcher struct {
	eventsCfg  EventsConfig
	handlerCfg HandlerConfig
}

// NewTenantFetcher creates new fetcher
func NewTenantFetcher(eventsCfg EventsConfig, handlerCfg HandlerConfig) *fetcher {
	return &fetcher{
		eventsCfg:  eventsCfg,
		handlerCfg: handlerCfg,
	}
}

func (f *fetcher) FetchTenantOnDemand(ctx context.Context, tenantID string) (error, *tenantfetcher.ClientError) {
	transact, closeFunc, err := persistence.Configure(ctx, f.handlerCfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	tenantFetcherOnDemandSvc, err := f.createTenantFetcherOnDemandSvc(transact)
	exitOnError(err, "failed to create tenant fetcher on-demand service")

	return tenantFetcherOnDemandSvc.SyncTenant(ctx, tenantID)
}

func (f *fetcher) createTenantFetcherOnDemandSvc(transact persistence.Transactioner) (*tenantfetcher.SubaccountOnDemandService, error) {
	eventAPIClient, err := tenantfetcher.NewClient(f.eventsCfg.OAuthConfig, f.eventsCfg.AuthMode, f.eventsCfg.APIConfig, f.handlerCfg.ClientTimeout)
	if nil != err {
		return nil, err
	}

	tenantStorageConv := tenant.NewConverter()
	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	labelDefRepository := labeldef.NewRepository(labelDefConverter)

	labelConverter := label.NewConverter()
	labelRepository := label.NewRepository(labelConverter)
	labelService := label.NewLabelService(labelRepository, labelDefRepository, uidSvc)

	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepository, labelService)

	gqlClient := newInternalGraphQLClient(f.handlerCfg.DirectorGraphQLEndpoint, f.handlerCfg.ClientTimeout, f.handlerCfg.HTTPClientSkipSslValidation)
	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}
	directorClient := graphqlclient.NewDirector(gqlClient)

	return tenantfetcher.NewSubaccountOnDemandService(f.eventsCfg.QueryConfig, f.eventsCfg.TenantFieldMapping, eventAPIClient, transact, tenantStorageSvc, directorClient, f.handlerCfg.TenantProvider, tenantStorageConv), nil
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
