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
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/client/clientfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director/directorfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
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
	webhookID   = "d09731af-bc0a-4abf-9b09-f3c9d25d064b"
)

var (
	mockedErr   = errors.New("mocked error")
	ctrlRequest = ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: opNamespace,
			Name:      opName,
		},
	}
)

func TestReconcile_FailureToGetOperationCR_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, nil, nil, nil, logger)
	res, err := controller.Reconcile(ctrl.Request{})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, mockedErr)
}

func TestReconcile_FailureToParseData_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opName,
			Namespace: opNamespace,
		},
	}
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	controller := NewOperationReconciler(k8sClient, nil, nil, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError.Error(), "unexpected end of JSON input")
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_But_DeleteOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	logger := &mockedLogger{}

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
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, err, mockedErr)
	require.Equal(t, logger.RecordedError, mockedErr)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_DeleteOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.DeleteReturns(nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, mockedErr)
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	logger := &mockedLogger{}

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
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, mockedErr)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, err, mockedErr)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, err, mockedErr)
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
}

func TestReconcile_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, err, mockedErr)
}

func TestReconcile_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

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
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
}

func TestReconcile_MultipleWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{"id1", "id2", "id3"},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError.Error(), "multiple webhooks per operation are not supported")
}

func TestReconcile_ZeroWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError.Error(), "no webhooks found for operation")
}

func TestReconcile_WebhookIsMissing_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{webhookID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, fmt.Errorf("missing webhook with ID: %s", webhookID))
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{webhookID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, mockedErr)
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{webhookID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, err, mockedErr)
}

func TestReconcile_ReconciliationTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: "{}",
			WebhookIDs:  []string{webhookID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
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

func prepareApplicationOutput(app *graphql.Application, webhooks ...graphql.Webhook) *director.ApplicationOutput {
	return &director.ApplicationOutput{Result: &graphql.ApplicationExt{
		Application: *app,
		Webhooks:    webhooks,
	}}
}

func strToStrPtr(str string) *string {
	return &str
}

func intToIntPtr(i int) *int {
	return &i
}
