/*
Copyright 2021.

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
	directoroperation "github.com/kyma-incubator/compass/components/director/pkg/operation"
	"strings"
	"time"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/metrics"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/errors"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/log"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

// OperationReconciler reconciles an Operation object
type OperationReconciler struct {
	config           *webhook.Config
	statusManager    StatusManager
	k8sClient        KubernetesClient
	directorClient   DirectorClient
	webhookClient    WebhookClient
	metricsCollector *metrics.Collector
}

func NewOperationReconciler(config *webhook.Config, statusManager StatusManager, k8sClient KubernetesClient, directorClient DirectorClient, webhookClient WebhookClient, collector *metrics.Collector) *OperationReconciler {
	return &OperationReconciler{
		config:           config,
		statusManager:    statusManager,
		k8sClient:        k8sClient,
		directorClient:   directorClient,
		webhookClient:    webhookClient,
		metricsCollector: collector,
	}
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

// Reconcile contains the Operations Controller logic concerned with processing and executing webhooks based on Operation CRs
func (r *OperationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := ctrl.Log.WithValues("operation", req.NamespacedName)
	ctx = log.ContextWithLogger(ctx, logger)

	operation, err := r.k8sClient.Get(ctx, req.NamespacedName)
	if err != nil {
		return r.handleGetOperationError(ctx, &req, err)
	}

	if err := r.statusManager.Initialize(ctx, operation); err != nil {
		return r.handleInitializationError(ctx, err, operation)
	}

	requestObject, err := operation.RequestObject()
	if err != nil {
		log.C(ctx).Error(err, "Unable to parse request object")
		return r.finalizeStatusWithError(ctx, operation, err, nil)
	}

	ctx = tenant.SaveToContext(ctx, requestObject.TenantID)
	app, err := r.directorClient.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		return r.handleFetchApplicationError(ctx, operation, err)
	}

	if app.Result.Ready {
		return r.finalizeStatus(ctx, operation, app.Result.Error, nil)
	}

	if len(operation.Spec.WebhookIDs) == 0 {
		log.C(ctx).Info("No webhook defined. Operation executed successfully")
		return r.finalizeStatusSuccess(ctx, operation, nil)
	}

	webhookEntity, err := extractWebhook(app.Result.Webhooks, operation.Spec.WebhookIDs[0])
	if err != nil {
		log.C(ctx).Error(err, "Unable to retrieve webhook")
		return r.finalizeStatusWithError(ctx, operation, err, nil)
	}

	if operation.TimeoutReached(r.determineTimeout(webhookEntity)) {
		log.C(ctx).Info("Reconciliation timeout reached")
		return r.finalizeStatusWithError(ctx, operation, errors.ErrWebhookTimeoutReached, webhookEntity)
	}

	if !operation.HasPollURL() {
		log.C(ctx).Info("Webhook Poll URL is not found. Will attempt to execute the webhook")
		request := webhookclient.NewRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID)
		isDeleteOrUnpair := operation.Spec.OperationType == v1alpha1.OperationTypeDelete ||
			(operation.Spec.OperationType == v1alpha1.OperationTypeUpdate && operation.Spec.OperationCategory == directoroperation.OperationCategoryUnpairApplication)
		response, err := r.webhookClient.Do(ctx, request)
		if errors.IsWebhookStatusGoneErr(err) && isDeleteOrUnpair {
			log.C(ctx).Info(fmt.Sprintf("%s webhook initial request returned gone status %d", *(webhookEntity.Mode), *response.GoneStatusCode))
			return r.finalizeStatusSuccess(ctx, operation, webhookEntity)
		}
		if err != nil {
			log.C(ctx).Error(err, "Unable to execute Webhook request")
			return r.requeueUnlessTimeoutOrFatalError(ctx, operation, webhookEntity, err)
		}

		return r.handleWebhookResponse(ctx, operation, webhookEntity.Mode, response)
	}

	log.C(ctx).Info("Webhook Poll URL is found. Will calculate next poll time")
	requeueAfter, err := operation.NextPollTime(webhookEntity.RetryInterval, r.config.TimeLayout)
	if err != nil {
		log.C(ctx).Error(err, "Unable to calculate next poll time")
		return r.finalizeStatusWithError(ctx, operation, err, webhookEntity)
	}

	if requeueAfter > 0 {
		log.C(ctx).Info(fmt.Sprintf("Poll interval has not passed. Will requeue after: %d seconds", requeueAfter*time.Second))
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	request := webhookclient.NewPollRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID, operation.PollURL())
	response, err := r.webhookClient.Poll(ctx, request)
	if err != nil {
		log.C(ctx).Error(err, "Unable to execute Webhook Poll request")
		return r.requeueUnlessTimeoutOrFatalError(ctx, operation, webhookEntity, err)
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
		log.C(ctx).Error(err, "Validation error occurred during operation status initialization")
		return r.finalizeStatusWithError(ctx, operation, err, nil)
	}

	return ctrl.Result{}, err
}

func (r *OperationReconciler) handleGetOperationError(ctx context.Context, req *ctrl.Request, err error) (ctrl.Result, error) {
	log.C(ctx).Error(err, fmt.Sprintf("Unable to retrieve %s resource from API server", req.NamespacedName))
	if kubeerrors.IsNotFound(err) {
		log.C(ctx).Error(err, fmt.Sprintf("%s resource was not found in API server", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, err
}

func (r *OperationReconciler) handleFetchApplicationError(ctx context.Context, operation *v1alpha1.Operation, err error) (ctrl.Result, error) {
	log.C(ctx).Error(err, fmt.Sprintf("Unable to fetch application with ID %s", operation.Spec.ResourceID))
	if operation.TimeoutReached(time.Duration(r.config.TimeoutFactor) * r.config.WebhookTimeout) {
		if err := r.k8sClient.Delete(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.C(ctx).Info("Successfully deleted operation")
		return ctrl.Result{}, nil
	}

	if isNotFoundError(err) {
		if operation.Status.Phase == v1alpha1.StateSuccess || operation.Status.Phase == v1alpha1.StateFailed {
			log.C(ctx).Info(fmt.Sprintf("Last state of operation for application with ID %s is %s, will not requeue", operation.Spec.ResourceID, operation.Status.Phase))
			return ctrl.Result{}, nil
		}

		if operation.Spec.OperationType == v1alpha1.OperationTypeDelete && operation.Status.Phase == v1alpha1.StateInProgress {
			return r.finalizeStatus(ctx, operation, nil, nil)
		}
		if operation.Spec.OperationType == v1alpha1.OperationTypeUpdate && operation.Status.Phase == v1alpha1.StateInProgress {
			return r.finalizeStatus(ctx, operation, str.Ptr("Application not found in director"), nil)
		}
	}

	// requeue in case of async create (app is not created in director yet), or a glitch in controller<->director communication
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
		if err := r.statusManager.InProgressWithPollURL(ctx, operation, *response.Location); err != nil {
			return ctrl.Result{}, err
		}
		log.C(ctx).Info("Successfully updated operation status with poll URL: " + *response.Location)
		return ctrl.Result{Requeue: true}, nil
	case graphql.WebhookModeSync:
		log.C(ctx).Info("Synchronous webhook has been executed successfully")
		return r.finalizeStatusSuccess(ctx, operation, nil)
	default:
		log.C(ctx).Error(errors.ErrUnsupportedWebhookMode, "Unable to post-process Webhook response")
		return ctrl.Result{}, nil
	}
}

func (r *OperationReconciler) handleWebhookPollResponse(ctx context.Context, operation *v1alpha1.Operation, webhookEntity *graphql.Webhook, response *webhookdir.ResponseStatus) (ctrl.Result, error) {
	log.C(ctx).Info(fmt.Sprintf("Asynchronous webhook polling request has been executed successfully with response status: %s", *response.Status))
	switch *response.Status {
	case *response.InProgressStatusIdentifier:
		lastPollTimestamp := time.Now().Format(r.config.TimeLayout)
		retryCount := operation.Status.Webhooks[0].RetriesCount + 1
		if err := r.statusManager.InProgressWithPollURLAndLastPollTimestamp(ctx, operation, operation.PollURL(), lastPollTimestamp, retryCount); err != nil {
			return ctrl.Result{}, err
		}
		log.C(ctx).Info(fmt.Sprintf("Successfully updated operation status last poll timestamp to %s", lastPollTimestamp), "status", operation.Status)
		return r.requeueUnlessTimeoutOrFatalError(ctx, operation, webhookEntity, errors.ErrWebhookPollTimeExpired)
	case *response.SuccessStatusIdentifier:
		return r.finalizeStatusSuccess(ctx, operation, webhookEntity)
	case *response.FailedStatusIdentifier:
		return r.finalizeStatusWithError(ctx, operation, errors.ErrFailedWebhookStatus, webhookEntity)
	default:
		log.C(ctx).Error(fmt.Errorf("unexpected poll status response: %s", *response.Status), "Polling will be stopped due to an unknown status code received")
		return ctrl.Result{}, nil
	}
}

func (r *OperationReconciler) requeueUnlessTimeoutOrFatalError(ctx context.Context, operation *v1alpha1.Operation, webhookEntity *graphql.Webhook, webhookErr error) (ctrl.Result, error) {
	_, isFatalErr := webhookErr.(*errors.FatalReconcileErr)
	if !operation.TimeoutReached(r.determineTimeout(webhookEntity)) && !isFatalErr {
		requeueAfter := r.config.RequeueInterval
		if webhookEntity.RetryInterval != nil {
			requeueAfter = time.Duration(*webhookEntity.RetryInterval)
		}

		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	if !isFatalErr {
		webhookErr = fmt.Errorf("%s: %s", errors.ErrWebhookTimeoutReached, webhookErr)
	}

	return r.finalizeStatusWithError(ctx, operation, webhookErr, webhookEntity)
}

func (r *OperationReconciler) finalizeStatus(ctx context.Context, operation *v1alpha1.Operation, errorMsg *string, webhook *graphql.Webhook) (ctrl.Result, error) {
	if isCloseToTimeout(operation.Status.InitializedAt.Time, r.determineTimeout(webhook)) {
		r.metricsCollector.RecordOperationInProgressNearTimeout(string(operation.Spec.OperationType), operation.ObjectMeta.Name)
	}

	if errorMsg != nil && *errorMsg != "" {
		r.metricsCollector.RecordError(operation.ObjectMeta.Name,
			operation.Spec.CorrelationID,
			string(operation.Spec.OperationType),
			operation.Spec.OperationCategory,
			trimRequestObject(operation),
			*errorMsg)

		if err := r.statusManager.FailedStatus(ctx, operation, *errorMsg); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.statusManager.SuccessStatus(ctx, operation); err != nil {
		return ctrl.Result{}, err
	}

	duration := time.Now().Sub(operation.Status.InitializedAt.Time)
	r.metricsCollector.RecordLatency(string(operation.Spec.OperationType), duration)

	return ctrl.Result{}, nil
}

func (r *OperationReconciler) finalizeStatusSuccess(ctx context.Context, operation *v1alpha1.Operation, webhook *graphql.Webhook) (ctrl.Result, error) {
	if isCloseToTimeout(operation.Status.InitializedAt.Time, r.determineTimeout(webhook)) {
		r.metricsCollector.RecordOperationInProgressNearTimeout(string(operation.Spec.OperationType), operation.ObjectMeta.Name)
	}

	if err := r.directorClient.UpdateOperation(ctx, prepareDirectorRequest(operation)); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.statusManager.SuccessStatus(ctx, operation); err != nil {
		return ctrl.Result{}, err
	}

	log.C(ctx).Info("Successfully updated operation status to succeeded")

	duration := time.Now().Sub(operation.Status.InitializedAt.Time)
	r.metricsCollector.RecordLatency(string(operation.Spec.OperationType), duration)

	return ctrl.Result{}, nil
}

func (r *OperationReconciler) finalizeStatusWithError(ctx context.Context, operation *v1alpha1.Operation, opErr error, webhook *graphql.Webhook) (ctrl.Result, error) {
	if operation != nil && isCloseToTimeout(operation.Status.InitializedAt.Time, r.determineTimeout(webhook)) {
		r.metricsCollector.RecordOperationInProgressNearTimeout(string(operation.Spec.OperationType), operation.ObjectMeta.Name)
	}

	r.metricsCollector.RecordError(operation.ObjectMeta.Name,
		operation.Spec.CorrelationID,
		string(operation.Spec.OperationType),
		operation.Spec.OperationCategory,
		trimRequestObject(operation),
		opErr.Error())

	if err := r.directorClient.UpdateOperation(ctx, prepareDirectorRequestWithError(operation, opErr)); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.statusManager.FailedStatus(ctx, operation, opErr.Error()); err != nil {
		return ctrl.Result{}, err
	}

	log.C(ctx).Info("Successfully updated operation status to failed")
	return ctrl.Result{}, nil
}

func (r *OperationReconciler) determineTimeout(webhook *graphql.Webhook) time.Duration {
	if webhook == nil {
		return r.config.WebhookTimeout
	}

	if webhook.Timeout == nil {
		return r.config.WebhookTimeout
	}

	return time.Duration(*webhook.Timeout) * time.Second
}

func prepareDirectorRequest(operation *v1alpha1.Operation) *director.Request {
	return prepareDirectorRequestWithError(operation, nil)
}

func prepareDirectorRequestWithError(operation *v1alpha1.Operation, err error) *director.Request {
	request := &director.Request{
		OperationType:     graphql.OperationType(operation.Spec.OperationType),
		ResourceType:      resource.Type(operation.Spec.ResourceType),
		ResourceID:        operation.Spec.ResourceID,
		OperationCategory: operation.Spec.OperationCategory,
	}

	if err != nil {
		request.Error = err.Error()
	}

	return request
}

func extractWebhook(appWebhooks []graphql.Webhook, operationWebhookID string) (*graphql.Webhook, error) {
	for _, appWebhook := range appWebhooks {
		if appWebhook.ID == operationWebhookID {
			return &appWebhook, nil
		}
	}

	return nil, fmt.Errorf("missing webhook with ID: %s", operationWebhookID)
}

func trimRequestObject(operation *v1alpha1.Operation) string {
	index := strings.Index(operation.Spec.RequestObject, ",\"Headers\"")
	if index != -1 {
		requestObj := []rune(operation.Spec.RequestObject)[:index]
		requestObj = append(requestObj, '}')
		return string(requestObj)
	}

	return operation.Spec.RequestObject
}

func isCloseToTimeout(initialization time.Time, timeout time.Duration) bool {
	inProgress := time.Now().Sub(initialization)
	return inProgress.Seconds() > timeout.Seconds()*0.9
}
