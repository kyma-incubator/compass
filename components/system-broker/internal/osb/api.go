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
	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi/v7"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
)

func API(rootAPI string, serviceBroker domain.ServiceBroker, logger lager.Logger) func(router *mux.Router) {
	return func(router *mux.Router) {

		r := router.PathPrefix(rootAPI).Subrouter()

		r.Use(middlewares.AddCorrelationIDToContext)
		r.Use(middlewares.AddOriginatingIdentityToContext)
		r.Use(middlewares.APIVersionMiddleware{LoggerFactory: logger}.ValidateAPIVersionHdr)
		r.Use(middlewares.AddInfoLocationToContext)

		brokerapi.AttachRoutes(r, serviceBroker, logger)
	}
}
