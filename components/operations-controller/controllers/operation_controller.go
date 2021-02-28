/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/log"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

// OperationReconciler reconciles an Operation object
type OperationReconciler struct {
	config         *webhook.Config
	statusManager  StatusManager
	k8sClient      KubernetesClient
	directorClient DirectorClient
	webhookClient  WebhookClient
}

func NewOperationReconciler(config *webhook.Config, statusManager StatusManager, k8sClient KubernetesClient, directorClient DirectorClient, webhookClient WebhookClient) *OperationReconciler {
	return &OperationReconciler{
		config:         config,
		statusManager:  statusManager,
		k8sClient:      k8sClient,
		directorClient: directorClient,
		webhookClient:  webhookClient,
	}
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

// Reconcile contains the Operations Controller logic concerned with processing and executing webhooks based on Operation CRs
func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.Log.WithValues("operation", req.NamespacedName)
	ctx := log.ContextWithLogger(context.Background(), logger)

	operation, err := r.k8sClient.Get(ctx, req.NamespacedName)
	if err != nil {
		return r.handleGetOperationError(ctx, &req, err)
	}

	if err := r.statusManager.Initialize(ctx, req.NamespacedName); err != nil {
		return r.handleInitializationError(ctx, err, operation)
	}

	requestObject, err := operation.RequestObject()
	if err != nil {
		log.C(ctx).Error(err, "Unable to parse request object")
		return r.finalizeStatusWithError(ctx, operation, err)
	}

	ctx = tenant.SaveToContext(ctx, requestObject.TenantID)
	app, err := r.directorClient.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		return r.handleFetchApplicationError(ctx, operation, err)
	}

	if app.Result.Ready {
		return r.finalizeStatusForApplication(ctx, req.NamespacedName, app.Result.Error)
	}

	webhookEntity, err := retrieveWebhook(app.Result.Webhooks, operation.Spec.WebhookIDs[0])
	if err != nil {
		log.C(ctx).Error(err, "Unable to retrieve webhook")
		return r.finalizeStatusWithError(ctx, operation, err)
	}

	if operation.TimeoutReached(calculateTimeout(webhookEntity, r.config.WebhookTimeout)) {
		log.C(ctx).Info("Reconciliation timeout reached")
		return r.finalizeStatusWithError(ctx, operation, ErrReconciliationTimeoutReached)
	}

	if !operation.HasPollURL() {
		log.C(ctx).Info("Webhook Poll URL is not found. Will attempt to execute the webhook")
		request := webhook.NewRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID)

		response, err := r.webhookClient.Do(ctx, request)
		if err != nil {
			log.C(ctx).Error(err, "Unable to execute Webhook request")
			return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
		}

		return r.handleWebhookResponse(ctx, operation, webhookEntity.Mode, response)
	}

	log.C(ctx).Info("Webhook Poll URL is found. Will calculate next poll time")
	requeueAfter, err := operation.NextPollTime(webhookEntity.RetryInterval, r.config.TimeLayout)
	if err != nil {
		log.C(ctx).Error(err, "Unable to calculate next poll time")
		return r.finalizeStatusWithError(ctx, operation, err)
	}

	if requeueAfter > 0 {
		log.C(ctx).Info(fmt.Sprintf("Poll interval has not passed. Will requeue after: %d seconds", requeueAfter*time.Second))
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	request := webhook.NewPollRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID, operation.PollURL())
	response, err := r.webhookClient.Poll(ctx, request)
	if err != nil {
		log.C(ctx).Error(err, "Unable to execute Webhook Poll request")
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
	}

	return r.handleWebhookPollResponse(ctx, operation, webhookEntity, response)
}

func (r *OperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		}).
		For(&v1alpha1.Operation{}).
		Complete(r)
}

func (r *OperationReconciler) handleInitializationError(ctx context.Context, err error, operation *v1alpha1.Operation) (ctrl.Result, error) {
	log.C(ctx).Error(err, "Failed to initialize operation status")
	if _, ok := err.(*v1alpha1.OperationValidationErr); ok {
		return r.finalizeStatusWithError(ctx, operation, err)
	}
	return ctrl.Result{}, err
}

func (r *OperationReconciler) handleGetOperationError(ctx context.Context, req *ctrl.Request, err error) (ctrl.Result, error) {
	log.C(ctx).Error(err, fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))
	if kubeerrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}

