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
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pkg/errors"
)

type GetBindingEndpoint struct {
	credentialsGetter packageCredentialsFetcherForInstance
}

func (b *GetBindingEndpoint) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	log.C(ctx).Infof("GetBindingEndpoint instanceID: %s bindingID: %s", instanceID, bindingID)

	details := struct { // TODO: Missing from pivotal-cf/brokerapi library...
		ServiceID string
		PlanID    string
	}{
		ServiceID: "",
		PlanID:    "",
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

	resp, err := b.credentialsGetter.FindPackageInstanceCredentialsForContext(ctx, &director.FindPackageInstanceCredentialsByContextInput{ // TODO: Use queryPackageInstanceAuth + saanity check
		ApplicationID: appID,
		PackageID:     packageID,
		Context: map[string]string{
			"instance_id": instanceID,
			"binding_id":  bindingID,
		},
	})
	if err != nil {
		return domain.GetBindingSpec{}, errors.Wrapf(err, "while getting package instance credentials from director")
	}

	if len(resp.InstanceAuths) != 1 {
		return domain.GetBindingSpec{}, errors.Errorf("expected 1 auth but got %d", len(resp.InstanceAuths))
	}

	auth := resp.InstanceAuths[0]

	if !IsSucceeded(auth.Status) {
		return domain.GetBindingSpec{}, errors.Wrapf(err, "credentials status is not success: %+v", *auth.Status)
	}

	bindingCredentials, err := mapPackageInstanceAuthToModel(*auth, resp.TargetURLs)
	if err != nil {
		return domain.GetBindingSpec{}, errors.Wrap(err, "while mapping to binding credentials")
	}

	logger.Info("Successfully obtained binding details for package instance credentials")

	return domain.GetBindingSpec{
		Credentials: bindingCredentials,
	}, nil
}
