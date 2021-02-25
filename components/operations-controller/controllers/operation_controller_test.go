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
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/client/clientfakes"
	ctrl_director "github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director/directorfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook/webhookfakes"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
	"time"
)

const (
	appGUID     = "f92f1fce-631a-4231-b43a-8f9fccebb22c"
	tenantGUID  = "4b7aa2e1-e060-4633-a795-1be0d207c3e2"
	webhookGUID = "d09731af-bc0a-4abf-9b09-f3c9d25d064b"
	opName      = "application-f92f1fce-631a-4231-b43a-8f9fccebb22c"
	opNamespace = "compass-system"
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
)

func TestReconcile_FailureToGetOperationCR_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(nil, logger, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())
}

func TestReconcile_MultipleWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		Spec: v1alpha1.OperationSpec{
			WebhookIDs: []string{"id1", "id2", "id3"},
		},
	}
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	// WHEN:
	controller := NewOperationReconciler(nil, logger, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, fmt.Errorf("expected 1 webhook for execution, found %d", len(operation.Spec.WebhookIDs)), logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())
}

func TestReconcile_ZeroWebhooksRetrievedForExecution_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		Spec: v1alpha1.OperationSpec{
			WebhookIDs: []string{},
		},
	}
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	// WHEN:
	controller := NewOperationReconciler(nil, logger, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, fmt.Errorf("expected 1 webhook for execution, found %d", len(operation.Spec.WebhookIDs)), logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())
}

func TestReconcile_FailureToParseData_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opName,
			Namespace: opNamespace,
		},
		Spec: v1alpha1.OperationSpec{
			WebhookIDs: []string{webhookGUID},
		},
	}
	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	// WHEN:
	controller := NewOperationReconciler(nil, logger, k8sClient, nil, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError.Error(), "unexpected end of JSON input")

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_But_DeleteOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.DeleteReturns(mockedErr)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())

	require.Equal(t, 1, k8sClient.DeleteCallCount())
	_, actualOperation, _ := k8sClient.DeleteArgsForCall(0)
	expectedOperation := *operation
	expectedOperation.Status = prepareDefaultOperationStatus()
	expectedOperation.Status.Phase = ""
	require.Equal(t, &expectedOperation, actualOperation)

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_DeleteOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.DeleteReturns(nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())

	require.Equal(t, 1, k8sClient.DeleteCallCount())
	_, actualOperation, _ := k8sClient.DeleteArgsForCall(0)
	expectedOperation := *operation
	expectedOperation.Status = prepareDefaultOperationStatus()
	expectedOperation.Status.Phase = ""
	require.Equal(t, &expectedOperation, actualOperation)

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_DeleteOperationSucceeds_For_OperationWithDifferentObservedGeneration_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
			Generation:        2,
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			ObservedGeneration: 1,
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:    v1alpha1.ConditionTypeError,
					Status:  corev1.ConditionTrue,
					Message: mockedErr.Error(),
				},
			},
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID:         webhookGUID,
					RetriesCount:      2,
					WebhookPollURL:    mockedLocationURL,
					LastPollTimestamp: time.Now().Format(time.RFC3339Nano),
					State:             v1alpha1.StateFailed,
				},
			},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.DeleteReturns(nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())

	require.Equal(t, 1, k8sClient.DeleteCallCount())
	_, actualOperation, _ := k8sClient.DeleteArgsForCall(0)
	expectedOperation := *operation
	expectedOperation.Status = prepareDefaultOperationStatus()
	expectedOperation.Status.Phase = ""
	expectedOperation.Status.ObservedGeneration = operation.ObjectMeta.Generation
	require.Equal(t, &expectedOperation, actualOperation)

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutReached_And_DeleteOperationSucceeds_For_OperationWithSameObservedGeneration_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{},
			Generation:        2,
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			ObservedGeneration: 2,
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:    v1alpha1.ConditionTypeError,
					Status:  corev1.ConditionTrue,
					Message: mockedErr.Error(),
				},
			},
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID:         webhookGUID,
					RetriesCount:      2,
					WebhookPollURL:    mockedLocationURL,
					LastPollTimestamp: time.Now().Format(time.RFC3339Nano),
					State:             v1alpha1.StateFailed,
				},
			},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.DeleteReturns(nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())

	require.Equal(t, 1, k8sClient.DeleteCallCount())
	_, actualOperation, _ := k8sClient.DeleteArgsForCall(0)
	require.Equal(t, operation, actualOperation)

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_FailureToFetchApplication_And_ReconciliationTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		} else {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true, Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		} else {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationIsReady_And_ApplicationHasNoError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Ready: true}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationHasError_But_UpdateOperationFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ApplicationHasError_And_UpdateOperationSucceeds_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{Error: strToStrPtr(mockedErr.Error())}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_WebhookIsMissing_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:  appGUID,
			RequestData: fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, logger.RecordedError, fmt.Errorf("missing webhook with ID: %s", webhookGUID))

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrReconciliationTimeoutReached.Error(),
	}
	require.Equal(t, expectedRequest, actualRequest)
}

func TestReconcile_ReconciliationTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrReconciliationTimeoutReached.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrReconciliationTimeoutReached.Error(),
	}
	require.Equal(t, expectedRequest, actualRequest)
}