func (r *OperationReconciler) handleFetchApplicationError(ctx context.Context, operation *v1alpha1.Operation, err error) (ctrl.Result, error) {
	log.C(ctx).Error(err, "Unable to fetch application")
	if operation.TimeoutReached(time.Duration(r.config.TimeoutFactor) * r.config.WebhookTimeout) {
		if err := r.k8sClient.Delete(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}
		log.C(ctx).Info("Successfully deleted operation")
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}

func (r *OperationReconciler) handleWebhookResponse(ctx context.Context, operation *v1alpha1.Operation, webhookMode *graphql.WebhookMode, response *webhookdir.Response) (ctrl.Result, error) {
	mode := graphql.WebhookModeSync
	if webhookMode != nil {
		mode = *webhookMode
	}

	switch mode {
	case graphql.WebhookModeAsync:
		log.C(ctx).Info("Asynchronous webhook initial request has been executed successfully")
		namespacedName := types.NamespacedName{Name: operation.Name, Namespace: operation.Namespace}
		if err := r.statusManager.InProgressWithPollURL(ctx, namespacedName, *response.Location); err != nil {
			return ctrl.Result{}, err
		}
		log.C(ctx).Info("Successfully updated operation status", "status", operation.Status)
		return ctrl.Result{Requeue: true}, nil
	case graphql.WebhookModeSync:
		log.C(ctx).Info("Synchronous webhook has been executed successfully")
		return r.finalizeStatusSuccess(ctx, operation)
	default:
		log.C(ctx).Error(ErrUnsupportedWebhookMode, "Unable to post-process Webhook response")
		return ctrl.Result{}, nil
	}
}

func (r *OperationReconciler) handleWebhookPollResponse(ctx context.Context, operation *v1alpha1.Operation, webhookEntity *graphql.Webhook, response *webhookdir.ResponseStatus) (ctrl.Result, error) {
	log.C(ctx).Info(fmt.Sprintf("Asynchronous webhook polling request has been executed successfully with response status: %s", *response.Status))
	switch *response.Status {
	case *response.InProgressStatusIdentifier:
		namespacedName := types.NamespacedName{Name: operation.Name, Namespace: operation.Namespace}
		lastPollTimestamp := time.Now().Format(r.config.TimeLayout)
		retryCount := operation.Status.Webhooks[0].RetriesCount + 1
		if err := r.statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, namespacedName, operation.PollURL(), lastPollTimestamp, retryCount); err != nil {
			return ctrl.Result{}, err
		}
		log.C(ctx).Info(fmt.Sprintf("Successfully updated operation status last poll timestamp to %s", lastPollTimestamp), "status", operation.Status)
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, ErrWebhookTimeoutReached)
	case *response.SuccessStatusIdentifier:
		return r.finalizeStatusSuccess(ctx, operation)
	case *response.FailedStatusIdentifier:
		return r.finalizeStatusWithError(ctx, operation, ErrFailedWebhookStatus)
	default:
		log.C(ctx).Error(fmt.Errorf("unexpected poll status response: %s", *response.Status), "Polling will be stopped due to an unknown status code received")
		return ctrl.Result{}, nil
	}
}

func (r *OperationReconciler) requeueIfTimeoutNotReached(ctx context.Context, operation *v1alpha1.Operation, webhook *graphql.Webhook, webhookErr error) (ctrl.Result, error) {
	if !operation.TimeoutReached(calculateTimeout(webhook, r.config.WebhookTimeout)) {
		requeueAfter := r.config.RequeueInterval
		if webhook.RetryInterval != nil {
			requeueAfter = time.Duration(*webhook.RetryInterval)
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	return r.finalizeStatusWithError(ctx, operation, webhookErr)
}

func (r *OperationReconciler) finalizeStatusForApplication(ctx context.Context, name types.NamespacedName, errorMsg *string) (ctrl.Result, error) {
	if errorMsg != nil && *errorMsg != "" {
		if err := r.statusManager.FailedStatus(ctx, name, *errorMsg); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.statusManager.SuccessStatus(ctx, name); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *OperationReconciler) finalizeStatusSuccess(ctx context.Context, operation *v1alpha1.Operation) (ctrl.Result, error) {
	if err := r.directorClient.UpdateOperation(ctx, prepareDirectorRequest(operation)); err != nil {
		return ctrl.Result{}, err
	}

	namespacedName := types.NamespacedName{Name: operation.Name, Namespace: operation.Namespace}
	if err := r.statusManager.SuccessStatus(ctx, namespacedName); err != nil {
		return ctrl.Result{}, err
	}

	log.C(ctx).Info("Successfully updated operation status", "status", operation.Status)
	return ctrl.Result{}, nil
}

func (r *OperationReconciler) finalizeStatusWithError(ctx context.Context, operation *v1alpha1.Operation, opErr error) (ctrl.Result, error) {
	if err := r.directorClient.UpdateOperation(ctx, prepareDirectorRequestWithError(operation, opErr)); err != nil {
		return ctrl.Result{}, err
	}

	namespacedName := types.NamespacedName{Name: operation.Name, Namespace: operation.Namespace}
	if err := r.statusManager.FailedStatus(ctx, namespacedName, opErr.Error()); err != nil {
		return ctrl.Result{}, err
	}

	log.C(ctx).Info("Successfully updated operation status", "status", operation.Status)
	return ctrl.Result{}, nil
}

func prepareDirectorRequest(operation *v1alpha1.Operation) *director.Request {
	return prepareDirectorRequestWithError(operation, nil)
}

func prepareDirectorRequestWithError(operation *v1alpha1.Operation, err error) *director.Request {
	request := &director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}

	if err != nil {
		request.Error = err.Error()
	}

	return request
}

func retrieveWebhook(appWebhooks []graphql.Webhook, operationWebhookID string) (*graphql.Webhook, error) {
	for _, appWebhook := range appWebhooks {
		if appWebhook.ID == operationWebhookID {
			return &appWebhook, nil
		}
	}

	return nil, fmt.Errorf("missing webhook with ID: %s", operationWebhookID)
}

func calculateTimeout(webhook *graphql.Webhook, defaultWebhookTimeout time.Duration) time.Duration {
	if webhook.Timeout == nil {
		return defaultWebhookTimeout
	}
	return time.Duration(*webhook.Timeout) * time.Second
}
