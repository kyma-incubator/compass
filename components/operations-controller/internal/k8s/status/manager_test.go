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
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
	"time"
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

	t.Run("Test Initialize when operation CR is in invalid state", func(t *testing.T) {
		invalidOperation := operation.DeepCopy()
		invalidOperation.ResourceVersion = ""
		invalidOperation.ObjectMeta.Name = "invalid-operation"
		invalidOperation.Spec.WebhookIDs = make([]string, 0)

		err = k8sClient.Create(ctx, invalidOperation)
		require.NoError(t, err)

		invalidOpNamespacedName := types.NamespacedName{Namespace: invalidOperation.ObjectMeta.Namespace, Name: invalidOperation.ObjectMeta.Name}

		err = statusManager.Initialize(ctx, invalidOpNamespacedName)
		require.Error(t, err)

		_, isValErr := err.(*v1alpha1.OperationValidationErr)
		require.True(t, isValErr)
		require.Contains(t, err.Error(), "expected 1 webhook for execution, found: 0")
	})

	t.Run("Test Initialize when generation and observed generation mismatch", func(t *testing.T) {
		err = statusManager.Initialize(ctx, namespacedName)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		require.Equal(t, int64(1), *actualOperation.Status.ObservedGeneration)
		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Phase)

		require.Len(t, actualOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, actualOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Webhooks[0].State)
		require.Equal(t, 0, actualOperation.Status.Webhooks[0].RetriesCount)
		require.Empty(t, actualOperation.Status.Webhooks[0].WebhookPollURL)
		require.Empty(t, actualOperation.Status.Webhooks[0].LastPollTimestamp)

		require.Len(t, actualOperation.Status.Conditions, 2)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[0].Status)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[1].Status)
		require.Empty(t, actualOperation.Status.Conditions[0].Message)
		require.Empty(t, actualOperation.Status.Conditions[1].Message)
	})

	t.Run("Test Initialize when generation and observed generation match shouldn't affect status", func(t *testing.T) {
		err = statusManager.Initialize(ctx, namespacedName)
		require.NoError(t, err)

		err = statusManager.SuccessStatus(ctx, namespacedName)
		require.NoError(t, err)

		err = statusManager.Initialize(ctx, namespacedName)
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
		err = statusManager.InProgressWithPollURL(ctx, namespacedName, mockedPollURL)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Phase)

		require.Len(t, actualOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, actualOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Webhooks[0].State)
		require.Equal(t, 0, actualOperation.Status.Webhooks[0].RetriesCount)
		require.Equal(t, mockedPollURL, actualOperation.Status.Webhooks[0].WebhookPollURL)
		require.Empty(t, actualOperation.Status.Webhooks[0].LastPollTimestamp)

		require.Len(t, actualOperation.Status.Conditions, 2)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[0].Status)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[1].Status)
		require.Empty(t, actualOperation.Status.Conditions[0].Message)
		require.Empty(t, actualOperation.Status.Conditions[1].Message)

	})

	t.Run("Test In Progress with Poll URL And Last Poll Tiimestamp", func(t *testing.T) {
		retryCount := 1
		lastPollTimestamp := time.Now().Format(time.RFC3339Nano)
		err = statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, namespacedName, mockedPollURL, lastPollTimestamp, retryCount)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Phase)

		require.Len(t, actualOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, actualOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateInProgress, actualOperation.Status.Webhooks[0].State)
		require.Equal(t, retryCount, actualOperation.Status.Webhooks[0].RetriesCount)
		require.Equal(t, mockedPollURL, actualOperation.Status.Webhooks[0].WebhookPollURL)
		require.Equal(t, lastPollTimestamp, actualOperation.Status.Webhooks[0].LastPollTimestamp)

		require.Len(t, actualOperation.Status.Conditions, 2)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[0].Status)
		require.Equal(t, corev1.ConditionFalse, actualOperation.Status.Conditions[1].Status)
		require.Empty(t, actualOperation.Status.Conditions[0].Message)
		require.Empty(t, actualOperation.Status.Conditions[1].Message)
	})

	t.Run("Test Success Status", func(t *testing.T) {
		err = statusManager.SuccessStatus(ctx, namespacedName)
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

	t.Run("Test Failed Status", func(t *testing.T) {
		errMsg := "test error"
		err = statusManager.FailedStatus(ctx, namespacedName, errMsg)
		require.NoError(t, err)

		var actualOperation = &v1alpha1.Operation{}
		err = k8sClient.Get(ctx, namespacedName, actualOperation)
		require.NoError(t, err)

		require.Equal(t, v1alpha1.StateFailed, actualOperation.Status.Phase)

		require.Len(t, actualOperation.Status.Webhooks, 1)
		require.Equal(t, webhookID, actualOperation.Status.Webhooks[0].WebhookID)
		require.Equal(t, v1alpha1.StateFailed, actualOperation.Status.Webhooks[0].State)

		require.Len(t, actualOperation.Status.Conditions, 2)
		for _, condition := range actualOperation.Status.Conditions {
			if condition.Type == v1alpha1.ConditionTypeReady {
				require.Equal(t, corev1.ConditionTrue, condition.Status)
				require.Empty(t, condition.Message)
			}

			if condition.Type == v1alpha1.ConditionTypeError {
				require.Equal(t, corev1.ConditionTrue, condition.Status)
				require.Equal(t, errMsg, condition.Message)
			}
		}

	})
}
