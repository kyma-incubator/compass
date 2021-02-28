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

package controllers

import (
	"context"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	typesbroker "github.com/kyma-incubator/compass/components/system-broker/pkg/types"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrReconciliationTimeoutReached = errors.New("reconciliation timeout reached")
	ErrWebhookTimeoutReached        = errors.New("webhook timeout reached")
	ErrFailedWebhookStatus          = errors.New("webhook operation has finished with failed status")
	ErrUnsupportedWebhookMode       = errors.New("unsupported webhook mode")
)

// StatusManager defines an abstraction for managing the status of a given kubernetes resource
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . StatusManager
type StatusManager interface {
	Initialize(ctx context.Context, name types.NamespacedName) error
	InProgressWithPollURL(ctx context.Context, name types.NamespacedName, pollURL string) error
	InProgressWithPollURLAndLastPollTimestamp(ctx context.Context, name types.NamespacedName, pollURL, lastPollTimestamp string, retryCount int) error
	SuccessStatus(ctx context.Context, name types.NamespacedName) error
	FailedStatus(ctx context.Context, name types.NamespacedName, errorMsg string) error
}

// KubernetesClient is a defines a Kubernetes client capable of retrieving and deleting resources as well as updating their status
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KubernetesClient
type KubernetesClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.Operation, error)
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
}

// DirectorClient defines a Director client which is capable of fetching an application
// and notifying Director for operation state changes
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DirectorClient
type DirectorClient interface {
	typesbroker.ApplicationLister
	UpdateOperation(ctx context.Context, request *director.Request) error
}

// WebhookClient defines a general purpose Webhook executor client
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . WebhookClient
type WebhookClient interface {
	Do(ctx context.Context, request *webhook.Request) (*webhookdir.Response, error)
	Poll(ctx context.Context, request *webhook.PollRequest) (*webhookdir.ResponseStatus, error)
}
