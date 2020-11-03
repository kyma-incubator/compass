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
	"encoding/json"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type BindEndpoint struct {
	credentialsCreator packageCredentialsCreateRequester
	credentialsGetter  packageCredentialsFetcher
}

// option 1 - have a storage for operations and spawn a go routine during provision that will eventually update the op status, poll last op api just fetches op from storage; benefit - no switch case in last op handler
// option 2 - have naming pattern on operation id; no operation storage and no go routine during provision, instead, fetch creds during poll last op (op_id returned to platform returns all credential coordinates so that during poll last op we can see the credentials status); drawback - magic op_id pattern, switch case in last op handler
// option 3 - have a storage for operations and operations also store credential coordinates (appid,packageid), in provision op in inserted, in poll last op, op is fetched and based on coords creds are fetched and status is checked - drawbacks - storage is needed, we still have a switch on poll last op
// preferred option is 2 - requires less cpu/memory (no extra background go routines) and also does not need persistent storage
func (b *BindEndpoint) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	log.C(ctx).Infof("Bind instanceID: %s bindingID: %s parameters: %s context: %s asyncAllowed: %t", instanceID, bindingID, string(details.RawParameters), string(details.RawContext), asyncAllowed)

	if !asyncAllowed {
		return domain.Binding{}, apiresponses.ErrAsyncRequired
	}

	appID := details.ServiceID
	packageID := details.PlanID
	logger := log.C(ctx).WithFields(map[string]interface{}{
		"appID":      appID,
		"packageID":  packageID,
		"instanceID": instanceID,
		"bindingID":  bindingID,
	})

	logger.Info("Fetching package instance credentials")
	var auths []*schema.PackageInstanceAuth
	getResp, err := b.credentialsGetter.FindPackageInstanceCredentials(ctx, &director.FindPackageInstanceCredentialInput{
		InstanceAuthID: bindingID,
		Context: map[string]string{
			"instance_id": instanceID,
			"binding_id":  bindingID,
		},
	})
	if err != nil && !IsNotFoundError(err) {
		return domain.Binding{}, errors.Wrapf(err, "while getting package instance credentials from director")
	}
	exists := !IsNotFoundError(err)
	if !exists {
		logger.Info("Package credentials for binding do not exist. Requesting new credentials")

		rawParams := director.Values{}
		if details.RawParameters != nil {
			if err := json.Unmarshal(details.RawParameters, &rawParams); err != nil {
				return domain.Binding{}, errors.Wrap(err, "while unmarshaling raw parameters")
			}
		}

		rawContext := director.Values{}
		if details.RawContext != nil {
			if err := json.Unmarshal(details.RawContext, &rawContext); err != nil {
				return domain.Binding{}, errors.Wrap(err, "while unmarshaling raw context")
			}
		}

		rawContext["instance_id"] = instanceID
		rawContext["binding_id"] = bindingID

		createResp, err := b.credentialsCreator.RequestPackageInstanceCredentialsCreation(ctx, &director.RequestPackageInstanceCredentialsInput{ // TODO: Use createPackageInstanceAuth
			PackageID:   packageID,
			Context:     rawContext,
			InputSchema: rawParams,
		})
		if err != nil {
			return domain.Binding{}, errors.Wrap(err, "while requesting package instance credentials creation from director")
		}
		auths = []*schema.PackageInstanceAuth{createResp.InstanceAuth}
	} else {
		auths = getResp.InstanceAuths
	}

	if len(auths) != 1 {
		return domain.Binding{}, errors.Errorf("expected 1 auth but got %d", len(auths))
	}
	auth := auths[0]

	logger.Infof("package instance credentials have status %s", auth.Status.Condition)

	if IsFailed(auths[0].Status) {
		return domain.Binding{}, errors.Errorf("requesting package instance credentials from director failed, got status %+v", *auth.Status)
	}

	logger.Info("Successfully found package instance credentials")

	return domain.Binding{
		IsAsync:       true,
		AlreadyExists: exists,
		OperationData: string(BindOp),
	}, nil
}
