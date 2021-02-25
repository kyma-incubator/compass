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
	"github.com/kyma-incubator/compass/components/operations-controller/internal/client"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	web_hook "github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

var (
	ErrReconciliationTimeoutReached = errors.New("reconciliation timeout reached")
	ErrWebhookTimeoutReached        = errors.New("webhook timeout reached")
	ErrFailedWebhookStatus          = errors.New("webhook operation has finished with failed status")
	ErrUnsupportedWebhookMode       = errors.New("unsupported webhook mode")
)

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	k8sClient      client.Client
	webhookConfig  *web_hook.Config
	directorClient director.Client
	webhookClient  web_hook.Client
	log            logr.Logger
}

func NewOperationReconciler(k8sClient client.Client, webhookConfig *web_hook.Config, directorClient director.Client, webhookClient web_hook.Client, logger logr.Logger) *OperationReconciler {
	return &OperationReconciler{
		k8sClient:      k8sClient,
		webhookConfig:  webhookConfig,
		log:            logger,
		directorClient: directorClient,
		webhookClient:  webhookClient,
	}
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.log.WithValues("operation", req.NamespacedName)

	operation, err := r.k8sClient.Get(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))
		if statusError, ok := err.(*k8s_errors.StatusError); ok && statusError.ErrStatus.Code == http.StatusNotFound {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	operation = prepareInitialStatus(*operation)

	if !isWebhookInProgress(operation) {
		log.Info(operationResourceMessage("Webhook associated with operation has already been executed. Will return with no requeue", operation))
		return ctrl.Result{}, err
	}

	webhookCount := len(operation.Spec.WebhookIDs)
	if webhookCount == 0 || webhookCount > 1 {
		err := fmt.Errorf("expected 1 webhook for execution, found %d", webhookCount)
		logger.Error(err, operationResourceMessage("Invalid operation webhooks provided", operation))
		return r.updateStatusWithError(ctx, operation, err)
	}

	requestData, err := parseRequestData(operation)
	if err != nil {
		logger.Error(err, operationResourceMessage("Unable to parse request data", operation))
		return r.updateStatusWithError(ctx, operation, err)
	}

	ctx = tenant.SaveToContext(ctx, requestData.TenantID)
	app, err := r.directorClient.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		logger.Error(err, operationResourceMessage("Unable to fetch application", operation))

		if isReconciliationTimeoutReached(operation, time.Duration(r.webhookConfig.TimeoutFactor)*r.webhookConfig.ReconciliationTimeout) {
			if err := r.k8sClient.Delete(ctx, operation); err != nil {
				return ctrl.Result{}, err
			}

			log.Info(operationResourceMessage("Successfully deleted operation", operation))
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if app.Result.Ready {
		operation = setConditionStatus(*operation, v1alpha1.ConditionTypeReady, corev1.ConditionTrue, "")
		operation.Status.Phase = v1alpha1.StateSuccess
		operation.Status.Webhooks[0].State = v1alpha1.StateSuccess
		if !isStringPointerEmpty(app.Result.Error) {
			operation = setConditionStatus(*operation, v1alpha1.ConditionTypeError, corev1.ConditionTrue, *app.Result.Error)
			operation.Status.Phase = v1alpha1.StateFailed
			operation.Status.Webhooks[0].State = v1alpha1.StateFailed
		}
		if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(operationResourceMessage("Successfully updated operation status", operation))
		return ctrl.Result{}, nil
	}

	if !isStringPointerEmpty(app.Result.Error) {
		operation = setErrorStatus(*operation, *app.Result.Error)
		operation.Status.Phase = v1alpha1.StateFailed
		operation.Status.Webhooks[0].State = v1alpha1.StateFailed
		if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(operationResourceMessage("Successfully updated operation status", operation))
		return ctrl.Result{}, nil
	}

	webhooks, err := retrieveWebhooks(app.Result.Webhooks, operation.Spec.WebhookIDs)
	if err != nil {
		logger.Error(err, operationResourceMessage("Unable to retrieve webhooks", operation))
		return r.updateStatusWithError(ctx, operation, err)
	}

	if isReconciliationTimeoutReached(operation, calculateReconciliationTimeout(webhooks, r.webhookConfig.WebhookTimeout)) {
		log.Info(operationResourceMessage("Reconciliation timeout reached", operation))
		return r.updateStatusWithError(ctx, operation, ErrReconciliationTimeoutReached)
	}

	webhookEntity := webhooks[operation.Spec.WebhookIDs[0]]
	webhookStatus := &operation.Status.Webhooks[0]

	if webhookStatus.WebhookPollURL == "" {
		log.Info(operationResourceMessage("Webhook Poll URL is not found. Will attempt to execute the webhook", operation))
		request := web_hook.NewRequest(webhookEntity, requestData, operation.Spec.CorrelationID)

		response, err := r.webhookClient.Do(ctx, request)
		if err != nil {
			logger.Error(err, operationResourceMessage("Unable to execute Webhook request", operation))
			return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
		}

		switch *webhookEntity.Mode {
		case graphql.WebhookModeAsync:
			log.Info(operationResourceMessage("Asynchronous webhook initial request has been executed successfully", operation))

			operation.Status.Webhooks[0].WebhookPollURL = *response.Location
			operation.Status.Webhooks[0].State = v1alpha1.StateInProgress
			operation.Status.Phase = v1alpha1.StateInProgress
			if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
				return ctrl.Result{}, err
			}

			log.Info(operationResourceMessage("Successfully updated operation status", operation))
			return ctrl.Result{Requeue: true}, nil
		case graphql.WebhookModeSync:
			log.Info(operationResourceMessage("Synchronous webhook has been executed successfully", operation))
			return r.updateStatusSuccess(ctx, operation)
		default:
			logger.Error(ErrUnsupportedWebhookMode, operationResourceMessage("Unable to post-process Webhook response", operation))
			return ctrl.Result{}, nil
		}
	}

	log.Info(operationResourceMessage("Webhook Poll URL is found. Will calculate next poll time", operation))
	requeueAfter, err := getNextPollTime(&webhookEntity, *webhookStatus, r.webhookConfig.TimeLayout)
	if err != nil {
		logger.Error(err, operationResourceMessage("Unable to calculate next poll time", operation))
		return r.updateStatusWithError(ctx, operation, err)
	}

	if requeueAfter > 0 {
		logger.Info(fmt.Sprintf("Poll interval has not passed. Will requeue after: %d seconds", requeueAfter*time.Second))
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	request := web_hook.NewPollRequest(webhookEntity, requestData, operation.Spec.CorrelationID, webhookStatus.WebhookPollURL)
	response, err := r.webhookClient.Poll(ctx, request)
	if err != nil {
		logger.Error(err, operationResourceMessage("Unable to execute Webhook Poll request", operation))
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
	}

	log.Info(operationResourceMessage(fmt.Sprintf("Asynchronous webhook polling request has been executed successfully with response status: %s", *response.Status), operation))
	switch *response.Status {
	case *response.InProgressStatusIdentifier:
		lastPollTimestamp := time.Now().Format(r.webhookConfig.TimeLayout)
		operation.Status.Webhooks[0].LastPollTimestamp = lastPollTimestamp
		if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(operationResourceMessage(fmt.Sprintf("Successfully updated operation status last poll timestamp to %s", lastPollTimestamp), operation))
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, ErrWebhookTimeoutReached)
	case *response.SuccessStatusIdentifier:
		return r.updateStatusSuccess(ctx, operation)
	case *response.FailedStatusIdentifier:
		return r.updateStatusWithError(ctx, operation, ErrFailedWebhookStatus)
	default:
		return ctrl.Result{}, fmt.Errorf("unexpected poll status response: %s", *response.Status)
	}
}

