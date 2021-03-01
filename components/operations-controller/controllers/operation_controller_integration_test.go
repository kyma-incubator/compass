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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers"
	"github.com/kyma-incubator/compass/components/operations-controller/controllers/controllersfakes"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s/status"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"testing"
)

const (
	eventuallyTimeout = 60 * time.Second
	eventuallyTick    = 1 * time.Second
)

var scheme = runtime.NewScheme()

func init() {
	err := clientgoscheme.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}

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
	defer func() {
	}()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
	require.NoError(t, err)

	directorClient := &controllersfakes.FakeDirectorClient{}
	webhookClient := &controllersfakes.FakeWebhookClient{}

	kubeClient := mgr.GetClient()

	webhookConfig := webhook.DefaultConfig()
	webhookConfig.RequeueInterval = 100 * time.Millisecond
	webhookConfig.TimeoutFactor = 1
	webhookConfig.WebhookTimeout = 10 * time.Second

	controller := controllers.NewOperationReconciler(webhookConfig,
		status.NewManager(kubeClient),
		k8s.NewClient(kubeClient),
		directorClient,
		webhookClient)

	err = controller.SetupWithManager(mgr)
	require.NoError(t, err)

	go func() {
		err = mgr.Start(ctrl.SetupSignalHandler())
		require.NoError(t, err)
	}()

	var fetchApplicationInvocations, updateOperationInvocations int
	var doInvocations, pollInvocations int

	t.Run("Successful Async Webhook flow", func(t *testing.T) {
		operation := *mockedOperation
		operation.Name = "async-success-operation"

		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("SUCCEEDED"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID:      operation.Spec.WebhookIDs[0],
					WebhookPollURL: mockedLocationURL,
					State:          v1alpha1.StateSuccess,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   v1alpha1.ConditionTypeError,
					Status: corev1.ConditionFalse,
				},
			},
			Phase:              v1alpha1.StateSuccess,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assert.Equal(t, expectedStatus, operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations+1)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

		require.Equal(t, pollInvocations+1, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &operation, application, pollInvocations)

	})

	t.Run("Successful Async Webhook flow after polling and requeue a few times", func(t *testing.T) {
		operation := *mockedOperation
		operation.Name = "async-polling-success-operation"

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

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID:      operation.Spec.WebhookIDs[0],
					WebhookPollURL: mockedLocationURL,
					State:          v1alpha1.StateSuccess,
					RetriesCount:   pollCallCount - 1,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   v1alpha1.ConditionTypeError,
					Status: corev1.ConditionFalse,
				},
			},
			Phase:              v1alpha1.StateSuccess,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			return assert.Equal(t, expectedStatus, operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+pollCallCount+1, directorClient.FetchApplicationCallCount())
		for i := fetchApplicationInvocations; i < pollCallCount+1; i++ {
			assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, i)
		}

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

		require.Equal(t, pollInvocations+pollCallCount, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &operation, application, pollInvocations)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			assertWebhookPollInvocation(t, webhookClient, &operation, application, i)
		}

	})

	t.Run("Failed Async Webhook flow after polling and requeue a few times", func(t *testing.T) {
		operation := *mockedOperation
		operation.Name = "async-polling-failed-operation"

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

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID:      operation.Spec.WebhookIDs[0],
					WebhookPollURL: mockedLocationURL,
					State:          v1alpha1.StateFailed,
					RetriesCount:   pollCallCount - 1,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:    v1alpha1.ConditionTypeError,
					Status:  corev1.ConditionTrue,
					Message: controllers.ErrFailedWebhookStatus.Error(),
				},
			},
			Phase:              v1alpha1.StateFailed,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			if len(operation.Status.Webhooks) == 0 {
				return false
			}

			expectedStatus.Webhooks[0].LastPollTimestamp = operation.Status.Webhooks[0].LastPollTimestamp
			return assert.Equal(t, expectedStatus, operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+pollCallCount+1, directorClient.FetchApplicationCallCount())
		for i := fetchApplicationInvocations; i < pollCallCount+1; i++ {
			assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, i)
		}

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

		require.Equal(t, pollInvocations+pollCallCount, webhookClient.PollCallCount())
		assertWebhookPollInvocation(t, webhookClient, &operation, application, pollInvocations)
		for i := pollInvocations; i < pollInvocations+pollCallCount; i++ {
			assertWebhookPollInvocation(t, webhookClient, &operation, application, i)
		}

	})

	t.Run("Failed Async Webhook flow after timeout", func(t *testing.T) {
		operation := *mockedOperation
		operation.Name = "async-timeout-failed-operation"

		mode := graphql.WebhookModeAsync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(4)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{Location: &mockedLocationURL}, nil)
		webhookClient.PollReturns(prepareResponseStatus("IN_PROGRESS"), nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			var errMsg string
			for _, actualCondition := range operation.Status.Conditions {
				if actualCondition.Type == v1alpha1.ConditionTypeError {
					errMsg = actualCondition.Message
				}
			}

			return strings.Contains(errMsg, controllers.ErrWebhookTimeoutReached.Error())
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

	})

	t.Run("Successful Sync Webhook flow", func(t *testing.T) {

		operation := *mockedOperation
		operation.Name = "sync-success-operation"

		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID: operation.Spec.WebhookIDs[0],
					State:     v1alpha1.StateSuccess,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   v1alpha1.ConditionTypeError,
					Status: corev1.ConditionFalse,
				},
			},
			Phase:              v1alpha1.StateSuccess,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assert.Equal(t, expectedStatus, operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+1, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

	})

	t.Run("Failed Sync Webhook flow after timeout", func(t *testing.T) {
		operation := *mockedOperation
		operation.Name = "sync-timeout-failed-operation"

		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode, Timeout: intToIntPtr(4)})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(nil, mockedErr)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		defer func() {
			err := kubeClient.Delete(ctx, &operation)
			require.NoError(t, err)
		}()

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			var errMsg string
			for _, actualCondition := range operation.Status.Conditions {
				if actualCondition.Type == v1alpha1.ConditionTypeError {
					errMsg = actualCondition.Message
				}
			}

			return strings.Contains(errMsg, controllers.ErrWebhookTimeoutReached.Error())
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)
	})

	t.Run("Generation Changed Predicate does not allow requeue as a result of status update", func(t *testing.T) {
		// GIVEN:
		operation := *mockedOperation
		operation.Name = "sync-delete-operation"

		mode := graphql.WebhookModeSync
		application := prepareApplicationOutput(&graphql.Application{BaseEntity: &graphql.BaseEntity{}}, graphql.Webhook{ID: webhookGUID, Mode: &mode})

		directorClient.FetchApplicationReturns(application, nil)
		directorClient.UpdateOperationReturns(nil)

		webhookClient.DoReturns(&web_hook.Response{}, nil)

		updateInvocationVars(&fetchApplicationInvocations, &updateOperationInvocations, &doInvocations, &pollInvocations, directorClient, webhookClient)

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID: operation.Spec.WebhookIDs[0],
					State:     v1alpha1.StateSuccess,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   v1alpha1.ConditionTypeError,
					Status: corev1.ConditionFalse,
				},
			},
			Phase:              v1alpha1.StateSuccess,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assert.Equal(t, expectedStatus, operation.Status)
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
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

	})

	t.Run("Operation is deleted after ROT has expired", func(t *testing.T) {

		operation := *mockedOperation
		operation.Name = "sync-rot-operation"

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

		err := kubeClient.Create(ctx, &operation)
		require.NoError(t, err)

		var generation int64 = 1
		expectedStatus := v1alpha1.OperationStatus{
			Webhooks: []v1alpha1.Webhook{
				{
					WebhookID: operation.Spec.WebhookIDs[0],
					State:     v1alpha1.StateSuccess,
				},
			},
			Conditions: []v1alpha1.Condition{
				{
					Type:   v1alpha1.ConditionTypeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   v1alpha1.ConditionTypeError,
					Status: corev1.ConditionFalse,
				},
			},
			Phase:              v1alpha1.StateSuccess,
			ObservedGeneration: &generation,
		}

		namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

		require.Eventually(t, func() bool {
			var operation = &v1alpha1.Operation{}
			err := kubeClient.Get(ctx, namespacedName, operation)
			require.NoError(t, err)

			return assert.Equal(t, expectedStatus, operation.Status)
		}, eventuallyTimeout, eventuallyTick)

		var latestOperation = &v1alpha1.Operation{}
		err = kubeClient.Get(ctx, namespacedName, latestOperation)
		require.NoError(t, err)

		latestOperation.Spec.OperationCategory = "test"

		err = kubeClient.Update(ctx, latestOperation)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			err := kubeClient.Get(ctx, namespacedName, latestOperation)

			return kubeerrors.IsNotFound(err)
		}, eventuallyTimeout, eventuallyTick)

		require.Equal(t, fetchApplicationInvocations+2, directorClient.FetchApplicationCallCount())
		assertDirectorFetchApplicationInvocation(t, directorClient, operation.Spec.ResourceID, tenantGUID, fetchApplicationInvocations)

		require.Equal(t, updateOperationInvocations+1, directorClient.UpdateOperationCallCount())
		assertDirectorUpdateOperationInvocation(t, directorClient, &operation, updateOperationInvocations)

		require.Equal(t, doInvocations+1, webhookClient.DoCallCount())
		assertWebhookDoInvocation(t, webhookClient, &operation, application, doInvocations)

	})

	err = testEnv.Stop() // deferring the Stop earlier at the top does not seem to work, this is why the Stop is left here at the bottom
	require.NoError(t, err)
}

func updateInvocationVars(fetchAppCount, updateOpCount, doCount, pollCount *int, directorClient *controllersfakes.FakeDirectorClient, webhookClient *controllersfakes.FakeWebhookClient) {
	*fetchAppCount = directorClient.FetchApplicationCallCount()
	*updateOpCount = directorClient.UpdateOperationCallCount()

	*doCount = webhookClient.DoCallCount()
	*pollCount = webhookClient.PollCallCount()
}
