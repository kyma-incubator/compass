package gardener

import (
	"context"
	"strings"
	"sync"
	"time"

	gclientset "github.com/gardener/gardener/pkg/client/core/clientset/versioned"
	ginformers "github.com/gardener/gardener/pkg/client/core/informers/externalversions"
	shootsinformer "github.com/gardener/gardener/pkg/client/core/informers/externalversions/core/v1beta1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	secretsinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	labelAccountID       = "account"
	labelSubAccountID    = "subaccount"
	labelHyperscalerType = "hyperscalerType"
	labelTenantName      = "tenantName"

	fieldSecretBindingName = "spec.secretBindingName"

	defaultResyncPeriod = time.Second * 30
)

// Account is a representation of the skr accounts
type Account struct {
	Name           string
	ProviderType   string
	AccountID      string
	SubAccountID   string
	TechnicalID    string
	TenantName     string
	CredentialName string
	CredentialData map[string][]byte
}

// Controller is the controller that watch for shoots and secrets
type Controller struct {
	providertype            string
	gclientset              gclientset.Interface
	kclientset              kubernetes.Interface
	gardenerInformerFactory ginformers.SharedInformerFactory
	kubeInformerFactory     kubeinformers.SharedInformerFactory
	shootInformer           shootsinformer.ShootInformer
	secretInformer          secretsinformer.SecretInformer
	accountsChan            chan<- *Account
	logger                  *zap.SugaredLogger
	shootQueue              map[string]bool
}

// NewController return a new controller for watching shoots and secrets
func NewController(client *Client, provider string, accountsChan chan<- *Account, logger *zap.SugaredLogger) (*Controller, error) {
	gardenerInformerFactory := ginformers.NewSharedInformerFactoryWithOptions(
		client.GardenerClientset,
		defaultResyncPeriod,
		ginformers.WithNamespace(client.Namespace),
	)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		client.KubernetesClientset,
		defaultResyncPeriod,
		kubeinformers.WithNamespace(client.Namespace),
		kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) { opts.LabelSelector = labelHyperscalerType }),
	)

	shootInformer := gardenerInformerFactory.Core().V1beta1().Shoots()
	secretInformer := kubeInformerFactory.Core().V1().Secrets()

	controller := &Controller{
		providertype:            strings.ToLower(provider),
		gclientset:              client.GardenerClientset,
		kclientset:              client.KubernetesClientset,
		gardenerInformerFactory: gardenerInformerFactory,
		kubeInformerFactory:     kubeInformerFactory,
		shootInformer:           shootInformer,
		secretInformer:          secretInformer,
		accountsChan:            accountsChan,
		logger:                  logger.With("component", "gardener"),
		shootQueue:              make(map[string]bool),
	}

	// Set up event handlers for Shoot resources
	shootInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.shootAddHandlerFunc,
		UpdateFunc: controller.shootUpdateFunc,
		DeleteFunc: controller.shootDeleteHandlerFunc,
	})

	// Set up event handler for Secret resources
	secretInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: controller.secretUpdateFunc,
	})

	return controller, nil
}

// Run will set up the event handlers for secrets and shoots, as well
// as syncing informer caches.
func (c *Controller) Run(parentCtx context.Context, parentwg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	parentwg.Add(1)
	defer parentwg.Done()

	c.logger.Info("starting controller")

	// Start the informer factories to begin populating the informer caches
	c.gardenerInformerFactory.Start(ctx.Done())
	c.kubeInformerFactory.Start(ctx.Done())

	c.logger.Debug("waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.shootInformer.Informer().HasSynced, c.secretInformer.Informer().HasSynced); !ok {
		c.logger.Errorf("failed to wait for caches to sync")
		return
	}

	c.logger.Debug("informer caches sync completed")

	<-ctx.Done()

	c.logger.Info("stopping controller")
}
