package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/cis"
	"github.com/kyma-incubator/compass/components/director/internal/dircleaner"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Database          persistence.DatabaseConfig
	Log               log.Config
	KubeConfig        cis.KubeConfig
	KubeClientConfig  kube.Config
	ClientTimeout     time.Duration `envconfig:"default=60s"`
	SkipSSLValidation bool          `envconfig:"default=false"`
	IsExecutedOnProd  bool          `envconfig:"default=false,APP_PROD_EXECUTION"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(err, "Error while loading app config")

	ctx, err := log.Configure(context.Background(), &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	tenantFetcherSvc := createDirCleanerSvc(ctx, cfg, transact)
	err = tenantFetcherSvc.Clean(ctx)

	exitOnError(err, "Error while cleaning directories from the database")

	log.C(ctx).Info("Successfully cleaned directories")
}

func createDirCleanerSvc(ctx context.Context, cfg config, transact persistence.Transactioner) dircleaner.CleanerService {
	uidSvc := uid.NewService()
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

	labelConv := label.NewConverter()
	labelRepo := label.NewRepository(labelConv)
	labelDefConv := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConv)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)

	client := http.Client{
		Timeout: cfg.ClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		}}

	kubeClient, err := cis.NewKubernetesClient(ctx, cfg.KubeConfig, cfg.KubeClientConfig)
	exitOnError(err, "Failed to initialize Kubernetes client")

	cisService := cis.NewCisService(client, kubeClient, cfg.IsExecutedOnProd)
	return dircleaner.NewCleaner(transact, tenantSvc, labelSvc, cisService)
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
