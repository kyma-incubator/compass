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
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pivotal-cf/brokerapi/v7/domain"
)

type GetInstanceEndpoint struct {
}

func (b *GetInstanceEndpoint) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	log.C(ctx).Infof("GetInstanceEndpoint instanceID: %s", instanceID)

	return domain.GetInstanceDetailsSpec{}, errors.New("not supported")
}
