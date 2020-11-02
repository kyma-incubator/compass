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
	"encoding/json"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"strings"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type BindLastOperationEndpoint struct {
	credentialsGetter packageCredentialsFetcher
}

func (b *BindLastOperationEndpoint) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	log.C(ctx).Infof("LastBindingOperation instanceID: %s bindingID: %s details: %+v", instanceID, bindingID, details)

	opBytes, err := base64.URLEncoding.DecodeString(details.OperationData)
	if err != nil {
		return domain.LastOperation{}, errors.Wrapf(err, "while decoding base64 encoded op %s", details.OperationData)
	}

	args := strings.Split(string(opBytes), ":")
	if len(args) != 4 {
		return domain.LastOperation{}, errors.Errorf("operation must contain 4 segments separated by : but was %s", details.OperationData)
	}
	opType := args[0]
	appID := args[1]
	packageID := args[2]
	authID := args[3]

	logger := log.C(ctx).WithFields(map[string]interface{}{
		"opType":     opType,
		"appID":      appID,
		"packageID":  packageID,
		"authID":     authID,
		"instanceID": instanceID,
		"bindingID":  bindingID,
	})

	logger.Info("Fetching package instance credentials")
	resp, err := b.credentialsGetter.FindPackageInstanceCredentials(ctx, &director.FindPackageInstanceCredentialInput{
		PackageID:      packageID,
		ApplicationID:  appID,
		InstanceAuthID: authID,
	})
	if err != nil && !IsNotFoundError(err) {
		return domain.LastOperation{}, errors.Wrapf(err, "while getting package instance credentials from director")
	}

	if IsNotFoundError(err) {
		logger.Info("Package instance credentials not found")
		if opType == string(UnbindOp) {
			return domain.LastOperation{
				State:       domain.Succeeded,
				Description: "credentials were successfully deleted",
			}, nil
		}
		return domain.LastOperation{}, apiresponses.ErrBindingDoesNotExist
	}

	auth := resp.InstanceAuth

	if auth.Context == nil {
		return domain.LastOperation{}, errors.New("auth context must contain service binding id but was empty")
	}
	var authContext map[string]string
	if err := json.Unmarshal([]byte(*auth.Context), &authContext); err != nil {
		return domain.LastOperation{}, errors.Wrap(err, "while unmarshaling auth context")
	}

	if authContext["instance_id"] != instanceID && authContext["binding_id"] != bindingID {
		logger.Infof("Package instance credentials instance id in context %s and/or binding id in context %s do not match instance id and binding id from request", authContext["instance_id"], authContext["binding_id"])
		return domain.LastOperation{}, apiresponses.ErrBindingDoesNotExist
	}

	logger.Infof("Found package credentials during poll last op with status %+v", *auth.Status)

	var state domain.LastOperationState
	var opErr error
	switch opType {
	case string(BindOp):
		switch auth.Status.Condition {
		case schema.PackageInstanceAuthStatusConditionSucceeded: // success
			state = domain.Succeeded
		case schema.PackageInstanceAuthStatusConditionPending: // in progress
			state = domain.InProgress
		case schema.PackageInstanceAuthStatusConditionFailed: // failed
			// this would trigger orphan mitigation
			state = domain.Failed
		case schema.PackageInstanceAuthStatusConditionUnused: // error
			// pretty questionable status, may happen if deprovisioning is triggered before async provisioning succeeds
			//TODO do we want to trigger orphan mitigation, just return error or force platform to continue polling here?
			fallthrough
		default:
			// this should force platform to continue polling, should be the more flexiable approach
			opErr = errors.Errorf("operation reached unexpected state: op %s, status %+v", opType, *auth.Status)
		}
	case string(UnbindOp):
		//TODO strict condition checks vs genericly force platform to continue polling
		//switch auth.Status.Condition {
		//case schema.PackageInstanceAuthStatusConditionFailed: // failed
		//case schema.PackageInstanceAuthStatusConditionUnused: // in progress
		//case schema.PackageInstanceAuthStatusConditionSucceeded: // error/unexpected
		//case schema.PackageInstanceAuthStatusConditionPending: // error/unexpected
		//}
		// this would be the more flexible approach (platform continues to poll)
		// actual "success" return is above the switch/410 gone
		state = domain.InProgress
	}

	if opErr != nil {
		return domain.LastOperation{}, opErr
	}

	return domain.LastOperation{
		State:       state,
		Description: auth.Status.Message,
	}, nil
}
