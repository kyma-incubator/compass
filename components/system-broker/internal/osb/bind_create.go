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

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/pkg/errors"
)

type BindEndpoint struct {
	credentialsCreator types.BundleCredentialsCreateRequester
	credentialsGetter  types.BundleCredentialsFetcher
}

func NewBindEndpoint(credentialsCreator types.BundleCredentialsCreateRequester, credentialsGetter types.BundleCredentialsFetcher) *BindEndpoint {
	return &BindEndpoint{
		credentialsCreator: credentialsCreator,
		credentialsGetter:  credentialsGetter,
	}
}

func (b *BindEndpoint) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	log.C(ctx).Infof("Bind instanceID: %s bindingID: %s parameters: %s context: %s asyncAllowed: %t", instanceID, bindingID, string(details.RawParameters), string(details.RawContext), asyncAllowed)

	if !asyncAllowed {
		return domain.Binding{}, apiresponses.ErrAsyncRequired
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
	var instanceAuth *schema.BundleInstanceAuth
	getResp, err := b.credentialsGetter.FetchBundleInstanceAuth(ctx, &director.BundleInstanceInput{
		InstanceAuthID: bindingID,
		Context: map[string]string{
			"instance_id": instanceID,
			"binding_id":  bindingID,
		},
	})
	if err != nil && !IsNotFoundError(err) {
		return domain.Binding{}, errors.Wrapf(err, "while getting bundle instance credentials from director")
	}
	exists := !IsNotFoundError(err)
	if !exists {
		logger.Info("Bundle credentials for binding do not exist. Requesting new credentials")

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

		createResp, err := b.credentialsCreator.RequestBundleInstanceCredentialsCreation(ctx, &director.BundleInstanceCredentialsInput{
			BundleID:    bundleID,
			AuthID:      bindingID,
			Context:     rawContext,
			InputSchema: rawParams,
		})
		if err != nil {
			return domain.Binding{}, errors.Wrap(err, "while requesting bundle instance credentials creation from director")
		}
		instanceAuth = createResp.InstanceAuth
	} else {
		instanceAuth = getResp.InstanceAuth
	}

	logger.Infof("bundle instance credentials have status %s", instanceAuth.Status.Condition)

	if IsFailed(instanceAuth.Status) {
		return domain.Binding{}, errors.Errorf("requesting bundle instance credentials from director failed, "+
			"got status %+v", *instanceAuth.Status)
	}

	logger.Info("Successfully found bundle instance credentials")

	return domain.Binding{
		IsAsync:       true,
		OperationData: string(BindOp),
	}, nil
}
