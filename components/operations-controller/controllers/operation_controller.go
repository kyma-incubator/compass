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
	webhook_dir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// KubernetesClient is a defines a Kubernetes client capable of retrieving and deleting resources as well as updating their status
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KubernetesClient
type KubernetesClient interface {
	Get(ctx context.Context, key client.ObjectKey) (*v1alpha1.Operation, error)
	UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
}

// DirectorClient defines a Director client which is capable of fetching an application
// and notifying Director for operation state changes
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DirectorClient
type DirectorClient interface {
	types.ApplicationLister
	UpdateOperation(ctx context.Context, request *director.Request) error
}

// WebhookClient defines a general purpose Webhook executor client
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . WebhookClient
type WebhookClient interface {
	Do(ctx context.Context, request *webhook.Request) (*webhook_dir.Response, error)
	Poll(ctx context.Context, request *webhook.PollRequest) (*webhook_dir.ResponseStatus, error)
}

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	config         *webhook.Config
	logger         logr.Logger
	k8sClient      KubernetesClient
	directorClient DirectorClient
	webhookClient  WebhookClient
}

func NewOperationReconciler(config *webhook.Config, logger logr.Logger, k8sClient KubernetesClient, directorClient DirectorClient, webhookClient WebhookClient) *OperationReconciler {
	return &OperationReconciler{
		config:         config,
		logger:         logger,
		k8sClient:      k8sClient,
		directorClient: directorClient,
		webhookClient:  webhookClient,
	}
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.logger.WithValues("operation", req.NamespacedName)

	operation, err := r.k8sClient.Get(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))
		if statusError, ok := err.(*k8s_errors.StatusError); ok && statusError.ErrStatus.Code == http.StatusNotFound {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	webhookCount := len(operation.Spec.WebhookIDs)
	if webhookCount != 1 {
		err := fmt.Errorf("expected 1 webhook for execution, found %d", webhookCount)
		logger.Error(err, "Invalid operation webhooks provided")
		return r.updateStatusWithError(ctx, operation, err)
	}

	prepareInitialStatus(operation)

	webhookStatus := &operation.Status.Webhooks[0]
	if webhookStatus.State != v1alpha1.StateInProgress {
		log.Info(fmt.Sprintf("Webhook associated with operation has already been executed and finished with state %q. Will return with no requeue", webhookStatus.State))
		return ctrl.Result{}, nil
	}

	requestObject, err := parseRequestObject(operation)
	if err != nil {
		logger.Error(err, "Unable to parse request object")
		return r.updateStatusWithError(ctx, operation, err)
	}

	ctx = tenant.SaveToContext(ctx, requestObject.TenantID)
	app, err := r.directorClient.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		logger.Error(err, "Unable to fetch application")

		if isReconciliationTimeoutReached(operation, time.Duration(r.config.TimeoutFactor)*r.config.ReconciliationTimeout) {
			if err := r.k8sClient.Delete(ctx, operation); err != nil {
				return ctrl.Result{}, err
			}

			log.Info("Successfully deleted operation")
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

		log.Info("Successfully updated operation status", "status", operation.Status)
		return ctrl.Result{}, nil
	}

	if !isStringPointerEmpty(app.Result.Error) {
		operation = setErrorStatus(*operation, *app.Result.Error)
		operation.Status.Phase = v1alpha1.StateFailed
		operation.Status.Webhooks[0].State = v1alpha1.StateFailed
		if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Successfully updated operation status", "status", operation.Status)
		return ctrl.Result{}, nil
	}

	webhookEntity, err := retrieveWebhook(app.Result.Webhooks, operation.Spec.WebhookIDs[0])
	if err != nil {
		logger.Error(err, "Unable to retrieve webhooks")
		return r.updateStatusWithError(ctx, operation, err)
	}

	if isReconciliationTimeoutReached(operation, calculateReconciliationTimeout(webhookEntity, r.config.WebhookTimeout)) {
		log.Info("Reconciliation timeout reached")
		return r.updateStatusWithError(ctx, operation, ErrReconciliationTimeoutReached)
	}

	if webhookStatus.WebhookPollURL == "" {
		log.Info("Webhook Poll URL is not found. Will attempt to execute the webhook")
		request := webhook.NewRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID)

		response, err := r.webhookClient.Do(ctx, request)
		if err != nil {
			logger.Error(err, "Unable to execute Webhook request")
			return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
		}

		switch *webhookEntity.Mode {
		case graphql.WebhookModeAsync:
			log.Info("Asynchronous webhook initial request has been executed successfully")

			operation.Status.Webhooks[0].WebhookPollURL = *response.Location
			operation.Status.Webhooks[0].State = v1alpha1.StateInProgress
			operation.Status.Phase = v1alpha1.StateInProgress
			if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
				return ctrl.Result{}, err
			}

			log.Info("Successfully updated operation status", "status", operation.Status)
			return ctrl.Result{Requeue: true}, nil
		case graphql.WebhookModeSync:
			log.Info("Synchronous webhook has been executed successfully")
			return r.updateStatusSuccess(ctx, operation)
		default:
			logger.Error(ErrUnsupportedWebhookMode, "Unable to post-process Webhook response")
			return ctrl.Result{}, nil
		}
	}

	log.Info("Webhook Poll URL is found. Will calculate next poll time")
	requeueAfter, err := getNextPollTime(webhookEntity, *webhookStatus, r.config.TimeLayout)
	if err != nil {
		logger.Error(err, "Unable to calculate next poll time")
		return r.updateStatusWithError(ctx, operation, err)
	}

	if requeueAfter > 0 {
		logger.Info(fmt.Sprintf("Poll interval has not passed. Will requeue after: %d seconds", requeueAfter*time.Second))
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	request := webhook.NewPollRequest(*webhookEntity, requestObject, operation.Spec.CorrelationID, webhookStatus.WebhookPollURL)
	response, err := r.webhookClient.Poll(ctx, request)
	if err != nil {
		logger.Error(err, "Unable to execute Webhook Poll request")
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, err)
	}

	log.Info(fmt.Sprintf("Asynchronous webhook polling request has been executed successfully with response status: %s", *response.Status))
	switch *response.Status {
	case *response.InProgressStatusIdentifier:
		lastPollTimestamp := time.Now().Format(r.config.TimeLayout)
		operation.Status.Webhooks[0].LastPollTimestamp = lastPollTimestamp
		if err := r.k8sClient.UpdateStatus(ctx, operation); err != nil {
			return ctrl.Result{}, err
		}

		log.Info(fmt.Sprintf("Successfully updated operation status last poll timestamp to %s", lastPollTimestamp), "status", operation.Status)
		return r.requeueIfTimeoutNotReached(ctx, operation, webhookEntity, ErrWebhookTimeoutReached)
	case *response.SuccessStatusIdentifier:
		return r.updateStatusSuccess(ctx, operation)
	case *response.FailedStatusIdentifier:
		return r.updateStatusWithError(ctx, operation, ErrFailedWebhookStatus)
	default:
		return ctrl.Result{}, fmt.Errorf("unexpected poll status response: %s", *response.Status)
	}
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

	log.Info("Successfully updated operation status", "status", operation.Status)
	return ctrl.Result{}, nil
}

