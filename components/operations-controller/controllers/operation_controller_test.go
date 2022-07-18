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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	corev1 "k8s.io/api/core/v1"

	collector "github.com/kyma-incubator/compass/components/operations-controller/internal/metrics"

	"github.com/stretchr/testify/assert"

	recerr "github.com/kyma-incubator/compass/components/operations-controller/internal/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers/controllersfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	appGUID                = "f92f1fce-631a-4231-b43a-8f9fccebb22c"
	correlationGUID        = "dea327aa-c173-46d5-9431-497e4057a83a"
	correlationGUID2       = "ce887479-bccf-4bf1-96a9-05b731196b1e"
	anotherCorrelationGUID = "575b8042-8bb1-4ffa-9464-8ec633eae0d3"
	tenantGUID             = "4b7aa2e1-e060-4633-a795-1be0d207c3e2"
	webhookGUID            = "d09731af-bc0a-4abf-9b09-f3c9d25d064b"
	opName                 = "application-f92f1fce-631a-4231-b43a-8f9fccebb22c"
	opNamespace            = "compass-system"
)

var (
	mockedErr         = errors.New("mocked error")
	notFoundErr       = &director.NotFoundError{}
	mockedLocationURL = "https://test-domain.com/operation"
	ctrlRequest       = ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	}
	mockedOperation = &v1alpha1.Operation{
		ObjectMeta: ctrl.ObjectMeta{
			Name:              ctrlRequest.Name,
			Namespace:         ctrlRequest.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestObject: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationTypeCreate,
			WebhookIDs:    []string{webhookGUID},
			CorrelationID: correlationGUID,
		},
	}
	initializedMockedOperation = &v1alpha1.Operation{
		ObjectMeta: ctrl.ObjectMeta{
			Name:              ctrlRequest.Name,
			Namespace:         ctrlRequest.Namespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestObject: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationTypeCreate,
			WebhookIDs:    []string{webhookGUID},
			CorrelationID: correlationGUID,
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{WebhookID: webhookGUID, State: v1alpha1.StateInProgress},
			},
			Conditions: []v1alpha1.Condition{
				{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionFalse},
				{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionFalse},
			},
			Phase:              v1alpha1.StateInProgress,
			InitializedAt:      metav1.Time{Time: time.Now()},
			ObservedGeneration: int64ToInt64Ptr(1),
		},
	}
	originalLogger = *ctrl.Log
)

func TestReconcile_FailureToGetOperationCRDueToNotFoundError_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	notFoundErr := kubeerrors.NewNotFound(schema.GroupResource{}, "test-operation")
	stubLoggerAssertion(t, notFoundErr.Error(),
		fmt.Sprintf("Unable to retrieve %s resource from API server", ctrlRequest.NamespacedName),
		fmt.Sprintf("%s resource was not found in API server", ctrlRequest.NamespacedName))
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(nil, notFoundErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), nil, k8sClient, nil, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertZeroInvocations(t, k8sClient.DeleteCallCount)
}

func TestReconcile_FailureToGetOperationCRDueToGeneralError_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to retrieve")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), nil, k8sClient, nil, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertZeroInvocations(t, k8sClient.DeleteCallCount)
}