func isWebhookInProgress(operation *v1alpha1.Operation) bool {
	return operation.Status.Webhooks[0].State == v1alpha1.StateInProgress
}

func (r *OperationReconciler) updateStatusSuccess(ctx context.Context, operation *v1alpha1.Operation) (ctrl.Result, error) {
	return r.updateStatusWithError(ctx, operation, nil)
}

func (r *OperationReconciler) updateStatusWithError(ctx context.Context, operation *v1alpha1.Operation, opErr error) (ctrl.Result, error) {
	var state v1alpha1.State
	var request director.Request
	if opErr != nil {
		state = v1alpha1.StateFailed
		operation = setErrorStatus(*operation, opErr.Error())
		request = prepareDirectorRequestWithError(*operation, opErr)
	} else {
		state = v1alpha1.StateSuccess
		operation = setReadyStatus(*operation)
		request = prepareDirectorRequest(*operation)
	}

	if err := r.directorClient.UpdateOperation(ctx, &request); err != nil {
		return ctrl.Result{}, err
	}

	operation.Status.Phase = state
	operation.Status.Webhooks[0].State = state
	if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
		return ctrl.Result{}, err
	}

	log.Info(operationResourceMessage("Successfully updated operation status", operation))
	return ctrl.Result{}, nil
}