func (r *OperationReconciler) requeueIfTimeoutNotReached(ctx context.Context, operation *v1alpha1.Operation, webhook *graphql.Webhook, webhookErr error) (ctrl.Result, error) {
	if !isWebhookTimeoutReached(r.config, webhook, operation.CreationTimestamp.Time) {
		requeueAfter := r.config.RequeueInterval
		if webhook.RetryInterval != nil {
			requeueAfter = time.Duration(*webhook.RetryInterval)
		}
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}
	return r.updateStatusWithError(ctx, operation, webhookErr)
}

func parseRequestObject(operation *v1alpha1.Operation) (webhook_dir.RequestObject, error) {
	str := struct {
		Application graphql.Application
		TenantID    string
		Headers     map[string]string
	}{}

	if err := json.Unmarshal([]byte(operation.Spec.RequestObject), &str); err != nil {
		return webhook_dir.RequestObject{}, err
	}

	return webhook_dir.RequestObject{
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

func isWebhookTimeoutReached(cfg *webhook.Config, webhook *graphql.Webhook, creationTime time.Time) bool {
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

func prepareInitialStatus(operation *v1alpha1.Operation) {
	overrideStatus := false
	if operation.Generation != operation.Status.ObservedGeneration {
		overrideStatus = true
		operation.Status.ObservedGeneration = operation.Generation
	}

	if operation.Status.Phase == "" || overrideStatus {
		operation.Status.Phase = v1alpha1.StateInProgress
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

	return
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

func retrieveWebhook(appWebhooks []graphql.Webhook, operationWebhookID string) (*graphql.Webhook, error) {
	for _, appWebhook := range appWebhooks {
		if appWebhook.ID == operationWebhookID {
			return &appWebhook, nil
		}
	}

	return nil, fmt.Errorf("missing webhook with ID: %s", operationWebhookID)
}

func calculateReconciliationTimeout(webhook *graphql.Webhook, defaultWebhookTimeout time.Duration) time.Duration {
	if webhook.Timeout == nil {
		return defaultWebhookTimeout
	}
	return time.Duration(*webhook.Timeout) * time.Second
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
