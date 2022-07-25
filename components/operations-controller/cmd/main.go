/*
Copyright 2021.

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
	"os"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/utils"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/config"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	collector "github.com/kyma-incubator/compass/components/operations-controller/internal/metrics"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

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

	loggerConfig := zapcore.EncoderConfig{
		TimeKey:        "written_at",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	ctrl.SetLogger(zap.New(zap.UseDevMode(devLogging), zap.Encoder(zapcore.NewJSONEncoder(loggerConfig))))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     cfg.Server.MetricAddress,
		HealthProbeBindAddress: cfg.Server.HealthAddress,
		Port:                   port,
		LeaderElection:         cfg.Server.EnableLeaderElection,
		LeaderElectionID:       leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	httpClient, err := utils.PrepareHttpClient(cfg.HttpClient)
	fatalOnError(err)

	certCache, err := certloader.StartCertLoader(ctx, certloader.Config{
		ExternalClientCertSecret:  cfg.ExternalClient.CertSecret,
		ExternalClientCertCertKey: cfg.ExternalClient.CertKey,
		ExternalClientCertKeyKey:  cfg.ExternalClient.KeyKey,
	})
	fatalOnError(errors.Wrapf(err, "Failed to initialize certificate loader"))

	httpMTLSClient := utils.PrepareMTLSClient(cfg.HttpClient, certCache)

	directorClient, err := director.NewClient(cfg.Director.OperationEndpoint, cfg.GraphQLClient, httpClient)
	fatalOnError(err)

	collector := collector.NewCollector()
	metrics.Registry.MustRegister(collector)
	controller := controllers.NewOperationReconciler(cfg.Webhook,
		status.NewManager(mgr.GetClient()),
		k8s.NewClient(mgr.GetClient()),
		directorClient,
		webhookclient.NewClient(httpClient, httpMTLSClient),
		collector)

	if err = controller.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Operation")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

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
