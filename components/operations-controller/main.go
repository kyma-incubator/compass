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
	"flag"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
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

	cfg.GraphQLClient.GraphqlEndpoint = cfg.GraphQLClient.GraphqlEndpoint + "?useBundles=true" // TODO: Delete after bundles are adopted

	err = cfg.Validate()
	fatalOnError(err)

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(devLogging)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               port,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	tokenProviderFromHeader, err := oauth.NewTokenProviderFromHeader(cfg.GraphQLClient.GraphqlEndpoint)
	if err != nil {
		fatalOnError(err)
	}

	directorGraphQLClient, err := graphql.PrepareGqlClient(cfg.GraphQLClient, cfg.HttpClient, tokenProviderFromHeader)
	fatalOnError(err)

	if err = (&controllers.OperationReconciler{
		Client:         mgr.GetClient(),
		Log:            ctrl.Log.WithName("controllers").WithName("Operation"),
		Scheme:         mgr.GetScheme(),
		DirectorClient: director.NewClient("", directorGraphQLClient), // TODO: Add DirectorURL
	}).SetupWithManager(mgr); err != nil {
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
