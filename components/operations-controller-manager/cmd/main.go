/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/auth"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/config"
	collector "github.com/kyma-incubator/compass/components/operations-controller/internal/metrics"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	// +kubebuilder:scaffold:imports
)

var (
	devLogging       = true
	scheme           = runtime.NewScheme()
	port             = 9443
	leaderElectionID = "c8593142.compass"
	setupLog         = ctrl.Log.WithName("setup")
)

func init() {
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	// +kubebuilder:scaffold:scheme
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	env, err := env.Default(ctx, config.AddPFlags)
	fatalOnError(err)

	cfg, err := config.New(env)
	fatalOnError(err)

	err = cfg.Validate()
	fatalOnError(err)

	ctrl.SetLogger(zap.New(zap.UseDevMode(devLogging)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: cfg.Server.MetricAddress,
		Port:               port,
		LeaderElection:     cfg.Server.EnableLeaderElection,
		LeaderElectionID:   leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	httpClient, err := prepareHttpClient(cfg.HttpClient)
	fatalOnError(err)

	directorClient, err := director.NewClient(cfg.Director.OperationEndpoint, cfg.GraphQLClient, httpClient)
	fatalOnError(err)

	collector := collector.NewCollector()
	metrics.Registry.MustRegister(collector)
	controller := controllers.NewOperationReconciler(cfg.Webhook,
		status.NewManager(mgr.GetClient()),
		k8s.NewClient(mgr.GetClient()),
		directorClient,
		webhook.NewClient(httpClient),
		collector)

	if err = controller.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Operation")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func fatalOnError(err error) {
	if err != nil {
		log.D().Fatal(err.Error())
	}
}

func prepareHttpClient(cfg *httputil.Config) (*http.Client, error) {
	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransport(cfg))

	unsecuredClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.Timeout,
	}

	basicProvider := auth.NewBasicAuthorizationProvider()
	tokenProvider := auth.NewTokenAuthorizationProvider(unsecuredClient)
	unsignedTokenProvider := auth.NewUnsignedTokenAuthorizationProvider()

	securedTransport := httputil.NewSecuredTransport(httpTransport, basicProvider, tokenProvider, unsignedTokenProvider)
	securedClient := &http.Client{
		Transport: securedTransport,
		Timeout:   cfg.Timeout,
	}

	return securedClient, nil
}
