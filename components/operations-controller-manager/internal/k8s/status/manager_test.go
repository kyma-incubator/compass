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

package status_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	webhookID     = "866e6b9c-f03b-442b-a6a5-4b90e21e503a"
	mockedPollURL = "https://test-domain.com/operation"
)

func TestStatusManager(t *testing.T) {
	ctx := context.Background()

	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer func() {
		err := testEnv.Stop()
		require.NoError(t, err)
	}()

	err = v1alpha1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	statusManager := status.NewManager(k8sClient)

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operation",
			Namespace: "default",
		},
		Spec: v1alpha1.OperationSpec{
			OperationType: "Delete",
			WebhookIDs:    []string{webhookID},
		},
	}

	err = k8sClient.Create(ctx, operation)
	require.NoError(t, err)

	namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

	t.Run("Test Initialize when operation CR is in invalid state should return Validation Error", func(t *testing.T) {
		invalidOperation := operation.DeepCopy()
		invalidOperation.ResourceVersion = ""
		invalidOperation.ObjectMeta.Name = "invalid-operation"
		invalidOperation.Spec.WebhookIDs = make([]string, 2)

		err = k8sClient.Create(ctx, invalidOperation)
		require.NoError(t, err)

		err = statusManager.Initialize(invalidOperation)
		require.Error(t, err)

		_, isValErr := err.(*v1alpha1.OperationValidationErr)
		require.True(t, isValErr)
		require.Contains(t, err.Error(), "expected 0 or 1 webhook for execution, found: 2")
	})

	t.Run("Test Initialize when generation and observed generation mismatch should initialize status with initial values", func(t *testing.T) {
		originOperation := operation.DeepCopy()
		err = statusManager.Initialize(originOperation)
		require.NoError(t, err)

		require.Equal(t, int64(1), *originOperation.Status.ObservedGeneration)
		require.Equal(t, v1alpha1.StateInProgress, originOperation.Status.Phase)

		require.Len(t, originOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, originOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateInProgress, originOperation.Status.Webhooks[0].State)
		require.Equal(t, 0, originOperation.Status.Webhooks[0].RetriesCount)
		require.Empty(t, originOperation.Status.Webhooks[0].WebhookPollURL)
		require.Empty(t, originOperation.Status.Webhooks[0].LastPollTimestamp)

		require.Len(t, originOperation.Status.Conditions, 2)
		require.Equal(t, corev1.ConditionFalse, originOperation.Status.Conditions[0].Status)
		require.Equal(t, corev1.ConditionFalse, originOperation.Status.Conditions[1].Status)
		require.Empty(t, originOperation.Status.Conditions[0].Message)
		require.Empty(t, originOperation.Status.Conditions[1].Message)
	})

	t.Run("Test Initialize when generation and observed generation match shouldn not affect status", func(t *testing.T) {
		originOperation := operation.DeepCopy()
		err = statusManager.Initialize(originOperation)
		require.NoError(t, err)

		err = statusManager.SuccessStatus(ctx, originOperation)
		require.NoError(t, err)

		err = statusManager.Initialize(originOperation)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		require.Equal(t, v1alpha1.StateSuccess, actualOperation.Status.Phase)

		require.Len(t, actualOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, actualOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateSuccess, actualOperation.Status.Webhooks[0].State)

		require.Len(t, actualOperation.Status.Conditions, 2)
		for _, condition := range actualOperation.Status.Conditions {
			if condition.Type == v1alpha1.ConditionTypeReady {
				require.Equal(t, corev1.ConditionTrue, condition.Status)
				require.Empty(t, condition.Message)
			}

			if condition.Type == v1alpha1.ConditionTypeError {
				require.Equal(t, corev1.ConditionFalse, condition.Status)
				require.Empty(t, condition.Message)
			}
		}
	})

	t.Run("Test In Progress with Poll URL", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		err = statusManager.InProgressWithPollURL(ctx, originOperation, mockedPollURL)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateInProgress, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateInProgress, op.Status.Webhooks[0].State)
			require.Equal(t, 0, op.Status.Webhooks[0].RetriesCount)
			require.Equal(t, mockedPollURL, op.Status.Webhooks[0].WebhookPollURL)
			require.Empty(t, op.Status.Webhooks[0].LastPollTimestamp)

			require.Len(t, op.Status.Conditions, 2)
			require.Equal(t, corev1.ConditionFalse, op.Status.Conditions[0].Status)
			require.Equal(t, corev1.ConditionFalse, op.Status.Conditions[1].Status)
			require.Empty(t, op.Status.Conditions[0].Message)
			require.Empty(t, op.Status.Conditions[1].Message)
		}
	})

	t.Run("Test In Progress with Poll URL And Last Poll Tiimestamp should succeed", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		retryCount := 1
		lastPollTimestamp := time.Now().Format(time.RFC3339Nano)
		err = statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, originOperation, mockedPollURL, lastPollTimestamp, retryCount)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateInProgress, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateInProgress, op.Status.Webhooks[0].State)
			require.Equal(t, retryCount, op.Status.Webhooks[0].RetriesCount)
			require.Equal(t, mockedPollURL, op.Status.Webhooks[0].WebhookPollURL)
			require.Equal(t, lastPollTimestamp, op.Status.Webhooks[0].LastPollTimestamp)

			require.Len(t, op.Status.Conditions, 2)
			require.Equal(t, corev1.ConditionFalse, op.Status.Conditions[0].Status)
			require.Equal(t, corev1.ConditionFalse, op.Status.Conditions[1].Status)
			require.Empty(t, op.Status.Conditions[0].Message)
			require.Empty(t, op.Status.Conditions[1].Message)
		}
	})

	t.Run("Test Success Status should succeed", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		err = statusManager.SuccessStatus(ctx, originOperation)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateSuccess, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateSuccess, op.Status.Webhooks[0].State)

			require.Len(t, op.Status.Conditions, 2)
			for _, condition := range op.Status.Conditions {
				if condition.Type == v1alpha1.ConditionTypeReady {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Empty(t, condition.Message)
				}

				if condition.Type == v1alpha1.ConditionTypeError {
					require.Equal(t, corev1.ConditionFalse, condition.Status)
					require.Empty(t, condition.Message)
				}
			}
		}
	})

	t.Run("Test Success Status after In Progress with Poll URL And Last Poll Timestamp should not discard Poll URL and Last Poll Timestamp from the status", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		retryCount := 1
		lastPollTimestamp := time.Now().Format(time.RFC3339Nano)
		err = statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, originOperation, mockedPollURL, lastPollTimestamp, retryCount)
		require.NoError(t, err)

		err = statusManager.SuccessStatus(ctx, originOperation)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateSuccess, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateSuccess, op.Status.Webhooks[0].State)
			require.Equal(t, retryCount, op.Status.Webhooks[0].RetriesCount)
			require.Equal(t, mockedPollURL, op.Status.Webhooks[0].WebhookPollURL)
			require.Equal(t, lastPollTimestamp, op.Status.Webhooks[0].LastPollTimestamp)

			require.Len(t, op.Status.Conditions, 2)
			for _, condition := range op.Status.Conditions {
				if condition.Type == v1alpha1.ConditionTypeReady {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Empty(t, condition.Message)
				}

				if condition.Type == v1alpha1.ConditionTypeError {
					require.Equal(t, corev1.ConditionFalse, condition.Status)
					require.Empty(t, condition.Message)
				}
			}
		}
	})

	t.Run("Test Failed Status should succeed", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		errMsg := "test error"
		err = statusManager.FailedStatus(ctx, originOperation, errMsg)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateFailed, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateFailed, op.Status.Webhooks[0].State)

			require.Len(t, op.Status.Conditions, 2)
			for _, condition := range op.Status.Conditions {
				if condition.Type == v1alpha1.ConditionTypeReady {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Empty(t, condition.Message)
				}

				if condition.Type == v1alpha1.ConditionTypeError {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Equal(t, errMsg, condition.Message)
				}
			}
		}
	})

	t.Run("Test Failed Status after In Progress with Poll URL And Last Poll Timestamp should not discard Poll URL and Last Poll Timestamp from the status", func(t *testing.T) {
		var originOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, originOperation)
		require.NoError(t, err)

		retryCount := 1
		lastPollTimestamp := time.Now().Format(time.RFC3339Nano)
		err = statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, originOperation, mockedPollURL, lastPollTimestamp, retryCount)
		require.NoError(t, err)

		errMsg := "test error"
		err = statusManager.FailedStatus(ctx, originOperation, errMsg)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		for _, op := range []*v1alpha1.Operation{originOperation, actualOperation} {
			require.Equal(t, v1alpha1.StateFailed, op.Status.Phase)

			require.Len(t, op.Status.Webhooks, 1)
			require.Equal(t, webhookID, op.Status.Webhooks[0].WebhookID)
			require.Equal(t, v1alpha1.StateFailed, op.Status.Webhooks[0].State)
			require.Equal(t, retryCount, op.Status.Webhooks[0].RetriesCount)
			require.Equal(t, mockedPollURL, op.Status.Webhooks[0].WebhookPollURL)
			require.Equal(t, lastPollTimestamp, op.Status.Webhooks[0].LastPollTimestamp)

			require.Len(t, op.Status.Conditions, 2)
			for _, condition := range op.Status.Conditions {
				if condition.Type == v1alpha1.ConditionTypeReady {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Empty(t, condition.Message)
				}

				if condition.Type == v1alpha1.ConditionTypeError {
					require.Equal(t, corev1.ConditionTrue, condition.Status)
					require.Equal(t, errMsg, condition.Message)
				}
			}
		}
	})
}