func TestReconcile_ReconciliationTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, nil)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrReconciliationTimeoutReached.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrReconciliationTimeoutReached.Error(),
	}
	require.Equal(t, expectedRequest, actualRequest)
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedRequest, actualRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateInProgress
	expectedOp.Status.Webhooks[0].WebhookPollURL = mockedLocationURL
	expectedOp.Status.Phase = v1alpha1.StateInProgress
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_And_UpdateOperationStatusSucceeds_ShouldResultRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	mode := graphql.WebhookModeAsync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.True(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateInProgress
	expectedOp.Status.Webhooks[0].WebhookPollURL = mockedLocationURL
	expectedOp.Status.Phase = v1alpha1.StateInProgress
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_And_UpdateDirectorAndUpdateOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 1, webhookClient.DoCallCount())
	_, actualWebhookRequest := webhookClient.DoArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.Request{
		Webhook: application.Result.Webhooks[0],
		Data:    expectedRequestData,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)

	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_But_TimeLayoutParsingFails_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL, LastPollTimestamp: "abc"}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Contains(t, logger.RecordedError.Error(), "cannot parse")

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollIntervalHasNotPassed_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL, LastPollTimestamp: time.Now().Format(time.RFC3339Nano)}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, RetryInterval: intToIntPtr(60)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 0, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(nil, mockedErr)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 0, webhookClient.DoCallCount())

	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedRequest, actualRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = mockedErr.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         mockedErr.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_But_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrWebhookTimeoutReached.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrWebhookTimeoutReached.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrWebhookTimeoutReached.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.PollRequest) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrWebhookTimeoutReached.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrWebhookTimeoutReached.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateSuccess
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateSuccess
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrFailedWebhookStatus.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrFailedWebhookStatus.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrFailedWebhookStatus.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateOperationReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 1, k8sClient.UpdateStatusCallCount())
	expectedOp := *operation
	expectedOp.Status = prepareDefaultOperationStatus()
	expectedOp.Status.Webhooks[0].WebhookID = webhookGUID
	expectedOp.Status.Webhooks[0].State = v1alpha1.StateFailed
	expectedOp.Status.Webhooks[0].WebhookPollURL = operation.Status.Webhooks[0].WebhookPollURL
	for i, condition := range expectedOp.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeError {
			expectedOp.Status.Conditions[i].Status = corev1.ConditionTrue
			expectedOp.Status.Conditions[i].Message = ErrFailedWebhookStatus.Error()
		}
	}
	expectedOp.Status.Phase = v1alpha1.StateFailed
	_, actualOp, _ := k8sClient.UpdateStatusArgsForCall(0)
	require.Equal(t, &expectedOp, actualOp)

	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 1, directorClient.UpdateOperationCallCount())
	_, actualDirectorRequest := directorClient.UpdateOperationArgsForCall(0)
	expectedDirectorRequest := &ctrl_director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
		Error:         ErrFailedWebhookStatus.Error(),
	}
	require.Equal(t, expectedDirectorRequest, actualDirectorRequest)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsUnknown_And_ShouldResultNoRequeueError(t *testing.T) {
	// GIVEN:
	logger := &mockedLogger{}

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              opName,
			Namespace:         opNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.OperationSpec{
			ResourceID:    appGUID,
			RequestData:   fmt.Sprintf(`{"TenantID":"%s"}`, tenantGUID),
			ResourceType:  "application",
			OperationType: v1alpha1.OperationType(graphql.OperationTypeDelete),
			WebhookIDs:    []string{webhookGUID},
		},
		Status: v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{{WebhookID: webhookGUID, WebhookPollURL: mockedLocationURL}},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	status := "UNKNOWN"
	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus(status), nil)

	// WHEN:
	controller := NewOperationReconciler(webhook.DefaultConfig(), logger, k8sClient, directorClient, webhookClient)
	res, err := controller.Reconcile(ctrlRequest)

	// THEN:
	// GENERAL ASSERTIONS:
	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, fmt.Errorf("unexpected poll status response: %s", status), err)

	// SPECIFIC CLIENT ASSERTIONS:
	require.Equal(t, 1, k8sClient.GetCallCount())
	_, namespacedName := k8sClient.GetArgsForCall(0)
	require.Equal(t, ctrlRequest.NamespacedName, namespacedName)

	require.Equal(t, 0, k8sClient.UpdateStatusCallCount())

	require.Equal(t, 1, directorClient.FetchApplicationCallCount())
	ctx, resourceID := directorClient.FetchApplicationArgsForCall(0)
	require.Equal(t, appGUID, resourceID)
	require.Equal(t, tenantGUID, ctx.Value(tenant.ContextKey))

	require.Equal(t, 0, directorClient.UpdateOperationCallCount())
	require.Equal(t, 0, k8sClient.DeleteCallCount())

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
	_, actualWebhookRequest := webhookClient.PollArgsForCall(0)
	expectedRequestData, err := parseRequestData(operation)
	require.NoError(t, err)
	expectedWebhookRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: application.Result.Webhooks[0],
			Data:    expectedRequestData,
		},
		PollURL: mockedLocationURL,
	}
	require.Equal(t, expectedWebhookRequest, actualWebhookRequest)
}

func prepareApplicationOutput(app *graphql.Application, webhooks ...graphql.Webhook) *director.ApplicationOutput {
	return &director.ApplicationOutput{Result: &graphql.ApplicationExt{
		Application: *app,
		Webhooks:    webhooks,
	}}
}

func prepareDefaultOperationStatus() v1alpha1.OperationStatus {
	return v1alpha1.OperationStatus{
		Webhooks: []v1alpha1.Webhook{
			{
				WebhookID: webhookGUID,
				State:     v1alpha1.StateInProgress,
			},
		},
		Conditions: []v1alpha1.Condition{
			{
				Type:   v1alpha1.ConditionTypeReady,
				Status: corev1.ConditionFalse,
			},
			{
				Type:   v1alpha1.ConditionTypeError,
				Status: corev1.ConditionFalse,
			},
		},
		Phase: v1alpha1.StateInProgress,
	}
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
