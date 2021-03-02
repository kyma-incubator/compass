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

package controllers_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers/controllersfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func stubLoggerAssertion(t *testing.T, errExpectation string, msgExpectations ...string) {
	ctrl.Log = log.NewDelegatingLogger(&mockedLogger{
		AssertErrorExpectations: func(err error, msg string) {
			require.Contains(t, err.Error(), errExpectation)

			matchedMsg := false
			for _, msgExpectation := range msgExpectations {
				if strings.Contains(msg, msgExpectation) {
					matchedMsg = true
				}
			}
			require.True(t, matchedMsg)
		},
	})
}

func stubLoggerNotLoggedAssertion(t *testing.T, errExpectation string, msgExpectations ...string) {
	ctrl.Log = log.NewDelegatingLogger(&mockedLogger{
		AssertErrorExpectations: func(err error, msg string) {
			require.NotContains(t, err.Error(), errExpectation)

			matchedMsg := false
			for _, msgExpectation := range msgExpectations {
				if strings.Contains(msg, msgExpectation) {
					matchedMsg = true
				}
			}
			require.False(t, matchedMsg)
		},
	})
}

func assertZeroInvocations(t *testing.T, callCountFunc ...func() int) {
	for _, callCount := range callCountFunc {
		require.Equal(t, callCount(), 0)
	}
}

func assertK8sGetCalledWithName(t *testing.T, k8sClient *controllersfakes.FakeKubernetesClient, expectedName types.NamespacedName) {
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
}

func assertK8sDeleteCalledWithOperation(t *testing.T, k8sClient *controllersfakes.FakeKubernetesClient, expectedOperation *v1alpha1.Operation) {
	require.Equal(t, 1, k8sClient.DeleteCallCount())
	_, actualOperation, _ := k8sClient.DeleteArgsForCall(0)
	require.Equal(t, expectedOperation, actualOperation)
}

func assertStatusManagerInitializeCalledWithName(t *testing.T, statusManagerClient *controllersfakes.FakeStatusManager, expectedName types.NamespacedName) {
	require.Equal(t, 1, statusManagerClient.InitializeCallCount())
	_, namespacedName := statusManagerClient.InitializeArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
}

func assertStatusManagerSuccessStatusCalledWithName(t *testing.T, statusManagerClient *controllersfakes.FakeStatusManager, expectedName types.NamespacedName) {
	require.Equal(t, 1, statusManagerClient.SuccessStatusCallCount())
	_, namespacedName := statusManagerClient.SuccessStatusArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
}

func assertStatusManagerInProgressWithPollURLCalled(t *testing.T, statusManagerClient *controllersfakes.FakeStatusManager, expectedName types.NamespacedName, expectedPollURL string) {
	require.Equal(t, 1, statusManagerClient.InProgressWithPollURLCallCount())
	_, namespacedName, pollURL := statusManagerClient.InProgressWithPollURLArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
	require.Equal(t, expectedPollURL, pollURL)
}

func assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t *testing.T, statusManagerClient *controllersfakes.FakeStatusManager, expectedName types.NamespacedName, operation v1alpha1.Operation, expectedPollURL string) {
	require.Equal(t, 1, statusManagerClient.InProgressWithPollURLAndLastPollTimestampCallCount())
	_, namespacedName, pollURL, lastPollTimestamp, retryCount := statusManagerClient.InProgressWithPollURLAndLastPollTimestampArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
	require.Equal(t, expectedPollURL, pollURL)

	timestamp, err := time.Parse(time.RFC3339Nano, lastPollTimestamp)
	require.NoError(t, err)
	require.True(t, timestamp.After(operation.CreationTimestamp.Time))
	require.Equal(t, operation.Status.Webhooks[0].RetriesCount+1, retryCount)
}

func assertStatusManagerFailedStatusCalledWithName(t *testing.T, statusManagerClient *controllersfakes.FakeStatusManager, expectedName types.NamespacedName, expectedErrorMsg string) {
	require.Equal(t, 1, statusManagerClient.FailedStatusCallCount())
	_, namespacedName, errorMsg := statusManagerClient.FailedStatusArgsForCall(0)
	require.Equal(t, expectedName, namespacedName)
	require.Contains(t, errorMsg, expectedErrorMsg)
}

func assertDirectorUpdateOperationCalled(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, operation *v1alpha1.Operation) {
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, operation, "")
}

func assertDirectorUpdateOperationInvocation(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, operation *v1alpha1.Operation, invocation int) {
	assertDirectorUpdateOperationWithErrorInvocation(t, directorClient, operation, "", invocation)
}

func assertDirectorUpdateOperationWithErrorCalled(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, operation *v1alpha1.Operation, errMsg string) {
	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	assertDirectorUpdateOperationWithErrorInvocation(t, directorClient, operation, errMsg, 0)
}

func assertDirectorUpdateOperationWithErrorInvocation(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, operation *v1alpha1.Operation, errMsg string, invocation int) {
	_, actualRequest := directorClient.UpdateOperationArgsForCall(invocation)
	require.Equal(t, graphql.OperationType(operation.Spec.OperationType), actualRequest.OperationType)
	require.Equal(t, resource.Type(operation.Spec.ResourceType), actualRequest.ResourceType)
	require.Equal(t, operation.Spec.ResourceID, actualRequest.ResourceID)
	require.Contains(t, actualRequest.Error, errMsg)
}

func assertDirectorFetchApplicationCalled(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, expectedResourceID, expectedTenantID string) {
	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	assertDirectorFetchApplicationInvocation(t, directorClient, expectedResourceID, expectedTenantID, 0)
}

func assertDirectorFetchApplicationInvocation(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, expectedResourceID, expectedTenantID string, invocation int) {
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(invocation)
	require.Equal(t, expectedResourceID, resourceID)
	require.Equal(t, expectedTenantID, ctx.Value(tenant.ContextKey))
}

func assertWebhookDoCalled(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput) {
	require.Equal(t, 1, webhookClient.DoCallCount())
	assertWebhookDoInvocation(t, webhookClient, operation, application, 0)
}

func assertWebhookDoInvocation(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput, invocation int) {
	_, actualRequest := webhookClient.DoArgsForCall(invocation)
	expectedRequestObject, err := operation.RequestObject()
	require.NoError(t, err)
	expectedRequest := webhook.NewRequest(application.Result.Webhooks[0], expectedRequestObject, operation.Spec.CorrelationID)
	require.Equal(t, expectedRequest, actualRequest)
}

func assertWebhookPollCalled(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput) {
	require.Equal(t, 1, webhookClient.PollCallCount())
	assertWebhookPollInvocation(t, webhookClient, operation, application, 0)
}

func assertWebhookPollInvocation(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput, invocation int) {
	_, actualRequest := webhookClient.PollArgsForCall(invocation)
	expectedRequestObject, err := operation.RequestObject()
	require.NoError(t, err)
	expectedRequest := webhook.NewPollRequest(application.Result.Webhooks[0], expectedRequestObject, operation.Spec.CorrelationID, mockedLocationURL)
	require.Equal(t, expectedRequest, actualRequest)
}
