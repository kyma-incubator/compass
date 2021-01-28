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

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type DeprovisionEndpoint struct {
}

func NewDeprovisionEndpoint() *DeprovisionEndpoint {
	return &DeprovisionEndpoint{}
}

func (b *DeprovisionEndpoint) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	log.C(ctx).Infof("Deprovision instance with instanceID: %s, serviceID: %s, planID %s, asyncAllowed: %t force: %t", instanceID, details.ServiceID, details.PlanID, asyncAllowed, details.Force)

	return domain.DeprovisionServiceSpec{}, nil
}
