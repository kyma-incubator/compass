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

	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

type UnbindEndpoint struct {
	credentialsGetter  types.BundleCredentialsFetcher
	credentialsDeleter types.BundleCredentialsDeleteRequester
}

func NewUnbindEndpoint(credentialsGetter types.BundleCredentialsFetcher, credentialsDeleter types.BundleCredentialsDeleteRequester) *UnbindEndpoint {
	return &UnbindEndpoint{
		credentialsGetter:  credentialsGetter,
		credentialsDeleter: credentialsDeleter,
	}
}

func (b *UnbindEndpoint) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	log.C(ctx).Infof("Unbind instanceID: %s bindingID: %s details: %+v asyncAllowed: %v", instanceID, bindingID, details, asyncAllowed)

	if !asyncAllowed {
		return domain.UnbindSpec{}, apiresponses.ErrAsyncRequired
	}

	appID := details.ServiceID
	bundleID := details.PlanID
	logger := log.C(ctx).WithFields(map[string]interface{}{
		"appID":      appID,
		"bundleID":   bundleID,
		"instanceID": instanceID,
		"bindingID":  bindingID,
	})

	logger.Info("Fetching bundle instance credentials")

	resp, err := b.credentialsGetter.FetchBundleInstanceAuth(ctx, &director.BundleInstanceInput{
		InstanceAuthID: bindingID,
		Context: map[string]string{
			"instance_id": instanceID,
			"binding_id":  bindingID,
		},
	})
	if err != nil && !IsNotFoundError(err) {
		return domain.UnbindSpec{}, errors.Wrapf(err, "while getting bundle instance credentials from director")
	}

	if IsNotFoundError(err) {
		logger.Info("Bundle credentials for binding are already gone")
		return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
	}

	instanceAuth := resp.InstanceAuth

	status := instanceAuth.Status
	if IsUnused(status) {
		logger.Info("Bundle credentials for binding exist and are not used. Deletion is already in progress")
		return domain.UnbindSpec{
			IsAsync:       true,
			OperationData: string(UnbindOp),
		}, nil
	}

	logger.Info("Bundle credentials for binding exist and are used. Requesting deletion")
	deleteResp, err := b.credentialsDeleter.RequestBundleInstanceCredentialsDeletion(ctx, &director.BundleInstanceAuthDeletionInput{
		InstanceAuthID: instanceAuth.ID,
	})
	if err != nil {
		if IsNotFoundError(err) {
			logger.Info("Bundle credentials for binding are already gone")
			return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
		}

		return domain.UnbindSpec{}, errors.Wrapf(err, "while requesting bundle instance credentials deletion from director")
	}

	status = &deleteResp.Status
	logger.Infof("bundle instance credentials have status %+v", *status)

	logger.Info("Successfully requested deletion of bundle instance credentials")

	return domain.UnbindSpec{
		IsAsync:       true,
		OperationData: string(UnbindOp),
	}, nil
}
