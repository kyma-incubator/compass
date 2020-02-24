package main

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/gardener"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/http_client"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	gcli "github.com/machinebox/graphql"
	"github.com/pivotal-cf/brokerapi"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Config holds configuration for the whole application
type Config struct {
	Auth struct {
		Username string
		Password string
	}
	Host string `envconfig:"optional"`
	Port string `envconfig:"default=8080"`

	Provisioning broker.ProvisioningConfig
	Director     director.Config
	Database     storage.Config
	Gardener     gardener.Config

	ServiceManager internal.ServiceManagerOverride

	KymaVersion                          string
	ManagedRuntimeComponentsYAMLFilePath string

	Broker broker.Config
}

func main() {
	// create and fill config
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err)

	// create logger
	logger := lager.NewLogger("kyma-env-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Info("Starting Kyma Environment Broker")

	// create broker credentials
	brokerCredentials := brokerapi.BrokerCredentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	provisionerClient := provisioner.NewProvisionerClient(cfg.Provisioning.URL, true)

	// create kubernetes client
	k8sCfg, err := config.GetConfig()
	fatalOnError(err)
	cli, err := client.New(k8sCfg, client.Options{})
	fatalOnError(err)

	// create director client on the base of graphQL client and OAuth client
	httpClient := http_client.NewHTTPClient(30, cfg.Director.SkipCertVerification)
	graphQLClient := gcli.NewClient(cfg.Director.URL, gcli.WithHTTPClient(httpClient))
	// TODO: remove after debug mode
	graphQLClient.Log = func(s string) { log.Println(s) }
	oauthClient := oauth.NewOauthClient(httpClient, cli, cfg.Director.OauthCredentialsSecretName, cfg.Director.Namespace)
	fatalOnError(oauthClient.WaitForCredentials())
	directorClient := director.NewDirectorClient(oauthClient, graphQLClient)

	// create storage
	db, err := storage.New(cfg.Database.ConnectionURL())
	fatalOnError(err)

	// Register disabler. Convention:
	// {component-name} : {component-disabler-service}
	//
	// Using map is intentional - we ensure that component name is not duplicated.
	optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Loki":       runtime.NewLokiDisabler(),
		"Kiali":      runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger":     runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		"Monitoring": runtime.NewGenericComponentDisabler("monitoring", "kyma-system"),
	}

	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)

	runtimeProvider := runtime.NewComponentsListProvider(cfg.KymaVersion, cfg.ManagedRuntimeComponentsYAMLFilePath)
	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	fatalOnError(err)

	gardenerClusterConfig, err := newGardenerClusterConfig(cfg)
	fatalOnError(err)

	gardenerSecrets, err := newGardenerSecretsInterface(gardenerClusterConfig, cfg)
	fatalOnError(err)

	// TODO: check if is it possible to use compass account pool someday?
	var compassAccountPool hyperscaler.AccountPool = nil
	var gardenerAccountPool hyperscaler.AccountPool = nil

	if gardenerSecrets != nil {
		gardenerAccountPool = hyperscaler.NewAccountPool(gardenerSecrets)
	}

	accountProvider := hyperscaler.NewAccountProvider(compassAccountPool, gardenerAccountPool)

	inputFactory := broker.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, cfg.KymaVersion, cfg.ServiceManager, accountProvider)

	dumper, err := broker.NewDumper()
	fatalOnError(err)

	kymaEnvBroker := &broker.KymaEnvironmentBroker{
		broker.NewServices(cfg.Broker, optComponentsSvc, dumper),
		broker.NewProvision(cfg.Broker, db.Instances(), inputFactory, cfg.Provisioning, provisionerClient, dumper),
		broker.NewDeprovision(db.Instances(), provisionerClient, dumper),
		broker.NewUpdate(dumper),
		broker.NewGetInstance(db.Instances(), dumper),
		broker.NewLastOperation(db.Instances(), provisionerClient, directorClient, dumper),
		broker.NewBind(dumper),
		broker.NewUnbind(dumper),
		broker.NewGetBinding(dumper),
		broker.NewLastBindingOperation(dumper),
	}

	// create and run broker OSB API
	brokerAPI := brokerapi.New(kymaEnvBroker, logger, brokerCredentials)
	r := handlers.LoggingHandler(os.Stdout, brokerAPI)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

func newGardenerClusterConfig(cfg Config) (*restclient.Config, error) {

	rawKubeconfig, err := ioutil.ReadFile(cfg.Gardener.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gardener Kubeconfig from path %s: %s", cfg.Gardener.KubeconfigPath, err.Error())
	}

	gardenerClusterConfig, err := gardener.RESTConfig(rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	return gardenerClusterConfig, nil
}

func newGardenerSecretsInterface(gardenerClusterCfg *restclient.Config, cfg Config) (corev1.SecretInterface, error) {

	gardenerNamespace := fmt.Sprintf("garden-%s", cfg.Gardener.Project)

	gardenerClusterClient, err := kubernetes.NewForConfig(gardenerClusterCfg)
	if err != nil {
		return nil, err
	}

	return gardenerClusterClient.CoreV1().Secrets(gardenerNamespace), nil
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
