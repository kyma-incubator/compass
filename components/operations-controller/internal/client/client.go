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

package client

import (
	"context"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Client
type Client interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.Operation, error)
	UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
}

type defaultClient struct {
	client.Client
}

func New(client client.Client) *defaultClient {
	return &defaultClient{Client: client}
}

func (dc *defaultClient) Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.Operation, error) {
	var operation = &v1alpha1.Operation{}
	err := dc.Client.Get(ctx, key, operation)
	if err != nil {
		return nil, err
	}
	return operation, nil
}

func (dc *defaultClient) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return dc.Status().Update(ctx, obj, opts...)
}
