/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

//
import (
	"context"
	"os"
	"sync"

	"github.com/kyma-incubator/compass/components/system-broker/internal/metrics"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/internal/config"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	sblog "github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/signal"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	environment, err := env.Default(ctx, config.AddPFlags)
	fatalOnError(err)

	cfg, err := config.New(environment)
	fatalOnError(err)

	err = cfg.Validate()
	fatalOnError(err)

	metrics.MapPathToInstrumentation(cfg.Server.RootAPI)

	ctx, err = log.Configure(ctx, cfg.Log)
	fatalOnError(err)

	directorGraphQLClient, err := graphql.PrepareGqlClient(cfg.GraphQLClient, cfg.HttpClient, oauth.NewTokenAuthorizationProviderFromHeader())
	fatalOnError(err)

	systemBroker := osb.NewSystemBroker(directorGraphQLClient, cfg.ORD.ServiceURL+cfg.ORD.StaticPath)
	collector := metrics.NewCollector()
	osbApi := osb.API(cfg.Server.RootAPI, systemBroker, sblog.NewDefaultLagerAdapter(), collector)
	prometheus.MustRegister(collector)

	middlewares := []mux.MiddlewareFunc{
		httputil.HeaderForwarder(cfg.HttpClient.ForwardHeaders),
	}
	srv := server.New(cfg.Server, middlewares, osbApi)

	metricsHandler := mux.NewRouter()
	metricsHandler.Handle("/metrics", promhttp.Handler())
	metricsServer := server.NewServerWithRouter(&server.Config{
		Port:            cfg.Metrics.Port,
		Timeout:         cfg.Metrics.Timeout,
		ShutdownTimeout: cfg.Metrics.ShutDownTimeout,
	}, metricsHandler)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go metricsServer.Start(ctx, wg)
	go srv.Start(ctx, wg)

	wg.Wait()
}

func fatalOnError(err error) {
	if err != nil {
		log.D().Fatal(err.Error())
	}
}
