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
	director_graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	operationsv1alpha1 "github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/config"
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

	err = operationsv1alpha1.AddToScheme(scheme)
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

	httpClient := http.Client{
		Transport: httputil.NewHTTPTransport(cfg.HttpClient),
		Timeout:   cfg.HttpClient.Timeout,
	}

	unsignedTokenProvider, err := director.NewUnsignedTokenProvider(cfg.GraphQLClient.GraphqlEndpoint)
	if err != nil {
		fatalOnError(err)
	}

	directorGraphQLClient, err := graphql.PrepareGqlClient(cfg.GraphQLClient, cfg.HttpClient, unsignedTokenProvider)
	fatalOnError(err)

	controller := controllers.NewOperationReconciler(cfg.Webhook, ctrl.Log.WithName("controllers").WithName("Operation"),
		k8s.NewClient(mgr.GetClient()),
		director.NewClient(cfg.Director.InternalAddress, httpClient, directorGraphQLClient),
		webhook.NewClient(httpClient, defaultOAuthClientProviderFunc))

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

func defaultOAuthClientProviderFunc(ctx context.Context, client http.Client, oauthCreds *director_graphql.OAuthCredentialData) *http.Client {
	conf := &clientcredentials.Config{
		ClientID:     oauthCreds.ClientID,
		ClientSecret: oauthCreds.ClientSecret,
		TokenURL:     oauthCreds.URL,
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	securedClient := conf.Client(ctx)
	securedClient.Timeout = client.Timeout
	return securedClient
}
