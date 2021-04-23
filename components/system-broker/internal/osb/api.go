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

package osb

import (
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/system-broker/internal/metrics"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pivotal-cf/brokerapi/v7"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
)

func API(rootAPI string, serviceBroker domain.ServiceBroker, logger lager.Logger, c *metrics.Collector) func(router *mux.Router) {
	return func(router *mux.Router) {

		r := router.PathPrefix(rootAPI).Subrouter()

		r.Use(middlewares.AddCorrelationIDToContext)
		r.Use(middlewares.AddOriginatingIdentityToContext)
		r.Use(middlewares.APIVersionMiddleware{LoggerFactory: logger}.ValidateAPIVersionHdr)
		r.Use(middlewares.AddInfoLocationToContext)
		r.Use(http.UnauthorizedMiddleware())

		brokerapi.AttachRoutes(r, serviceBroker, logger)

		err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			t, err := route.GetPathTemplate()
			if err != nil {
				return err
			}

			m, err := route.GetMethods()
			if err != nil {
				return err
			}

			methods := strings.Join(m, " ")
			instrumentation, found := metrics.PathToInstrumentationMapping[metrics.Path{
				PathTemplate: t,
				HTTPMethods:  methods,
			}]
			if found {
				route.Handler(instrumentation(c, route.GetHandler()))
			}
			return nil
		})
		if err != nil {
			log.D().Fatal(err.Error())
		}
	}
}
