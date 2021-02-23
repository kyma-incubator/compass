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
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/client/clientfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director/directorfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
	"time"
)

const (
	appGUID     = "f92f1fce-631a-4231-b43a-8f9fccebb22c"
	opName      = "application-f92f1fce-631a-4231-b43a-8f9fccebb22c"
	opNamespace = "compass-system"
)

var (
	logger    = ctrl.Log.WithName("controllers").WithName("Operation")
	mockedErr = errors.New("mocked error")
)

func TestReconcile_FailureToGetOperationCR_ShouldResultNoRequeueNoError(t *testing.T) {
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, nil, nil, nil, logger)

	res, err := controller.Reconcile(ctrl.Request{})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)
	require.NoError(t, err)
}

func TestReconcile_FailureToParseData_ShouldResultNoRequeueNoError(t *testing.T) {
	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opName,
			Namespace: opNamespace,
		},
	}
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	controller := NewOperationReconciler(k8sClient, nil, nil, nil, logger)

	res, err := controller.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)
	require.NoError(t, err)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_But_DeleteOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.DeleteReturns(mockedErr)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)

	res, err := controller.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)
	require.Error(t, err)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_DeleteOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)

	res, err := controller.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)
	require.NoError(t, err)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)

	res, err := controller.Reconcile(ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	})

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)
	require.NoError(t, err)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_MultipleWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_ZeroWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_ReconciliationTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_And_UpdateOperationStatusSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_And_UpdateDirectorAndUpdateOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_But_TimeLayoutParsingFails_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollIntervalHasNotPassed_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_But_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsUnknown_And_ShouldResultNoRequeueError(t *testing.T) {
}
