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

	"github.com/kyma-incubator/compass/components/system-broker/internal/director"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type GetBindingEndpoint struct {
	credentialsGetter types.BundleCredentialsFetcherForInstance
}

func NewGetBindingEndpoint(credentialsGetter types.BundleCredentialsFetcherForInstance) *GetBindingEndpoint {
	return &GetBindingEndpoint{
		credentialsGetter: credentialsGetter,
	}
}

func (b *GetBindingEndpoint) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	log.C(ctx).Infof("GetBindingEndpoint instanceID: %s bindingID: %s", instanceID, bindingID)

	logger := log.C(ctx).WithFields(map[string]interface{}{
		"instanceID": instanceID,
		"bindingID":  bindingID,
	})

	logger.Debug("Fetching bundle instance credentials")

	resp, err := b.credentialsGetter.FetchBundleInstanceCredentials(ctx, &director.BundleInstanceInput{
		InstanceAuthID: bindingID,
		Context: map[string]string{
			"instance_id": instanceID,
			"binding_id":  bindingID,
		},
	})

	if err != nil && !IsNotFoundError(err) {
		return domain.GetBindingSpec{}, errors.Wrapf(err, "while getting bundle instance credentials from director")
	}

	if IsNotFoundError(err) {
		logger.Debug("Bundle credentials for binding were not found")
		return domain.GetBindingSpec{}, apiresponses.ErrBindingNotFound
	}

	instanceAuth := resp.InstanceAuth

	switch instanceAuth.Status.Condition {
	case schema.BundleInstanceAuthStatusConditionPending:
		logger.Info("Bundle credentials for binding are still pending")
		return domain.GetBindingSpec{}, apiresponses.ErrBindingNotFound
	case schema.BundleInstanceAuthStatusConditionUnused:
		logger.Info("Bundle credentials for binding are unused")
		return domain.GetBindingSpec{}, apiresponses.ErrBindingNotFound
	case schema.BundleInstanceAuthStatusConditionFailed:
		logger.Info("Bundle credentials for binding are in failed state")
		return domain.GetBindingSpec{}, errors.Errorf("credentials status is not success: %+v", *instanceAuth.Status)
	default:
	}

	bindingCredentials, err := mapBundleInstanceAuthToModel(*instanceAuth, resp.TargetURLs)
	if err != nil {
		return domain.GetBindingSpec{}, errors.Wrap(err, "while mapping to binding credentials")
	}

	logger.Info("Successfully obtained binding details for bundle instance credentials")

	return domain.GetBindingSpec{
		Credentials: bindingCredentials,
	}, nil
}
