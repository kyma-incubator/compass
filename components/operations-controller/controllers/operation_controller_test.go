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

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers/controllersfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	appGUID         = "f92f1fce-631a-4231-b43a-8f9fccebb22c"
	correlationGUID = "dea327aa-c173-46d5-9431-497e4057a83a"
	tenantGUID      = "4b7aa2e1-e060-4633-a795-1be0d207c3e2"
	webhookGUID     = "d09731af-bc0a-4abf-9b09-f3c9d25d064b"
	opName          = "application-f92f1fce-631a-4231-b43a-8f9fccebb22c"
	opNamespace     = "compass-system"
)

var (
	mockedErr         = errors.New("mocked error")
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
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
			CorrelationID: correlationGUID,
		},
	}
	originalLogger = *ctrl.Log
)

func TestReconcile_FailureToGetOperationCRDueToNotFoundError_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, "not found", "Unable to retrieve")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(nil, kubeerrors.NewNotFound(schema.GroupResource{}, "test-operation"))

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, nil, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

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
	controller := controllers.NewOperationReconciler(nil, nil, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

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
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToInitializeOperationStatusDueToValidationError_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})
	statusMgrClient.FailedStatusReturns(mockedErr)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToInitializeOperationStatusDueToValidationError_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Failed to initialize operation status")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(&v1alpha1.OperationValidationErr{Description: mockedErr.Error()})
	statusMgrClient.FailedStatusReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, mockedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToParseRequestObject_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "unexpected end of JSON input"
	stubLoggerAssertion(t, expectedErr, "Unable to parse request object")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Spec.RequestObject = ""

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(nil, statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.FetchApplicationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_When_K8sDeleteFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.CreationTimestamp = metav1.Time{}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)
	k8sClient.DeleteReturns(mockedErr)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertK8sDeleteCalledWithOperation(t, k8sClient, &operation)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_K8sDeleteSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.CreationTimestamp = metav1.Time{}

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(&operation, nil)
	k8sClient.DeleteReturns(nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertK8sDeleteCalledWithOperation(t, k8sClient, &operation)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutNotReached_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to fetch application")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_When_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.SuccessStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", mockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", mockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_WebhookIsMissing_When_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := fmt.Errorf("missing webhook with ID: %s", mockedOperation.Spec.WebhookIDs[0])
	stubLoggerAssertion(t, expectedErr.Error(), "Unable to retrieve webhook")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, expectedErr.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, controllers.ErrReconciliationTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrReconciliationTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, controllers.ErrReconciliationTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_ReconciliationTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

	statusMgrClient := &controllersfakes.FakeStatusManager{}
	statusMgrClient.InitializeReturns(nil)
	statusMgrClient.FailedStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &controllersfakes.FakeDirectorClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrReconciliationTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, controllers.ErrReconciliationTimeoutReached.Error())
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

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
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

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
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook is to be executed

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
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}
	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, mockedOperation, mockedErr.Error())
	assertWebhookDoCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_AsyncWebhookExecutionSucceeds_When_StatusManagerInProgressWithPollURLFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLCalled(t, statusMgrClient, ctrlRequest.NamespacedName, mockedLocationURL)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_AsyncWebhookExecutionSucceeds_And_StatusManagerInProgressWithPollURLSucceeds_ShouldResultRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.True(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLCalled(t, statusMgrClient, ctrlRequest.NamespacedName, mockedLocationURL)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, mockedOperation)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_When_StatusManagerSuccessStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, mockedOperation)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationWithoutWebhookPollURL_And_SyncWebhookExecutionSucceeds_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	k8sClient := &controllersfakes.FakeKubernetesClient{}
	k8sClient.GetReturns(mockedOperation, nil)

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, mockedOperation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, mockedOperation)
	assertWebhookDoCalled(t, webhookClient, mockedOperation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_TimeLayoutParsingFails_When_DirectorAndStatusManagerUpdatedSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, expectedErr)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, expectedErr)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollIntervalHasNotPassed_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	expectedErr := "cannot parse"
	stubLoggerAssertion(t, expectedErr, "Unable to calculate next poll time")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL, LastPollTimestamp: time.Now().Format(time.RFC3339Nano)}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount, webhookClient.PollCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	stubLoggerAssertion(t, mockedErr.Error(), "Unable to execute Webhook Poll request")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, mockedErr.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, mockedErr.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_When_StatusManagerInProgressWithPollURLAndLastTimestampFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, ctrlRequest.NamespacedName, operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_When_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}

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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, ctrlRequest.NamespacedName, operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_When_DirectorUpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, ctrlRequest.NamespacedName, operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, controllers.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, ctrlRequest.NamespacedName, operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, controllers.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
	operation.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()} // This is necessary because later in the test we rely on ROT to not be reached by the time the Webhook Poll is to be executed

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
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerInProgressWithPollURLAndLastTimestampCalled(t, statusMgrClient, ctrlRequest.NamespacedName, operation, operation.Status.Webhooks[0].WebhookPollURL)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrWebhookTimeoutReached.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalledWithError(t, directorClient, &operation, controllers.ErrWebhookTimeoutReached.Error())
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.SuccessStatusCallCount, webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_When_DirectorUpdateOperationFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_When_StatusManagerSuccessStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerSuccessStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_When_DirectorUpdateOperationFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_When_StatusManagerFailedStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrFailedWebhookStatus.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_And_DirectorAndStatusManagerUpdateSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertStatusManagerFailedStatusCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName, controllers.ErrFailedWebhookStatus.Error())
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertDirectorUpdateOperationCalled(t, directorClient, &operation)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount,
		webhookClient.DoCallCount)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsUnknown_And_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	unknownStatus := "UNKNOWN"
	stubLoggerAssertion(t, fmt.Sprintf("unexpected poll status response: %s", unknownStatus), "unknown status code received")
	defer func() { ctrl.Log = &originalLogger }()

	operation := *mockedOperation
	operation.Status.Webhooks = []v1alpha1.Webhook{{WebhookID: mockedOperation.Spec.WebhookIDs[0], WebhookPollURL: mockedLocationURL}}
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
	controller := controllers.NewOperationReconciler(webhook.DefaultConfig(), statusMgrClient, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	assertK8sGetCalledWithName(t, k8sClient, ctrlRequest.NamespacedName)
	assertStatusManagerInitializeCalledWithName(t, statusMgrClient, ctrlRequest.NamespacedName)
	assertDirectorFetchApplicationCalled(t, directorClient, operation.Spec.ResourceID, tenantGUID)
	assertWebhookPollCalled(t, webhookClient, &operation, application)
	assertZeroInvocations(t, k8sClient.DeleteCallCount, directorClient.UpdateOperationCallCount, statusMgrClient.InProgressWithPollURLCallCount,
		statusMgrClient.InProgressWithPollURLAndLastPollTimestampCallCount, statusMgrClient.SuccessStatusCallCount, statusMgrClient.FailedStatusCallCount,
		webhookClient.DoCallCount)
}

