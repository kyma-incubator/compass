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
	"encoding/base64"
	"fmt"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type DeprovisionEndpoint struct {
	credentialsGetter  packageCredentialsFetcherForInstance
	credentialsDeleter packageCredentialsDeleteRequester
}

func (b *DeprovisionEndpoint) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	log.C(ctx).Infof("Deprovision instanceID: %s asyncAllowed: %t force: %t", instanceID, asyncAllowed, details.Force)

	if !asyncAllowed {
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrAsyncRequired
	}

	appID := details.ServiceID
	packageID := details.PlanID
	logger := log.C(ctx).WithFields(map[string]interface{}{
		"appID":      appID,
		"packageID":  packageID,
		"instanceID": instanceID,
	})

	logger.Info("Fetching package instance credentials")

	resp, err := b.credentialsGetter.FindPackageInstanceCredentialsForContext(ctx, &director.FindPackageInstanceCredentialsByContextInput{
		ApplicationID: appID,
		PackageID:     packageID,
		Context: map[string]string{
			"instance_id": instanceID,
		},
	})
	if err != nil && !IsNotFoundError(err) {
		return domain.DeprovisionServiceSpec{}, errors.Wrapf(err, "while getting package instance credentials from director")
	}

	exists := !IsNotFoundError(err)
	if !exists {
		logger.Info("Package credentials for instance are already gone")
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	auths := resp.InstanceAuths
	if len(auths) != 1 {
		return domain.DeprovisionServiceSpec{}, errors.Errorf("expected 1 auth but got %d", len(auths))
	}
	auth := auths[0]
	status := auth.Status
	if IsUnused(status) {
		logger.Info("Package credentials for instance exist and are not used. Deletion is already in progress")
		return domain.DeprovisionServiceSpec{
			IsAsync:       true,
			OperationData: fmt.Sprintf("%s:%s:%s:%s", DeprovisionOp, appID, packageID, auth.ID),
		}, nil
	}

	logger.Info("Package credentials for instance exist and are used. Requesting deletion")
	deleteResp, err := b.credentialsDeleter.RequestPackageInstanceCredentialsDeletion(ctx, &director.RequestPackageInstanceAuthDeletionInput{
		InstanceAuthID: auth.ID,
	})
	if err != nil {
		if IsNotFoundError(err) {
			logger.Info("Package credentials for instance are already gone")
			return domain.DeprovisionServiceSpec{}, apiresponses.ErrInstanceDoesNotExist
		}

		return domain.DeprovisionServiceSpec{}, errors.Wrapf(err, "while requesting package instance credentials deletion from director")
	}

	status = &deleteResp.Status
	logger.Infof("package instance credentials have status %+v", *status)

	logger.Info("Successfully requested deletion of package instance credentials")

	op := fmt.Sprintf("%s:%s:%s:%s", DeprovisionOp, appID, packageID, auth.ID)
	encodedOp := base64.URLEncoding.EncodeToString([]byte(op))

	return domain.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: encodedOp,
	}, nil
}
