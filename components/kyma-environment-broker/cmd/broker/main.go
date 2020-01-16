package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/pivotal-cf/brokerapi"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	// feature flag indicates whether use Provisioner API which returns RuntimeID
	ProcessRuntimeID bool `envconfig:"default=false"`
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err)

	logger := lager.NewLogger("kyma-env-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Info("Starting Kyma Environment Broker")

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	var provisionerClient provisioner.Client
	if cfg.ProcessRuntimeID {
		provisionerClient = provisioner.NewProvisionerClientV2(cfg.Provisioning.URL, true)
	} else {
		provisionerClient = provisioner.NewProvisionerClient(cfg.Provisioning.URL, true)
	}

	secrets, err := newSecretsInterface(cfg.Director.CredentialsNamespace)
	fatalOnError(err)
	oauthClient := oauth.NewOauthClient(newHTTPClient(false), secrets, cfg.Director.OauthCredentialsSecretName)
	tkn, err := oauthClient.GetAuthorizationToken()
	fatalOnError(err)
	fmt.Println("DUPA:", tkn.AccessToken)
	director.NewDirectorClient(oauthClient)

	db, err := storage.New(cfg.Database.ConnectionURL())
	fatalOnError(err)

	kymaEnvBroker, err := broker.NewBroker(provisionerClient, cfg.Provisioning, db.Instances())
	fatalOnError(err)

	brokerAPI := brokerapi.New(kymaEnvBroker, logger, brokerCredentials)
	r := handlers.LoggingHandler(os.Stdout, brokerAPI)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

func newHTTPClient(skipCertVeryfication bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVeryfication},
		},
		Timeout: 30 * time.Second,
	}
}

func newSecretsInterface(namespace string) (v1.SecretInterface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Warnf("Failed to read in cluster config: %s", err.Error())
		logrus.Info("Trying to initialize with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset.CoreV1().Secrets(namespace), nil
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
