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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	web_hook "github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	client.Client
	Config         web_hook.Config
	DirectorClient director.Client
	WebhookClient  web_hook.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("operation", req.NamespacedName)

	var operation = &v1alpha1.Operation{}
	err := r.Get(ctx, req.NamespacedName, operation)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	var requestData webhook.RequestData
	if err := json.Unmarshal([]byte(operation.Spec.RequestData), &requestData); err != nil {
		logger.Error(err, operationError("Unable to parse request data", operation))
		return ctrl.Result{}, nil
	}

	ctx = tenant.SaveToContext(ctx, requestData.TenantID)
	app, err := r.DirectorClient.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		logger.Error(err, operationError("Unable to fetch application", operation))

		if isReconciliationTimeoutReached(operation, time.Duration(r.Config.TimeoutFactor)*r.Config.ReconciliationTimeout) {
			if err := r.Delete(ctx, operation); err != nil {
				logger.Error(err, operationError("Unable to delete operation", operation))
				return ctrl.Result{Requeue: true}, err
			}
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: r.Config.RequeueInterval}, nil
	}

	operation = setDefaultStatus(*operation)

	if app.Result.Ready {
		operation = setConditionStatus(*operation, v1alpha1.ConditionTypeReady, corev1.ConditionTrue, "")
		if app.Result.Error != nil {
			operation = setConditionStatus(*operation, v1alpha1.ConditionTypeError, corev1.ConditionFalse, "")
		}
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	if app.Result.Error != nil {
		operation = setErrorStatus(*operation, *app.Result.Error)
		operation.Status.Phase = v1alpha1.StateFailed
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	webhooks, err := retrieveWebhooks(app.Result.Webhooks, operation.Spec.WebhookIDs)
	if err != nil {
		logger.Error(err, operationError("Unable to retrieve webhooks", operation))
		return ctrl.Result{}, nil
	}

	if isReconciliationTimeoutReached(operation, calculateReconciliationTimeout(webhooks, r.Config.WebhookTimeout)) {
		opErr := errors.New("reconciliation timeout reached")

		err := r.DirectorClient.Notify(ctx, prepareDirectorRequest(*operation, opErr))
		if err != nil {
			logger.Error(err, operationError("Unable to notify Director", operation))
			return ctrl.Result{RequeueAfter: r.Config.RequeueInterval}, nil
		}

		operation = setErrorStatus(*operation, opErr.Error())
		operation.Status.Phase = v1alpha1.StateFailed
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, opErr
	}

	webhookEntity := webhooks[operation.Spec.WebhookIDs[0]]
	webhookStatus := operation.Status.Webhooks[0]

	if webhookStatus.WebhookPollURL == "" {
		request := web_hook.NewRequest(webhookEntity, requestData, operation.Spec.CorrelationID, operation.ObjectMeta.CreationTimestamp.Time, r.Config.RequeueInterval)

		response, err := r.WebhookClient.Do(ctx, request)
		if err != nil {
			logger.Error(err, operationError("Unable to execute Webhook request", operation))
			return r.handleWebhookTimeoutCheckResponse(ctx, operation, webhookEntity, err)
		}

		switch *webhookEntity.Mode {
		case graphql.WebhookModeAsync:
			operation.Status.Webhooks[0].WebhookPollURL = *response.Location
		case graphql.WebhookModeSync:
			err := r.DirectorClient.Notify(ctx, prepareDirectorRequest(*operation, nil))
			if err != nil {
				return ctrl.Result{}, err
			}
			operation = setReadyStatus(*operation)
			operation.Status.Phase = v1alpha1.StateSuccess
		default:
			return ctrl.Result{}, errors.New("unsupported webhook mode")
		}

		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update poll URL", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	requeueAfter, err := getNextPollTime(&webhookEntity, webhookStatus, r.Config.TimeLayout)
	if err != nil {
		logger.Error(err, operationError("Unable to calculate next poll time", operation))
		return ctrl.Result{Requeue: true}, err
	}

	if requeueAfter > 0 {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	request := web_hook.NewRequest(webhookEntity, requestData, operation.Spec.CorrelationID, operation.ObjectMeta.CreationTimestamp.Time, r.Config.RequeueInterval)

	request.PollURL = &webhookStatus.WebhookPollURL
	response, err := r.WebhookClient.Poll(ctx, request)
	if err != nil {
		logger.Error(err, operationError("Unable to execute Webhook Poll request", operation))
		return r.handleWebhookTimeoutCheckResponse(ctx, operation, webhookEntity, err)
	}

	switch *response.Status {
	case *response.InProgressStatusIdentifier:
		return r.handleWebhookTimeoutCheckResponse(ctx, operation, webhookEntity, errors.New("webhook timeout reached"))
	case *response.SuccessStatusIdentifier:
		err := r.DirectorClient.Notify(ctx, prepareDirectorRequest(*operation, nil))
		if err != nil {
			return ctrl.Result{}, err
		}

		operation = setReadyStatus(*operation)
		operation.Status.Phase = v1alpha1.StateSuccess
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("unable to update operation after it has completed successfully", operation))
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	case *response.FailedStatusIdentifier:
		opErr := errors.New("webhook operation has finished unsuccessfully")

		err := r.DirectorClient.Notify(ctx, prepareDirectorRequest(*operation, opErr))
		if err != nil {
			return ctrl.Result{}, err
		}

		operation = setErrorStatus(*operation, opErr.Error())
		operation.Status.Phase = v1alpha1.StateFailed
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("unable to update operation after it has finished unsuccessfully", operation))
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unexpected poll status response: %s", *response.Status)
	}
}

func (r *OperationReconciler) handleWebhookTimeoutCheckResponse(ctx context.Context, operation *v1alpha1.Operation, webhook graphql.Webhook, webhookErr error) (ctrl.Result, error) {
	if !isWebhookTimeoutReached(operation.CreationTimestamp.Time, time.Duration(*webhook.Timeout)) {
		requeueAfter := r.Config.RequeueInterval
		if webhook.RetryInterval != nil {
			requeueAfter = time.Duration(*webhook.RetryInterval)
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	err := r.DirectorClient.Notify(ctx, prepareDirectorRequest(*operation, webhookErr))
	if err != nil {
		return ctrl.Result{}, err
	}

	operation = setErrorStatus(*operation, webhookErr.Error())
	operation.Status.Phase = v1alpha1.StateFailed
	if err := r.Update(ctx, operation); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func prepareDirectorRequest(operation v1alpha1.Operation, err error) director.Request {
	return director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceID),
		ResourceID:    operation.Spec.ResourceID,
		Error:         err.Error(),
	}
}

func isReconciliationTimeoutReached(operation *v1alpha1.Operation, reconciliationTimeout time.Duration) bool {
	operationEndTime := operation.ObjectMeta.CreationTimestamp.Time.Add(reconciliationTimeout)

	return time.Now().After(operationEndTime)
}

func isWebhookTimeoutReached(creationTime time.Time, webhookTimeout time.Duration) bool {
	operationEndTime := creationTime.Add(webhookTimeout)
	return time.Now().After(operationEndTime)
}

func getNextPollTime(webhook *graphql.Webhook, webhookStatus v1alpha1.Webhook, timeLayout string) (time.Duration, error) {
	lastPollTimestamp, err := time.Parse(timeLayout, webhookStatus.LastPollTimestamp)
	if err != nil {
		return 0, err
	}

	nextPollTime := lastPollTimestamp.Add(time.Duration(*webhook.RetryInterval) * time.Minute)
	return nextPollTime.Sub(time.Now()), nil
}

func setDefaultStatus(operation v1alpha1.Operation) *v1alpha1.Operation {
	overrideStatus := false
	if operation.Generation != operation.Status.ObservedGeneration {
		overrideStatus = true
		operation.Status.ObservedGeneration = operation.Generation
		operation.Status.Phase = v1alpha1.StatePolling
	}

	if operation.Status.Conditions == nil || overrideStatus {
		operation.Status.Conditions = make([]v1alpha1.Condition, 0)
	}

	hasReady := false
	hasError := false
	for _, condition := range operation.Status.Conditions {
		if condition.Type == v1alpha1.ConditionTypeReady {
			hasReady = true
		}

		if condition.Type == v1alpha1.ConditionTypeError {
			hasError = true
		}
	}

	if !hasReady {
		operation.Status.Conditions = append(operation.Status.Conditions, v1alpha1.Condition{
			Type:   v1alpha1.ConditionTypeReady,
			Status: corev1.ConditionFalse,
		})
	}

	if !hasError {
		operation.Status.Conditions = append(operation.Status.Conditions, v1alpha1.Condition{
			Type:   v1alpha1.ConditionTypeError,
			Status: corev1.ConditionFalse,
		})
	}

	if operation.Status.Webhooks == nil || overrideStatus {
		operation.Status.Webhooks = make([]v1alpha1.Webhook, 0)
	}

	for _, opWebhookID := range operation.Spec.WebhookIDs {
		webhookExists := false
		for _, opWebhook := range operation.Status.Webhooks {
			if opWebhookID == opWebhook.WebhookID {
				webhookExists = true
			}
		}

		if !webhookExists {
			operation.Status.Webhooks = append(operation.Status.Webhooks, v1alpha1.Webhook{
				WebhookID: opWebhookID,
				State:     v1alpha1.StatePolling,
			})
		}
	}

	return &operation
}

func setConditionStatus(operation v1alpha1.Operation, conditionType v1alpha1.ConditionType, status corev1.ConditionStatus, msg string) *v1alpha1.Operation {
	for _, condition := range operation.Status.Conditions {
		if condition.Type == conditionType {
			condition.Status = status
			condition.Message = msg
		}
	}
	return &operation
}

func setReadyStatus(operation v1alpha1.Operation) *v1alpha1.Operation {
	operation = *setConditionStatus(operation, v1alpha1.ConditionTypeReady, corev1.ConditionTrue, "")
	operation = *setConditionStatus(operation, v1alpha1.ConditionTypeError, corev1.ConditionFalse, "")
	return &operation
}

func setErrorStatus(operation v1alpha1.Operation, errorMsg string) *v1alpha1.Operation {
	operation = *setConditionStatus(operation, v1alpha1.ConditionTypeReady, corev1.ConditionFalse, "")
	operation = *setConditionStatus(operation, v1alpha1.ConditionTypeError, corev1.ConditionTrue, errorMsg)
	return &operation
}

func retrieveWebhooks(appWebhooks []graphql.Webhook, opWebhookIDs []string) (map[string]graphql.Webhook, error) {
	if len(opWebhookIDs) == 0 {
		return nil, errors.New("no webhooks found for operation")
	}

	if len(opWebhookIDs) > 1 {
		return nil, errors.New("multiple webhooks per operation are not supported")
	}

	webhooks := make(map[string]graphql.Webhook, 0)

	for _, opWebhookID := range opWebhookIDs {
		webhookExists := false
		for _, appWebhook := range appWebhooks {
			if opWebhookID == appWebhook.ID {
				webhooks[opWebhookID] = appWebhook
				webhookExists = true
			}
		}

		if !webhookExists {
			return nil, fmt.Errorf("missing webhook with ID: %s", opWebhookID)
		}
	}

	return webhooks, nil
}

func calculateReconciliationTimeout(webhooks map[string]graphql.Webhook, defaultWebhookTimeout time.Duration) time.Duration {
	var totalTimeout time.Duration
	for _, webhook := range webhooks {
		if webhook.Timeout == nil {
			totalTimeout += defaultWebhookTimeout
		} else {
			totalTimeout += time.Duration(*webhook.Timeout) * time.Minute
		}
	}

	return totalTimeout
}

func operationError(msg string, operation *v1alpha1.Operation) string {
	return fmt.Sprintf("%s for operation of resource with ID %s and type %s", msg, operation.Spec.ResourceID, operation.Spec.ResourceType)
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
