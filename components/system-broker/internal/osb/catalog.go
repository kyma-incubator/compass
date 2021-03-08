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
	"context"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . converter
type Converter interface {
	Convert(app *schema.ApplicationExt) (*domain.Service, error)
}

type CatalogEndpoint struct {
	lister    types.ApplicationsLister
	converter Converter
}

func NewCatalogEndpoint(l types.ApplicationsLister, c Converter) *CatalogEndpoint {
	return &CatalogEndpoint{l, c}
}

func (b *CatalogEndpoint) Services(ctx context.Context) ([]domain.Service, error) {
	resp := make([]domain.Service, 0)

	applications, err := b.lister.FetchApplications(ctx)
	if err != nil {
		//broker api does not log catalog errors
		err := errors.Wrap(err, "while listing applications from director")
		log.C(ctx).WithError(err).Error("catalog failure")
		return nil, errors.Wrap(err, "could not build catalog")
	}

	for _, app := range applications.Result.Data {
		if app == nil {
			continue
		}

		svc, err := b.converter.Convert(app)
		if err != nil {
			return nil, errors.Wrap(err, "while converting application to OSB service")
		}

		if len(svc.Plans) > 0 {
			resp = append(resp, *svc)
		}
	}

	return resp, nil
}
