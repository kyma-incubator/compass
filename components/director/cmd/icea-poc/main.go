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

import (
	"context"
	"crypto"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/director/internal/icea-poc/directorclient"
	"github.com/kyma-incubator/compass/components/director/internal/icea-poc/notifications"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"net/http"
	"os"
	"time"
)

const envPrefix = "APP"

type config struct {
	Database persistence.DatabaseConfig
	Log      log.Config
	Director struct {
		Certificate                    string
		PrivateKey                     string
		DirectorExternalCertSecuredURL string
		InternalURL                    string        `envconfig:"default=http://127.0.0.1:3000/graphql"`
		ClientTimeout                  time.Duration `envconfig:"default=115s"`
		SkipSSLValidation              bool          `envconfig:"default=false"`
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "")

	parsedCert, err := cert.ParseCertificate(cfg.Director.Certificate, cfg.Director.PrivateKey)
	exitOnError(err, "Failed to parse Certificate")

	certSecuredGraphQLClient := NewCertAuthorizedGraphQLClientWithCustomURL(cfg.Director.DirectorExternalCertSecuredURL, parsedCert.PrivateKey, parsedCert.Certificate, cfg.Director.SkipSSLValidation)

	internalDirectorClientProvider := directorclient.NewClientProvider(cfg.Director.InternalURL, cfg.Director.ClientTimeout, cfg.Director.SkipSSLValidation)

	defer func() {
		err := closeFunc()
		exitOnError(err, "")
	}()

	formationsHandler := &notifications.FormationNotificationHandler{
		Transact:                         transact,
		DirectorGraphQLClient:            internalDirectorClientProvider.Client(),
		DirectorCertSecuredGraphQLClient: certSecuredGraphQLClient,
	}

	faHandler := &notifications.FANotificationHandler{
		Transact:                         transact,
		DirectorGraphQLClient:            internalDirectorClientProvider.Client(),
		DirectorCertSecuredGraphQLClient: certSecuredGraphQLClient,
	}

	processor := notifications.NewNotificationProcessor(cfg.Database, map[notifications.HandlerKey]notifications.NotificationHandler{
		{
			NotificationChannel: "events",
			ResourceType:        resource.Formations,
		}: formationsHandler,
		{
			NotificationChannel: "events",
			ResourceType:        resource.FormationAssignment,
		}: faHandler,
	})

	if err := processor.Run(ctx); err != nil {
		exitOnError(err, "")
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func NewCertAuthorizedGraphQLClientWithCustomURL(url string, key crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *gcli.Client {
	certAuthorizedClient := NewCertAuthorizedHTTPClient(key, rawCertChain, skipSSLValidation)
	return gcli.NewClient(url, gcli.WithHTTPClient(certAuthorizedClient))
}

func NewCertAuthorizedHTTPClient(key crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: rawCertChain,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: skipSSLValidation,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Second * 30,
	}

	return httpClient
}