func TestReconcile_FailureToInitializeOperationStatusDueToValidationError_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status", "Validation error occurred during operation status initialization")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToInitializeOperationStatusDueToValidationError_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status", "Validation error occurred during operation status initialization")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})
	statusMgrClient.FailedStatusReturns(mockedErr)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, mockedErr.Error())
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToInitializeOperationStatusDueToValidationError_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status", "Validation error occurred during operation status initialization")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})
	statusMgrClient.FailedStatusReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, mockedErr.Error())
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_When_K8sDeleteFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.InitializedAt = metav1.Time{}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)
	k8sClient.DeleteReturns(mockedErr)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertK8sDeleteCalledWithOperation(t, k8sClient, &operation)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_K8sDeleteSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.InitializedAt = metav1.Time{}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)
	k8sClient.DeleteReturns(nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertK8sDeleteCalledWithOperation(t, k8sClient, &operation)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutNotReached_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FetchApplicationReturnsNotFoundErr_And_OperationIsDeleteAndInProgress_ShouldResultSuccessOperationNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, notFoundErr.Error(), fmt.Sprintf("Unable to fetch application with ID %s", appGUID))
	defer func() { ctrl.Log = &originalLogger }()
	operation := *initializedMockedOperation
	operation.Spec.OperationType = v1alpha1.OperationTypeDelete

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)
	k8sClient.DeleteReturns(nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, notFoundErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, &operation)
	assertZeroInvocations(t, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FetchApplicationReturnsNilApplication_And_OperationIsUpdateAndInProgress_ShouldResultFailedOperationNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, notFoundErr.Error(), fmt.Sprintf("Unable to fetch application with ID %s", appGUID))
	defer func() { ctrl.Log = &originalLogger }()
	operation := *initializedMockedOperation
	operation.Spec.OperationType = v1alpha1.OperationTypeUpdate

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, notFoundErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, "Application not found in director")
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FetchApplicationReturnsNilApplication_And_OperationIsNotDeleteNorInProgress_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, notFoundErr.Error(), fmt.Sprintf("Unable to fetch application with ID %s", appGUID))
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, notFoundErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FetchApplicationReturnsNilApplication_And_OperationIsSucceeded_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, notFoundErr.Error(), fmt.Sprintf("Unable to fetch application with ID %s", appGUID))
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Phase = v1alpha1.StateSuccess

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, notFoundErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_When_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", initializedMockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", initializedMockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", initializedMockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, recerr.ErrWebhookTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, recerr.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, recerr.ErrWebhookTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation, recerr.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, recerr.ErrWebhookTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_FatalErrorReturned_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse output template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_FatalErrorReturned_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse output template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, expectedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_FatalErrorReturned_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse output template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, expectedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookStatusGoneErrorReturned_When_OperationTypeIsDelete_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	goneStatusCode := 410
	webhookMode := graphql.WebhookModeAsync
	expectedErr := webhook_client.NewWebhookStatusGoneErr(goneStatusCode)

	stubLoggerAssertion(t, expectedErr.Error(), "gone response status")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	typeDelete := v1alpha1.OperationTypeDelete
	operation.Spec.OperationType = typeDelete
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &webhookMode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{GoneStatusCode: &goneStatusCode}, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{
		DoStub: func(_ context.Context, _ *webhook_client.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		DoStub: func(_ context.Context, _ *webhook_client.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		DoStub: func(_ context.Context, _ *webhook_client.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, initializedMockedOperation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_AsyncWebhookExecutionSucceeds_When_StatusManagerInProgressWithPollURLFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.InProgressWithPollURLReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerInProgressWithPollURLCalled(t, statusMgrClient, initializedMockedOperation, mockedLocationURL)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_AsyncWebhookExecutionSucceeds_And_StatusManagerInProgressWithPollURLSucceeds_ShouldResultRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.InProgressWithPollURLReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.True(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerInProgressWithPollURLCalled(t, statusMgrClient, initializedMockedOperation, mockedLocationURL)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, initializedMockedOperation)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_When_StatusManagerSuccessStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(mockedErr)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, initializedMockedOperation)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(initializedMockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, initializedMockedOperation)
	assertDirectorFetchApplicationCalled(t, directorClient, initializedMockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, initializedMockedOperation)
	assertWebhookDoCalled(t, webhookClient, initializedMockedOperation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_DirectorAndStatusManagerUpdatedSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollIntervalHasNotPassed_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: time.Now().Format(time.RFC3339Nano)}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_FatalErrorReturned_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse status template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_FatalErrorReturned_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse status template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_FatalErrorReturned_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := recerr.NewFatalReconcileError("unable to parse status template")

	stubLoggerAssertion(t, expectedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(nil, expectedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, expectedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_When_StatusManagerInProgressWithPollURLAndLastTimestampFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.InProgressWithPollURLAndLastPollTimestampReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, &operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_When_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.InProgressWithPollURLAndLastPollTimestampReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, &operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, &operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, recerr.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, &operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, recerr.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, recerr.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.Status.InitializedAt = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	webhookTimeout := 5
	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(webhookTimeout), RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{
		PollStub: func(_ context.Context, _ *webhook_client.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, &operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, recerr.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationWithErrorCalled(t, directorClient, &operation, recerr.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_When_DirectorUpdateOperationFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_When_StatusManagerSuccessStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerSuccessStatusCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_When_DirectorUpdateOperationFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, recerr.ErrFailedWebhookStatus.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertStatusManagerFailedStatusCalledWithOperation(t, statusMgrClient, &operation, recerr.ErrFailedWebhookStatus.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationNoWebhookPollURL_And_PollIntervalHasNotPassed_And_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Spec.WebhookIDs = []string{}
	operation.Status.Webhooks = []v1alpha1.Webhook{}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.DoReturns(&web_hook.Response{}, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assert.Equal(t, 1, directorClient.UpdateOperationCallCount())
	assert.Equal(t, 1, statusMgrClient.SuccessStatusCallCount())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsUnknown_And_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	unknownStatus := "UNKNOWN"
	stubLoggerAssertion(t, fmt.Sprintf("unexpected poll status response: %s", unknownStatus), "unknown status code received")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *initializedMockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: initializedMockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, RetryInterval: intToIntPtr(30)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &controllersfakes.FakeWebhookClient{}
	webhookClient.PollReturns(prepareResponseStatus(unknownStatus), nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient, collector.NewCollector())
	res, err := controller.Reconcile(context.Background(), ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithOperation(t, statusMgrClient, &operation)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, &application.Result.Webhooks[0])
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func prepareApplicationOutput(app *graphql.Application, webhooks ...graphql.Webhook) *director.ApplicationOutput {
	return &director.ApplicationOutput{Result: &graphql.ApplicationExt{
		Application: *app,
		Webhooks:    webhooks,
	}}
}

func prepareResponseStatus(status string) *web_hook.ResponseStatus {
	return &web_hook.ResponseStatus{
		Status:                     strToStrPtr(status),
		SuccessStatusIdentifier:    strToStrPtr("SUCCEEDED"),
		InProgressStatusIdentifier: strToStrPtr("IN_PROGRESS"),
		FailedStatusIdentifier:     strToStrPtr("FAILED"),
	}
}

func strToStrPtr(str string) *string {
	return &str
}

func intToIntPtr(i int) *int {
	return &i
}

func int64ToInt64Ptr(i int64) *int64 {
	return &i
}
