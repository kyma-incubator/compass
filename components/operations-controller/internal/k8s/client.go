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

package k8s

import (
	"context"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// client implements KubernetesClient and acts as a wrapper of the default kubernetes controller client
type client struct {
	ctrl_client.Client
}

// New constructs a new client instance
func NewClient(ctrlClient ctrl_client.Client) *client {
	return &client{Client: ctrlClient}
}

// Get wraps the default kubernetes controller client Get method
func (c *client) Get(ctx context.Context, key ctrl_client.ObjectKey) (*v1alpha1.Operation, error) {
	var operation = &v1alpha1.Operation{}
	err := c.Client.Get(ctx, key, operation)
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// UpdateStatus wraps the default kubernetes controller client Status().Update method
func (c *client) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...ctrl_client.UpdateOption) error {
	return c.Status().Update(ctx, obj, opts...)
}
