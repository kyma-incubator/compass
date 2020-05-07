package azure

import (
	"context"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/kyma-incubator/compass/components/metris/internal/gardener"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"
)

const (
	azureUserAgent = "metris"

	maxPollingDuration = 1 * time.Minute
)

var (
	ResourceSkusList ResourceSkus
)

// NewAzureProvider create new azure provider.
func NewAzureProvider(workers int, pollinterval time.Duration, accountsChannel <-chan *gardener.Account, eventsChannel chan<- *[]byte, logger *zap.SugaredLogger, tracelevel int) *Provider {
	ResourceSkusList = ResourceSkus{
		skus: make(map[string]*compute.ResourceSku),
	}

	return &Provider{
		mu:               sync.RWMutex{},
		workers:          workers,
		pollinterval:     pollinterval,
		accountsChannel:  accountsChannel,
		eventsChannel:    eventsChannel,
		queue:            workqueue.NewNamedDelayingQueue("azure-accounts"),
		clients:          make(map[string]*Client),
		logger:           logger,
		clientTraceLevel: tracelevel,
	}
}

// Collect start azure metrics gathering
func (a *Provider) Collect(ctx context.Context, parentwg *sync.WaitGroup) {
	var wg sync.WaitGroup

	parentwg.Add(1)
	defer parentwg.Done()

	// poll storage for azure accounts every minute
	go a.accountHandler(ctx.Done())

	a.logger.Info("starting provider")

	wg.Add(a.workers)

	for i := 0; i < a.workers; i++ {
		go func(i int) {
			defer wg.Done()

			for {
				clientname, quit := a.queue.Get()
				workerlogger := a.logger.With("worker", i, "account", clientname)

				if quit {
					workerlogger.Debug("shutting down")
					return
				}

				a.gatherMetrics(ctx, workerlogger, clientname.(string))

				a.queue.Done(clientname)

				// requeue item after X duration if client still in storage
				if !a.queue.ShuttingDown() {
					if _, ok := a.clients[clientname.(string)]; ok {
						workerlogger.Debugf("requeuing account in %s", a.pollinterval)
						a.queue.AddAfter(clientname, a.pollinterval)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	a.logger.Info("stopping provider")
}

func (a *Provider) accountHandler(stopCh <-chan struct{}) {
	for {
		select {
		case account := <-a.accountsChannel:
			subaccountid := account.Name

			a.mu.RLock()

			if _, ok := a.clients[account.Name]; ok {
				subaccountid = a.clients[account.Name].Account.SubAccountID
			}

			a.mu.RUnlock()

			logger := a.logger.With("account", subaccountid)
			logger.Debugf("received account")

			// account is getting deleted, remove from cache
			if len(account.CredentialData) == 0 {
				logger.Warn("account has been mark for deletion")
				delete(a.clients, account.Name)

				break
			}

			// have to decode secrets before mapping
			decodedSecrets := make(map[string]string)
			for k, v := range account.CredentialData {
				decodedSecrets[k] = string(v)
			}

			var (
				conf SecretMap
				err  error
			)

			err = mapstructure.Decode(decodedSecrets, &conf)
			if err != nil {
				a.logger.Error(err)
				break
			}

			clientID := conf.ClientID
			clientSecret := conf.ClientSecret
			tenantID := conf.TenantID
			subscriptionID := conf.SubscriptionID

			ccc := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)

			if conf.EnvironmentName != "" {
				if env, enverr := azure.EnvironmentFromName(conf.EnvironmentName); enverr == nil {
					ccc.Resource = env.ResourceManagerEndpoint
					ccc.AADEndpoint = env.ActiveDirectoryEndpoint
				}
			}

			authz, err := ccc.Authorizer()
			if err != nil {
				a.logger.Error(err)
				break
			}

			computeBaseClient := compute.NewWithBaseURI(ccc.Resource, subscriptionID)
			computeBaseClient.Authorizer = authz
			computeBaseClient.PollingDuration = maxPollingDuration
			computeBaseClient.SkipResourceProviderRegistration = true

			if err = computeBaseClient.AddToUserAgent(azureUserAgent); err != nil {
				a.logger.Error(err)
				break
			}

			networkBaseClient := network.NewWithBaseURI(ccc.Resource, subscriptionID)
			networkBaseClient.Authorizer = authz
			networkBaseClient.PollingDuration = maxPollingDuration
			networkBaseClient.SkipResourceProviderRegistration = true

			if err = networkBaseClient.AddToUserAgent(azureUserAgent); err != nil {
				a.logger.Error(err)
				break
			}

			insightsBaseClient := insights.NewWithBaseURI(ccc.Resource, subscriptionID)
			insightsBaseClient.Authorizer = authz
			insightsBaseClient.PollingDuration = maxPollingDuration
			insightsBaseClient.SkipResourceProviderRegistration = true

			if err = insightsBaseClient.AddToUserAgent(azureUserAgent); err != nil {
				a.logger.Error(err)
				break
			}

			resourcesBaseClient := resources.NewWithBaseURI(ccc.Resource, subscriptionID)
			resourcesBaseClient.Authorizer = authz
			resourcesBaseClient.PollingDuration = maxPollingDuration
			resourcesBaseClient.SkipResourceProviderRegistration = true

			if err = resourcesBaseClient.AddToUserAgent(azureUserAgent); err != nil {
				a.logger.Error(err)
				break
			}

			eventhubBaseClient := eventhub.NewWithBaseURI(ccc.Resource, subscriptionID)
			eventhubBaseClient.Authorizer = authz
			eventhubBaseClient.PollingDuration = maxPollingDuration
			eventhubBaseClient.SkipResourceProviderRegistration = true

			if err = eventhubBaseClient.AddToUserAgent(azureUserAgent); err != nil {
				a.logger.Error(err)
				break
			}

			if a.clientTraceLevel > 0 {
				dumpbody := false

				if a.clientTraceLevel > 1 {
					dumpbody = true
				}

				computeBaseClient.RequestInspector = LogRequest(logger, dumpbody)
				computeBaseClient.ResponseInspector = LogResponse(logger, dumpbody)
				networkBaseClient.RequestInspector = LogRequest(logger, dumpbody)
				networkBaseClient.ResponseInspector = LogResponse(logger, dumpbody)
				insightsBaseClient.RequestInspector = LogRequest(logger, dumpbody)
				insightsBaseClient.ResponseInspector = LogResponse(logger, dumpbody)
				resourcesBaseClient.RequestInspector = LogRequest(logger, dumpbody)
				resourcesBaseClient.ResponseInspector = LogResponse(logger, dumpbody)
				eventhubBaseClient.RequestInspector = LogRequest(logger, dumpbody)
				eventhubBaseClient.ResponseInspector = LogResponse(logger, dumpbody)
			}

			a.mu.Lock()
			a.clients[account.Name] = &Client{
				Account:             account,
				SubscriptionID:      subscriptionID,
				logger:              logger,
				computeBaseClient:   &computeBaseClient,
				networkBaseClient:   &networkBaseClient,
				insightsBaseClient:  &insightsBaseClient,
				resourcesBaseClient: &resourcesBaseClient,
				eventhubBaseClient:  &eventhubBaseClient,
			}

			resourceGroup, err := a.clients[account.Name].getResourceGroup(context.Background(), account.TechnicalID)
			if err == nil {
				a.clients[account.Name].Location = *resourceGroup.Location
			}

			a.mu.Unlock()

			a.queue.Add(account.Name)

		case <-stopCh:
			a.logger.Debug("stopping account handler")
			a.queue.ShutDown()

			return
		}
	}
}