func stubLoggerAssertion(t *testing.T, errExpectation, msgExpectation string) {
	ctrl.Log = log.NewDelegatingLogger(&mockedLogger{
		AssertErrorExpectations: func(err error, msg string) {
			require.Contains(t, err.Error(), errExpectation)
			if len(msgExpectation) != 0 {
				require.Contains(t, msg, msgExpectation)
			}
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
	assertDirectorUpdateOperationCalledWithError(t, directorClient, operation, "")
}

func assertDirectorUpdateOperationCalledWithError(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, operation *v1alpha1.Operation, errMsg string) {
	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualRequest := directorClient.UpdateOperationArgsForCall(0)
	require.Equal(t, graphql.OperationType(operation.Spec.OperationType), actualRequest.OperationType)
	require.Equal(t, resource.Type(operation.Spec.ResourceType), actualRequest.ResourceType)
	require.Equal(t, operation.Spec.ResourceID, actualRequest.ResourceID)
	require.Contains(t, actualRequest.Error, errMsg)
}

func assertDirectorFetchApplicationCalled(t *testing.T, directorClient *controllersfakes.FakeDirectorClient, expectedResourceID, expectedTenantID string) {
	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, expectedResourceID, resourceID)
	require.Equal(t, expectedTenantID, ctx.Value(tenant.ContextKey))
}

func assertWebhookDoCalled(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput) {
	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualRequest := webhookClient.DoArgsForCall(0)
	expectedRequestObject, err := operation.RequestObject()
	require.NoError(t, err)
	expectedRequest := webhook.NewRequest(application.Result.Webhooks[0], expectedRequestObject, operation.Spec.CorrelationID)
	require.Equal(t, expectedRequest, actualRequest)
}

func assertWebhookPollCalled(t *testing.T, webhookClient *controllersfakes.FakeWebhookClient, operation *v1alpha1.Operation, application *director.ApplicationOutput) {
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualRequest := webhookClient.PollArgsForCall(0)
	expectedRequestObject, err := operation.RequestObject()
	require.NoError(t, err)
	expectedRequest := webhook.NewPollRequest(application.Result.Webhooks[0], expectedRequestObject, operation.Spec.CorrelationID, mockedLocationURL)
	require.Equal(t, expectedRequest, actualRequest)
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
