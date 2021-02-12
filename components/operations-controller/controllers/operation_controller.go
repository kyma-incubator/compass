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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operationsv1alpha1 "github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operations.compass,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operations.compass,resources=operations/status,verbs=get;update;patch

func (r *OperationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("operation", req.NamespacedName)

	// your logic here
	var operation = &operationsv1alpha1.Operation{}
	err := r.Get(ctx, req.NamespacedName, operation)
	if err != nil {
		fmt.Println(">>>>>>", err)
		logger.Info(fmt.Sprintf("Unable to retrieve %s from API server", req.NamespacedName))
		// Do we need to requeue here if we return an error anyway?
		// Also requeue-ing when the event being processed is for an Operation which is deleted, might result in an infinite loop
		return ctrl.Result{Requeue: true}, err
	}

	operation.Status = operationsv1alpha1.OperationStatus{
		Conditions: []operationsv1alpha1.Condition{
			{
				Type:   operationsv1alpha1.ConditionType("Ready"),
				Status: v1.ConditionTrue,
			},
		},
	}

	if err := r.Status().Update(ctx, operation); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	fmt.Printf(">>>>>>>OP Status: %+v\n", operation.Status)

	return ctrl.Result{}, nil
}

func (r *OperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operationsv1alpha1.Operation{}).
		Complete(r)
}
