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
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/client/clientfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director/directorfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook/webhookfakes"
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
	logger := &mockedLogger{}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, nil, nil, nil, logger)
	res, err := controller.Reconcile(ctrl.Request{})

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)
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
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)
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
	require.Equal(t, mockedErr, logger.RecordedError)
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
	require.Equal(t, mockedErr, logger.RecordedError)
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
	require.Equal(t, mockedErr, err)
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
	require.Equal(t, mockedErr, err)
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
	require.Equal(t, mockedErr, err)
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
			WebhookIDs:  []string{webhookGUID},
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
	require.Equal(t, logger.RecordedError, fmt.Errorf("missing webhook with ID: %s", webhookGUID))
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, nil, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(0)})

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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(nil, mockedErr)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_WebhookExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	webhookTimeout := 5
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Timeout: intToIntPtr(webhookTimeout)})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		DoStub: func(_ context.Context, _ *webhook.Request) (*web_hook.Response, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_AsyncWebhookExecutionSucceeds_And_UpdateOperationStatusSucceeds_ShouldResultRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.True(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(mockedErr)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasNoWebhookPollURL_And_SyncWebhookExecutionSucceeds_And_UpdateDirectorAndUpdateOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
		},
	}

	k8sClient := &clientfakes.FakeClient{}
	k8sClient.GetReturns(operation, nil)
	k8sClient.UpdateStatusReturns(nil)

	mode := graphql.WebhookModeSync
	application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

	directorClient := &directorfakes.FakeClient{}
	directorClient.FetchApplicationReturns(application, nil)
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 1, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_But_TimeLayoutParsingFails_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Contains(t, logger.RecordedError.Error(), "cannot parse")

	require.Equal(t, 0, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollIntervalHasNotPassed_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionFails_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return nil, mockedErr
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_But_WebhookTimeoutNotReached_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsInProgress_And_WebhookTimeoutReached_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{
		PollStub: func(_ context.Context, _ *webhook.Request) (*web_hook.ResponseStatus, error) {
			time.Sleep(time.Duration(webhookTimeout) * time.Second)
			return prepareResponseStatus("IN_PROGRESS"), nil
		},
	}

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsSucceeded_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateDirectorStatusFails_ShouldResultRequeueAfterNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(mockedErr)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.NotZero(t, res.RequeueAfter)

	require.NoError(t, err)
	require.Equal(t, mockedErr, logger.RecordedError)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_But_UpdateOperationStatusFails_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, mockedErr, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsFailed_And_UpdateDirectorAndOperationStatusSucceed_ShouldResultNoRequeueNoError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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
	directorClient.UpdateStatusReturns(nil)

	webhookClient := &webhookfakes.FakeClient{}
	webhookClient.PollReturns(prepareResponseStatus("FAILED"), nil)

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.NoError(t, err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
}

func TestReconcile_OperationHasWebhookPollURL_And_PollExecutionSucceeds_And_StatusIsUnknown_And_ShouldResultNoRequeueError(t *testing.T) {
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
			WebhookIDs:  []string{webhookGUID},
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

	controller := NewOperationReconciler(k8sClient, webhook.DefaultConfig(), directorClient, webhookClient, logger)
	res, err := controller.Reconcile(ctrlRequest)

	require.False(t, res.Requeue)
	require.Zero(t, res.RequeueAfter)

	require.Error(t, err)
	require.Equal(t, fmt.Errorf("unexpected poll status response: %s", status), err)

	require.Equal(t, 0, webhookClient.DoCallCount())
	require.Equal(t, 1, webhookClient.PollCallCount())
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
