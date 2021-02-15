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
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

const (
	defaultReconciliationTimeout = 12 //TODO: Extract in environment variable
	defaultTimeoutFactor         = 2
	defaultWebhookTimeout        = 2
	defaultRequeueInterval       = 10 * time.Minute
)

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	client.Client
	Log    logr.Logger
	Lister types.ApplicationLister
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("operation", req.NamespacedName)

	// your logic here
	var operation = &v1alpha1.Operation{}
	err := r.Get(ctx, req.NamespacedName, operation)
	if err != nil {
		logger.Info(fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))

		return ctrl.Result{}, err
	}

	operation = setDefaultStatus(*operation)

	reconciliationTimeout := time.Duration(defaultTimeoutFactor * defaultReconciliationTimeout)

	app, err := r.Lister.FetchApplication(ctx, operation.Spec.ResourceID)
	if err != nil {
		logger.Error(err, operationError("Unable to fetch application", operation))

		if isReconciliationTimeoutReached(operation, reconciliationTimeout) {
			if err := r.Delete(ctx, operation); err != nil {
				logger.Error(err, operationError("Unable to delete operation", operation))
				return ctrl.Result{Requeue: true}, err
			}
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: defaultRequeueInterval}, nil
	}

	webhooks, err := retrieveWebhooks(app.Result.Webhooks, operation.Spec.WebhookIDs)
	if err != nil {
		logger.Error(err, operationError("Unable to retrieve webhooks", operation))
		return ctrl.Result{}, err
	}

	reconciliationTimeout = calculateReconciliationTimeout(webhooks)

	if app.Result.Ready {
		for _, condition := range operation.Status.Conditions {
			if condition.Type == v1alpha1.ConditionTypeReady {
				condition.Status = corev1.ConditionTrue
				condition.Message = ""
			}

			if condition.Type == v1alpha1.ConditionTypeError {
				condition.Status = corev1.ConditionFalse
				condition.Message = ""
			}
		}

		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, nil
	}

	if app.Result.Error != nil {
		for _, condition := range operation.Status.Conditions {
			if condition.Type == v1alpha1.ConditionTypeReady {
				condition.Status = corev1.ConditionFalse
				condition.Message = ""
			}

			if condition.Type == v1alpha1.ConditionTypeError {
				condition.Status = corev1.ConditionTrue
				condition.Message = *app.Result.Error
			}
		}

		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, nil
	}

	if isReconciliationTimeoutReached(operation, reconciliationTimeout) {
		for _, condition := range operation.Status.Conditions {
			if condition.Type == v1alpha1.ConditionTypeReady {
				condition.Status = corev1.ConditionFalse
				condition.Message = ""
			}

			if condition.Type == v1alpha1.ConditionTypeError {
				condition.Status = corev1.ConditionTrue
				condition.Message = "Reconciliation timeout reached"
			}
		}

		// TODO: Probably notify Director here

		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update operation", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, nil
	}

	webhook := webhooks[operation.Spec.WebhookIDs[0]]
	webhookStatus := operation.Status.Webhooks[0]

	if webhookStatus.WebhookPollURL != "" {
		fmt.Println(webhook)
		// process input templates, run webhooks, process output templates

		// update status poll URL
		operation.Status.Webhooks[0].WebhookPollURL = ""
		if err := r.Update(ctx, operation); err != nil {
			logger.Error(err, operationError("Unable to update poll URL", operation))
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, nil
	}

	if requeueAfter := getNextPollTime(&webhook, webhookStatus); requeueAfter > 0 {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	// Make Poll Request ( poll interval passed <-> operation in progress )

	return ctrl.Result{}, nil
}

func isReconciliationTimeoutReached(operation *v1alpha1.Operation, reconciliationTimeout time.Duration) bool {
	operationEndTime := operation.ObjectMeta.CreationTimestamp.Time.Add(reconciliationTimeout)

	return time.Now().After(operationEndTime)
}

func getNextPollTime(webhook *graphql.Webhook, webhookStatus v1alpha1.Webhook) time.Duration {
	nextPollTime := webhookStatus.LastPollTimestamp.Add(time.Duration(*webhook.RetryInterval) * time.Minute)

	return nextPollTime.Sub(time.Now())
}

func setDefaultStatus(operation v1alpha1.Operation) *v1alpha1.Operation {
	if operation.Status.Conditions == nil {
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

	if operation.Status.Webhooks == nil {
		operation.Status.Webhooks = make([]v1alpha1.Webhook, 0)
	}

	for _, webhookID := range operation.Spec.WebhookIDs {
		webhookExists := false
		for _, existingWebhook := range operation.Status.Webhooks {
			if webhookID == existingWebhook.WebhookID {
				webhookExists = true
			}
		}

		if !webhookExists {
			operation.Status.Webhooks = append(operation.Status.Webhooks, v1alpha1.Webhook{
				WebhookID: webhookID,
				State:     v1alpha1.StatePolling,
			})
		}
	}

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

func calculateReconciliationTimeout(webhooks map[string]graphql.Webhook) time.Duration {
	totalTimeout := 0
	for _, webhook := range webhooks {
		if webhook.Timeout == nil {
			totalTimeout += defaultWebhookTimeout
		} else {
			totalTimeout += *webhook.Timeout
		}
	}

	return time.Duration(totalTimeout) * time.Hour
}

func operationError(msg string, operation *v1alpha1.Operation) string {
	return fmt.Sprintf("%s for operation of resource with ID %s and type %s", msg, operation.Spec.ResourceID, operation.Spec.ResourceType)
}

func (r *OperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Operation{}).
		Complete(r)
}
