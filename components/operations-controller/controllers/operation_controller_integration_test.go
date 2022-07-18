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
	"path/filepath"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	collector "github.com/kyma-incubator/compass/components/operations-controller/internal/metrics"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers/controllersfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"testing"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	eventuallyTimeout = 60 * time.Second
	eventuallyTick    = 1 * time.Second
)

var (
	scheme          = runtime.NewScheme()
	createWebhookID = "730710e5-c09b-4f2b-9b89-db1f86cc1482"
	deleteWebhookID = "a0fb58f4-8ea6-4d85-a60c-44810655ef94"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

/*

 The following tests present semi-integration style tests where a lightweight
 in-memory Kubernetes API Server is deployed in order to test out some of the critical flows
 of the Operation Controller reconciliation loop.

 Unlike the tests in operation_controller_test.go the tests here do not aim to be as exhaustive as possible
 but rather strive to cover main flows and scenarios. Also - similarly to the operation_controller_test.go
 tests, here the Director and Webhook clients have been mocked.

*/

func TestController_Scenarios(t *testing.T) {
	ctx := context.Background()

	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
	require.NoError(t, err)

	directorClient := &controllersfakes.FakeDirectorClient{}
	webhookClient := &controllersfakes.FakeWebhookClient{}

	kubeClient := mgr.GetClient()

	err = kubeClient.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: opNamespace},
	})
	require.NoError(t, err)

	webhookConfig := webhook.DefaultConfig()
	webhookConfig.RequeueInterval = 100 * time.Millisecond
	webhookConfig.TimeoutFactor = 1
	webhookConfig.WebhookTimeout = 10 * time.Second

	controller := controllers.NewOperationReconciler(webhookConfig,
		status.NewManager(kubeClient),
		k8s.NewClient(kubeClient),
		directorClient,
		webhookClient,
		collector.NewCollector())

	err = controller.SetupWithManager(mgr)
	require.NoError(t, err)

	mgrCtx, cancel := context.WithCancel(context.Background())
	go func() {
		err = mgr.Start(mgrCtx)
		require.NoError(t, err)
	}()

	var fetchApplicationInvocations, updateOperationInvocations int
	var doInvocations, pollInvocations int

	t.Run("Successful Async Webhook flow", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

		require.Equal(t, pollInvocations+1, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], pollInvocations)
	})

	t.Run("Successful Async Webhook flow due to gone status", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		goneStatusCode := 410
		expectedErr := webhook_client.NewWebhookStatusGoneErr(goneStatusCode)
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{GoneStatusCode: &goneStatusCode}, expectedErr)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		operation.Spec.OperationType = v1alpha1.OperationTypeDelete
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

		require.Equal(t, pollInvocations, webhookClient.PollCallCount())
	})

	t.Run("Successful Async Webhook flow after polling and requeue a few times", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		pollCallCount := 5
		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			if i == pollInvocations+pollCallCount-1 {
				webhookClient.PollReturnsOnCall(i, prepareResponseStatus("SUCCEEDED"), nil)
				break
			}
			webhookClient.PollReturnsOnCall(i, prepareResponseStatus("IN_PROGRESS"), nil)
		}

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		expectedStatus.Webhooks[0].RetriesCount = pollCallCount - 1

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+pollCallCount+1, directorClient.FetchApplicationCallCount())
		for i := fetchApplicationInvocations; i < pollCallCount+1; i++ {
			assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, i)
		}

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

		require.Equal(t, pollInvocations+pollCallCount, webhookClient.PollCallCount())
		//assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], pollInvocations)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], i)
		}
	})

	t.Run("Successful consecutive Async Webhooks on same Operation CR flow", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}},
			graphql.Webhook{ID: createWebhookID, Mode: &mode, Timeout: intToIntPtr(120)},
			graphql.Webhook{ID: deleteWebhookID, Mode: &mode, Timeout: intToIntPtr(120)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		createOperation := *mockedOperation
		createOperation.Spec.OperationType = v1alpha1.OperationTypeCreate
		createOperation.Spec.WebhookIDs = []string{createWebhookID}
		err := kubeClient.Create(ctx, &createOperation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &createOperation)
			require.NoError(t, err)
		}()

		expectedCreateStatus := expectedSuccessStatus(createWebhookID)
		expectedCreateStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		namespacedName := types.NamespacedName{Namespace: createOperation.ObjectMeta.Namespace, Name: createOperation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedCreateStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.OperationType = v1alpha1.OperationTypeDelete
		latestOperation.Spec.WebhookIDs = []string{deleteWebhookID}
		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		expectedDeleteStatus := expectedSuccessStatus(deleteWebhookID)
		*expectedDeleteStatus.ObservedGeneration += 1
		expectedDeleteStatus.Webhooks[0].WebhookPollURL = mockedLocationURL

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedDeleteStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+4, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, createOperation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, createOperation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+2, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &createOperation, updateOperationInvocations)
		assertDirectorUpdateOperationInvocation(t, directorClient, latestOperation, updateOperationInvocations+1)

		require.Equal(t, doInvocations+2, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &createOperation, &application.Result.Webhooks[0], doInvocations)
		assertWebhookDoInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[1], doInvocations+1)

		require.Equal(t, pollInvocations+2, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &createOperation, &application.Result.Webhooks[0], pollInvocations)
		assertWebhookPollInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[1], pollInvocations+1)
	})

	t.Run("Successful Async Webhook execution after Operation CR has previously resulted in FAILED state", func(t *testing.T) {
		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		expectedErr := errors.NewFatalReconcileError("unable to parse output template")
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}},
			graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(120)},
			graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(120)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturnsOnCall(doInvocations, nil, expectedErr)
		webhookClient.DoReturnsOnCall(doInvocations+1, &web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(operation.Spec.WebhookIDs[0], expectedErr.Error())
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.CorrelationID = anotherCorrelationGUID // we need to updated the spec in order to bump the generation of the resource
		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		expectedStatus = expectedSuccessStatus(operation.Spec.WebhookIDs[0])
		*expectedStatus.ObservedGeneration += 1
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+3, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+2, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)
		assertDirectorUpdateOperationInvocation(t, directorClient, latestOperation, updateOperationInvocations+1)

		require.Equal(t, doInvocations+2, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
		assertWebhookDoInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[0], doInvocations+1)

		require.Equal(t, pollInvocations+1, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[0], pollInvocations)
	})

	t.Run("Failed Async Webhook flow after polling and requeue a few times", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		pollCallCount := 5
		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			if i == pollInvocations+pollCallCount-1 {
				webhookClient.PollReturnsOnCall(i, prepareResponseStatus("FAILED"), nil)
				break
			}
			webhookClient.PollReturnsOnCall(i, prepareResponseStatus("IN_PROGRESS"), nil)
		}

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(webhookGUID, errors.ErrFailedWebhookStatus.Error())
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		expectedStatus.Webhooks[0].RetriesCount = pollCallCount - 1

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+pollCallCount+1, directorClient.FetchApplicationCallCount())
		for i := fetchApplicationInvocations; i < pollCallCount+1; i++ {
			assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, i)
		}

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

		require.Equal(t, pollInvocations+pollCallCount, webhookClient.PollCallCount())
		//assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], pollInvocations)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], i)
		}
	})

	t.Run("Failed Async Webhook flow after timeout", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(4)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(webhookGUID, errors.ErrWebhookTimeoutReached.Error())
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].RetriesCount = operation.Status.Webhooks[0].RetriesCount
			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)
	})

	t.Run("Success after retriggering a failed Async Webhook due to timeout", func(t *testing.T) {
		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(4)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(webhookGUID, errors.ErrWebhookTimeoutReached.Error())
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].RetriesCount = operation.Status.Webhooks[0].RetriesCount
			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		t.Log("Operation timed out... Retrigger should succeed...")

		webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		// Retrigger by updating the correlationID
		err = kubeClient.Get(ctx, namespacedName, &operation)
		require.NoError(t, err)
		operation.Spec.CorrelationID = correlationGUID2

		err = kubeClient.Update(ctx, &operation)
		require.NoError(t, err)

		expectedStatus = expectedSuccessStatus(webhookGUID)
		expectedStatus.Webhooks[0].WebhookPollURL = mockedLocationURL

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

		require.Equal(t, pollInvocations+1, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], pollInvocations)
	})

	t.Run("Successful Sync Webhook flow", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
	})

	t.Run("Successful Sync Webhook flow due to gone status", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		goneStatusCode := 410
		expectedErr := webhook_client.NewWebhookStatusGoneErr(goneStatusCode)
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{GoneStatusCode: &goneStatusCode}, expectedErr)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		operation.Spec.OperationType = v1alpha1.OperationTypeDelete
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
	})

	t.Run("Successful consecutive Sync Webhooks on same Operation CR flow", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}},
			graphql.Webhook{ID: createWebhookID, Mode: &mode, Timeout: intToIntPtr(120)},
			graphql.Webhook{ID: deleteWebhookID, Mode: &mode, Timeout: intToIntPtr(120)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		createOperation := *mockedOperation
		createOperation.Spec.OperationType = v1alpha1.OperationTypeCreate
		createOperation.Spec.WebhookIDs = []string{createWebhookID}
		err := kubeClient.Create(ctx, &createOperation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &createOperation)
			require.NoError(t, err)
		}()

		expectedCreateStatus := expectedSuccessStatus(createWebhookID)
		namespacedName := types.NamespacedName{Namespace: createOperation.ObjectMeta.Namespace, Name: createOperation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedCreateStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.OperationType = v1alpha1.OperationTypeDelete
		latestOperation.Spec.WebhookIDs = []string{deleteWebhookID}
		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		expectedDeleteStatus := expectedSuccessStatus(deleteWebhookID)
		*expectedDeleteStatus.ObservedGeneration += 1

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedDeleteStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, createOperation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, createOperation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+2, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &createOperation, updateOperationInvocations)
		assertDirectorUpdateOperationInvocation(t, directorClient, latestOperation, updateOperationInvocations+1)

		require.Equal(t, doInvocations+2, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &createOperation, &application.Result.Webhooks[0], doInvocations)
		assertWebhookDoInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[1], doInvocations+1)
	})

	t.Run("Successful Sync Webhook execution after Operation CR has previously resulted in FAILED state", func(t *testing.T) {
		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		expectedErr := errors.NewFatalReconcileError("unable to parse output template")
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}},
			graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(120)},
			graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(120)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturnsOnCall(doInvocations, nil, expectedErr)
		webhookClient.DoReturnsOnCall(doInvocations+1, &web_hook.Response{Location: &mockedLocationURL}, nil)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, mockedOperation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(operation.Spec.WebhookIDs[0], expectedErr.Error())
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.CorrelationID = anotherCorrelationGUID
		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		expectedStatus = expectedSuccessStatus(operation.Spec.WebhookIDs[0])
		*expectedStatus.ObservedGeneration += 1

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+2, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)
		assertDirectorUpdateOperationInvocation(t, directorClient, latestOperation, updateOperationInvocations+1)

		require.Equal(t, doInvocations+2, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
		assertWebhookDoInvocation(t, webhookClient, latestOperation, &application.Result.Webhooks[0], doInvocations+1)
	})

	t.Run("Failed Sync Webhook flow after timeout", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(4)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(nil, mockedErr)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedFailedStatus(webhookGUID, errors.ErrWebhookTimeoutReached.Error())
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].RetriesCount = operation.Status.Webhooks[0].RetriesCount
			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)
	})

	t.Run("Operation is deleted after ROT has expired", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		fetchAppInvoked := false
		directorClient.FetchApplicationStub = func(ctx context.Context, s string) (*director.ApplicationOutput, error) {
			if !fetchAppInvoked {
				fetchAppInvoked = true
				return application, nil
			}

			time.Sleep(webhookConfig.WebhookTimeout)
			return nil, mockedErr
		}
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err = kubeClient.Delete(ctx, &operation)
			if err != nil {
				require.Contains(t, err.Error(), kubeerrors.NewNotFound(schema.GroupResource{}, operation.ObjectMeta.Name).Error())
			}
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.CorrelationID = anotherCorrelationGUID

		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			err := kubeClient.Get(ctx, namespacedName, latestOperation)

			return kubeerrors.IsNotFound(err)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
	})

	t.Run("Generation Changed Predicate does not allow requeue as a result of status update", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Status.Conditions[0].Message = "test"
		err = kubeClient.Status().Update(ctx, latestOperation)
		require.NoError(t, err)

		time.Sleep(2 * time.Second) // we have to give the controller some time to potentially execute or not the reconcile loop

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)
	})

	t.Run("Deleting Operation CR does not trigger reconciliation loop", func(t *testing.T) {
		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		operation := *mockedOperation
		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err = kubeClient.Delete(ctx, &operation)
			if err != nil {
				require.Contains(t, err.Error(), kubeerrors.NewNotFound(schema.GroupResource{}, operation.ObjectMeta.Name).Error())
			}
		}()

		expectedStatus := expectedSuccessStatus(webhookGUID)
		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assertStatusEquals(expectedStatus, &operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		stubLoggerNotLoggedAssertion(t, kubeerrors.NewNotFound(schema.GroupResource{}, operation.ObjectMeta.Name).Error(),
			fmt.Sprintf("Unable to retrieve %s resource from API server", namespacedName),
			fmt.Sprintf("%s resource was not found in API server", namespacedName))
		defer func() { ctrl.Log = &originalLogger }()

		err = kubeClient.Delete(ctx, &operation)
		require.NoError(t, err)

		time.Sleep(2 * time.Second) // we have to give the controller some time to potentially execute or not the reconcile loop

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, &application.Result.Webhooks[0], doInvocations)

	})

	cancel() // Stop controller manager before stopping testEnv

	err = testEnv.Stop() // deferring the Stop earlier at the top does not seem to work, this is why the Stop is left here at the bottom
	require.NoError(t, err)
}

