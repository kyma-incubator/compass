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

package status

import (
	"context"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusUpdaterFunc func(operation *v1alpha1.Operation)

// manager implements the StatusManager interface
type manager struct {
	k8sClient client.Client
}

// NewManager constructs a manager instance
func NewManager(k8sClient client.Client) *manager {
	return &manager{
		k8sClient: k8sClient,
	}
}

// Initialize sets the initial status of an Operation CR.
// The method executes only if the generation of the Operation CR mismatches the observed generation in the status,
// which allows the method to be used on the same resource over and over, for example when consecutive async requests
// are scheduled on the same Operation CR and the old status should be wiped off.
func (m *manager) Initialize(operation *v1alpha1.Operation) error {
	if err := operation.Validate(); err != nil {
		return err
	}

	status := &operation.Status
	if status.ObservedGeneration != nil && operation.ObjectMeta.Generation == *status.ObservedGeneration {
		return nil
	}

	status.ObservedGeneration = &operation.Generation

	status.Phase = v1alpha1.StateInProgress
	status.Conditions = []v1alpha1.Condition{
		{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionFalse},
		{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionFalse},
	}

	if len(operation.Spec.WebhookIDs) > 0 {
		status.Webhooks = []v1alpha1.Webhook{
			{WebhookID: operation.Spec.WebhookIDs[0], State: v1alpha1.StateInProgress},
		}
	}

	status.InitializedAt = metav1.Now()
	return nil
}

// InProgressWithPollURL sets the status of an Operation CR to In Progress, ensures that none of the conditions are set to True,
// and also initializes the slice of webhooks in the status with a single webhook with the provided pollURL.
func (m *manager) InProgressWithPollURL(ctx context.Context, operation *v1alpha1.Operation, pollURL string) error {
	return m.InProgressWithPollURLAndLastPollTimestamp(ctx, operation, pollURL, "", 0)
}

// InProgressWithPollURLAndLastPollTimestamp builds on what InProgressWithPollURL does, but also sets the last poll timestamp and retry count for the given webhook.
func (m *manager) InProgressWithPollURLAndLastPollTimestamp(ctx context.Context, operation *v1alpha1.Operation, pollURL, lastPollTimestamp string, retryCount int) error {
	return m.updateStatusFunc(ctx, operation, func(operation *v1alpha1.Operation) {
		status := &operation.Status

		status.Phase = v1alpha1.StateInProgress
		status.Conditions = []v1alpha1.Condition{
			{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionFalse},
			{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionFalse},
		}

		if len(operation.Spec.WebhookIDs) > 0 {
			status.Webhooks = []v1alpha1.Webhook{
				{WebhookID: operation.Spec.WebhookIDs[0], State: v1alpha1.StateInProgress, WebhookPollURL: pollURL, LastPollTimestamp: lastPollTimestamp, RetriesCount: retryCount},
			}
		}
	})
}

// SuccessStatus sets the status of an Operation CR to Success, ensures that the Ready condition is True, the Error condition is False,
// and that the webhook part of the webhooks slice in the status is marked with Success.
func (m *manager) SuccessStatus(ctx context.Context, operation *v1alpha1.Operation) error {
	return m.updateStatusFunc(ctx, operation, func(operation *v1alpha1.Operation) {
		status := &operation.Status

		status.Phase = v1alpha1.StateSuccess
		status.Conditions = []v1alpha1.Condition{
			{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionTrue},
			{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionFalse},
		}

		if len(operation.Spec.WebhookIDs) > 0 {
			webhook := v1alpha1.Webhook{WebhookID: operation.Spec.WebhookIDs[0], State: v1alpha1.StateSuccess}
			if len(operation.Status.Webhooks) > 0 {
				webhook.RetriesCount = operation.Status.Webhooks[0].RetriesCount
				webhook.WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
				webhook.LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			}
			status.Webhooks = []v1alpha1.Webhook{webhook}
		}
	})
}

// FailedStatus sets the status of an Operation CR to Failed, ensures that the Ready condition is False, the Error condition is True,
// and that the webhook part of the webhooks slice in the status is marked with Failed.
func (m *manager) FailedStatus(ctx context.Context, operation *v1alpha1.Operation, errorMsg string) error {
	return m.updateStatusFunc(ctx, operation, func(operation *v1alpha1.Operation) {
		status := &operation.Status

		status.Phase = v1alpha1.StateFailed
		status.Conditions = []v1alpha1.Condition{
			{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionTrue},
			{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionTrue, Message: errorMsg},
		}

		if len(operation.Spec.WebhookIDs) > 0 {
			webhook := v1alpha1.Webhook{WebhookID: operation.Spec.WebhookIDs[0], State: v1alpha1.StateFailed}
			if len(operation.Status.Webhooks) > 0 {
				webhook.RetriesCount = operation.Status.Webhooks[0].RetriesCount
				webhook.WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
				webhook.LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			}

			status.Webhooks = []v1alpha1.Webhook{webhook}
		}
	})
}

func (m *manager) updateStatusFunc(ctx context.Context, operation *v1alpha1.Operation, statusUpdaterFunc statusUpdaterFunc) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := operation.Validate(); err != nil {
			return err
		}

		statusUpdaterFunc(operation)

		return m.k8sClient.Status().Update(ctx, operation)
	})
}