func (r *OperationReconciler) requeueIfTimeoutNotReached(ctx context.Context, operation *v1alpha1.Operation, webhook graphql.Webhook, webhookErr error) (ctrl.Result, error) {
	if !isWebhookTimeoutReached(r.webhookConfig, webhook, operation.CreationTimestamp.Time) {
		requeueAfter := r.webhookConfig.RequeueInterval
		if webhook.RetryInterval != nil {
			requeueAfter = time.Duration(*webhook.RetryInterval)
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	return r.updateStatusWithError(ctx, operation, webhookErr)
}

func parseRequestData(operation *v1alpha1.Operation) (webhook.RequestData, error) {
	str := struct {
		Application graphql.Application
		TenantID    string
		Headers     map[string]string
	}{}

	if err := json.Unmarshal([]byte(operation.Spec.RequestData), &str); err != nil {
		return webhook.RequestData{}, err
	}

	return webhook.RequestData{
		Application: &str.Application,
		TenantID:    str.TenantID,
		Headers:     str.Headers,
	}, nil
}

func prepareDirectorRequest(operation v1alpha1.Operation) director.Request {
	return prepareDirectorRequestWithError(operation, nil)
}

func prepareDirectorRequestWithError(operation v1alpha1.Operation, err error) director.Request {
	request := director.Request{
		OperationType: graphql.OperationType(operation.Spec.OperationType),
		ResourceType:  resource.Type(operation.Spec.ResourceType),
		ResourceID:    operation.Spec.ResourceID,
	}

	if err != nil {
		request.Error = err.Error()
	}

	return request
}

func isReconciliationTimeoutReached(operation *v1alpha1.Operation, reconciliationTimeout time.Duration) bool {
	operationEndTime := operation.ObjectMeta.CreationTimestamp.Time.Add(reconciliationTimeout)

	return time.Now().After(operationEndTime)
}

func isWebhookTimeoutReached(cfg *web_hook.Config, webhook graphql.Webhook, creationTime time.Time) bool {
	webhookTimeout := cfg.WebhookTimeout
	if webhook.Timeout != nil {
		webhookTimeout = time.Duration(*webhook.Timeout)
	}

	operationEndTime := creationTime.Add(webhookTimeout)
	return time.Now().After(operationEndTime)
}

func getNextPollTime(webhook *graphql.Webhook, webhookStatus v1alpha1.Webhook, timeLayout string) (time.Duration, error) {
	if webhookStatus.LastPollTimestamp == "" {
		return 0, nil
	}

	lastPollTimestamp, err := time.Parse(timeLayout, webhookStatus.LastPollTimestamp)
	if err != nil {
		return 0, err
	}

	nextPollTime := lastPollTimestamp.Add(time.Duration(*webhook.RetryInterval) * time.Second)
	return time.Until(nextPollTime), nil
}

func prepareInitialStatus(operation v1alpha1.Operation) *v1alpha1.Operation {
	overrideStatus := false
	if operation.Generation != operation.Status.ObservedGeneration {
		overrideStatus = true
		operation.Status.ObservedGeneration = operation.Generation
		operation.Status.Phase = ""
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
				State:     v1alpha1.StateInProgress,
			})
		}
	}

	return &operation
}

func setConditionStatus(operation v1alpha1.Operation, conditionType v1alpha1.ConditionType, status corev1.ConditionStatus, msg string) *v1alpha1.Operation {
	for i, condition := range operation.Status.Conditions {
		if condition.Type == conditionType {
			operation.Status.Conditions[i].Status = status
			operation.Status.Conditions[i].Message = msg
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
	webhooks := make(map[string]graphql.Webhook)

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
			totalTimeout += time.Duration(*webhook.Timeout) * time.Second
		}
	}

	return totalTimeout
}

func operationResourceMessage(msg string, operation *v1alpha1.Operation) string {
	return fmt.Sprintf("%s [resource ID: %s, type: %s; conditions: %v]", msg, operation.Spec.ResourceID, operation.Spec.ResourceType, operation.Status.Conditions)
}

func isStringPointerEmpty(ptr *string) bool {
	return ptr != nil && len(*ptr) != 0
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