func updateInvocationVars(fetchAppCount, updateOpCount, doCount, pollCount *int, directorClient *controllersfakes.FakeDirectorClient, webhookClient *controllersfakes.FakeWebhookClient) {
	*fetchAppCount = directorClient.FetchApplicationCallCount()
	*updateOpCount = directorClient.UpdateOperationCallCount()

	*doCount = webhookClient.DoCallCount()
	*pollCount = webhookClient.PollCallCount()
}

func expectedSuccessStatus(webhookID string) *v1alpha1.OperationStatus {
	var generation int64 = 1
	status := &v1alpha1.OperationStatus{
		Phase: v1alpha1.StateSuccess,
		Conditions: []v1alpha1.Condition{
			{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionTrue},
			{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionFalse},
		},
		Webhooks: []v1alpha1.Webhook{
			{WebhookID: webhookID, State: v1alpha1.StateSuccess},
		},
		ObservedGeneration: &generation,
	}
	return status
}

func expectedFailedStatus(webhookID, errorMsg string) *v1alpha1.OperationStatus {
	var generation int64 = 1
	status := &v1alpha1.OperationStatus{
		Phase: v1alpha1.StateFailed,
		Conditions: []v1alpha1.Condition{
			{Type: v1alpha1.ConditionTypeReady, Status: corev1.ConditionTrue},
			{Type: v1alpha1.ConditionTypeError, Status: corev1.ConditionTrue, Message: errorMsg},
		},
		Webhooks: []v1alpha1.Webhook{
			{WebhookID: webhookID, State: v1alpha1.StateFailed},
		},
		ObservedGeneration: &generation,
	}
	return status
}

func webhookSliceToMap(webhooks []v1alpha1.Webhook) map[string]v1alpha1.Webhook {
	result := make(map[string]v1alpha1.Webhook)
	for i := range webhooks {
		result[webhooks[i].WebhookID] = webhooks[i]
	}
	return result
}

func conditionSliceToMap(conditions []v1alpha1.Condition) map[v1alpha1.ConditionType]v1alpha1.Condition {
	result := make(map[v1alpha1.ConditionType]v1alpha1.Condition)
	for i := range conditions {
		result[conditions[i].Type] = conditions[i]
	}
	return result
}
