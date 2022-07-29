package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/connector/config"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app Config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Filed to configure logger")

	log.C(ctx).Info("Starting Connector Service")
	log.C(ctx).Infof("Config: %s", cfg.String())

	k8sClientSet, appErr := newK8SClientSet(ctx, cfg.KubernetesClient.PollInteval, cfg.KubernetesClient.PollTimeout, cfg.KubernetesClient.Timeout)
	exitOnError(appErr, "Failed to initialize Kubernetes client.")

	directorGCLI := newInternalGraphQLClient(cfg.OneTimeTokenURL, cfg.HTTPClientTimeout, cfg.HttpClientSkipSslValidation)
	internalComponents, certsLoader := config.InitInternalComponents(cfg, k8sClientSet, directorGCLI)

	go certsLoader.Run(ctx)

	certificateResolver := api.NewCertificateResolver(
		internalComponents.Authenticator,
		internalComponents.TokenService,
		internalComponents.CertificateService,
		internalComponents.CSRSubjectConsts,
		cfg.DirectorURL,
		cfg.CertificateSecuredConnectorURL,
		internalComponents.RevokedCertsRepository)

	authContextMiddleware := authentication.NewAuthenticationContextMiddleware()

	externalGqlServer, err := config.PrepareExternalGraphQLServer(cfg, certificateResolver, correlation.AttachCorrelationIDToContext(), log.RequestLogger(), authContextMiddleware.PropagateAuthentication)
	exitOnError(err, "Failed configuring external graphQL handler")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go startServer(ctx, externalGqlServer, wg)

	wg.Wait()
}

func startServer(parentCtx context.Context, server *http.Server, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	go func() {
		defer wg.Done()
		<-ctx.Done()
		stopServer(server)
	}()

	log.C(ctx).Infof("Starting and listening on %s://%s", "http", server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.C(ctx).Fatalf("Could not listen on %s://%s: %v\n", "http", server.Addr, err)
	}
}

func stopServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()

		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.C(ctx).Panic("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	server.SetKeepAlivesEnabled(false)

	if err := server.Shutdown(ctx); err != nil {
		log.C(ctx).Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func newK8SClientSet(ctx context.Context, interval, pollingTimeout, timeout time.Duration) (*kubernetes.Clientset, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.C(ctx).WithError(err).Warn("Failed to read in cluster Config")
		log.C(ctx).Info("Trying to initialize with local Config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "Config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	k8sConfig.Timeout = timeout

	k8sClientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	err = wait.PollImmediate(interval, pollingTimeout, func() (bool, error) {
		select {
		case <-ctx.Done():
			return true, nil
		default:
		}
		_, err := k8sClientSet.ServerVersion()
		if err != nil {
			log.C(ctx).Debugf("Failed to access API Server: %s", err.Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	log.C(ctx).Info("Successfully initialized kubernetes client")
	return k8sClientSet, nil
}

func newInternalGraphQLClient(URL string, timeout time.Duration, skipSSLValidation bool) *gcli.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   timeout,
	}

	return gcli.NewClient(URL, gcli.WithHTTPClient(client))
}
