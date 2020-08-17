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
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	errors "github.com/pkg/errors"
)

type BindEndpoint struct {
	credentialsGetter packageCredentialsFetcherForInstance
}

func (b *BindEndpoint) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	log.C(ctx).Infof("Bind instanceID: %s bindingID: %s parameters: %s context: %s asyncAllowed: %s", instanceID, bindingID, string(details.RawParameters), string(details.RawContext), asyncAllowed)

	appID := details.ServiceID
	packageID := details.PlanID
	logger := log.C(ctx).WithFields(map[string]interface{}{
		"appID":      appID,
		"packageID":  packageID,
		"instanceID": instanceID,
		"bindingID":  bindingID,
	})

	logger.Info("Fetching package instance credentials")

	resp, err := b.credentialsGetter.FindPackageInstanceCredentialsForContext(ctx, &director.FindPackageInstanceCredentialsByContextInput{
		ApplicationID: appID,
		PackageID:     packageID,
		Context: map[string]string{
			"instance_id": instanceID,
		},
	})
	if err != nil {
		return domain.Binding{}, errors.Wrapf(err, "while getting package instance credentials from director")
	}
	auth := resp.InstanceAuths[0]

	if !IsSucceeded(auth.Status) {
		return domain.Binding{}, errors.Wrapf(err, "credentials status is not success: %+v", *auth.Status)
	}

	bindingCredentials, err := mapPackageInstanceAuthToModel(*auth)
	if err != nil {
		return domain.Binding{}, errors.Wrap(err, "while mapping to binding credentials")
	}

	formattedCredentials, err := format(bindingCredentials)
	if err != nil {
		return domain.Binding{}, errors.Wrap(err, "while converting to raw map")
	}

	logger.Info("Successfully obtained binding details for package instance credentials")

	return domain.Binding{
		Credentials: formattedCredentials,
	}, nil
}

func format(c BindingCredentials) (map[string]interface{}, error) {
	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling")
	}

	var mapConfig map[string]interface{}
	if err := json.Unmarshal(jsonConfig, &mapConfig); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling")
	}

	return map[string]interface{}{
		"CREDENTIALS_TYPE": c.Type,
		"CONFIGURATION":    mapConfig,
	}, nil
}
