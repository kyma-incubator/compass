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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Operation{}, &OperationList{})
}

// +kubebuilder:validation:Enum=Create;Update;Delete;Unpair
type OperationType string

const (
	OperationTypeCreate OperationType = "Create"
	OperationTypeUpdate OperationType = "Update"
	OperationTypeDelete OperationType = "Delete"
)

// OperationSpec defines the desired state of Operation
type OperationSpec struct {
	OperationID       string        `json:"operation_id"`
	OperationType     OperationType `json:"operation_type"`
	OperationCategory string        `json:"operation_category"`
	ResourceType      string        `json:"resource_type"`
	ResourceID        string        `json:"resource_id"`
	CorrelationID     string        `json:"correlation_id"`
	WebhookIDs        []string      `json:"webhook_ids"`
	RequestObject     string        `json:"request_object"`
}

// +kubebuilder:validation:Enum=Success;Failed;In Progress
type State string

const (
	StateSuccess    State = "Success"
	StateFailed     State = "Failed"
	StateInProgress State = "In Progress"
)

// Webhook is an entity part of the OperationStatus which holds information
// about the progression of the webhook execution
type Webhook struct {
	WebhookID         string `json:"webhook_id"`
	RetriesCount      int    `json:"retries_count"`
	WebhookPollURL    string `json:"webhook_poll_url"`
	LastPollTimestamp string `json:"last_poll_timestamp"`
	State             State  `json:"state"`
}

// +kubebuilder:validation:Enum=Ready;Error
type ConditionType string

const (
	ConditionTypeReady ConditionType = "Ready"
	ConditionTypeError ConditionType = "Error"
)

// Condition defines the states which the Operation CR can take
type Condition struct {
	Type    ConditionType          `json:"type"`
	Status  corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	Message string                 `json:"message,omitempty"`
}

// OperationStatus defines the observed state of Operation
type OperationStatus struct {
	Webhooks           []Webhook   `json:"webhooks,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
	Phase              State       `json:"phase,omitempty"`
	InitializedAt      metav1.Time `json:"initialized_at,omitempty"`
	ObservedGeneration *int64      `json:"observed_generation,omitempty"`
}

// +kubebuilder:object:root=true

// Operation is the Schema for the operations API
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.operation_type`
// +kubebuilder:printcolumn:name="Resource ID",type=string,JSONPath=`.spec.resource_id`
// +kubebuilder:printcolumn:name="Resource Type",type=string,JSONPath=`.spec.resource_type`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:subresource:status
type Operation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperationSpec   `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OperationList contains a list of Operation
type OperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Operation `json:"items"`
}

// OperationValidationErr represents an operation error potentially occurring during the intialization and validation of the OperationStatus
type OperationValidationErr struct {
	Description string
}

// Error implements the error interface
func (o *OperationValidationErr) Error() string {
	return o.Description
}

// Validate implements validation logic for the Operation CR
func (in *Operation) Validate() error {
	webhookCount := len(in.Spec.WebhookIDs)
	if webhookCount > 1 {
		return &OperationValidationErr{Description: fmt.Sprintf("expected 0 or 1 webhook for execution, found: %d", webhookCount)}
	}

	return nil
}

// HasPollURL checks whether the current Operation has been provided with a Poll URL
func (in *Operation) HasPollURL() bool {
	return in.PollURL() != ""
}

// PollURL returns the Poll URL for the current Operation
// and empty string if a URL has not been provided
func (in *Operation) PollURL() string {
	if len(in.Status.Webhooks) == 0 {
		return ""
	}

	return in.Status.Webhooks[0].WebhookPollURL
}

// NextPollTime calculates the remaining time until the Poll URL associated with
// the current Operation can be requested/polled again.
func (in *Operation) NextPollTime(retryInterval *int, timeLayout string) (time.Duration, error) {
	if len(in.Status.Webhooks) == 0 || in.Status.Webhooks[0].LastPollTimestamp == "" || retryInterval == nil {
		return 0, nil
	}

	lastPollTimestamp, err := time.Parse(timeLayout, in.Status.Webhooks[0].LastPollTimestamp)
	if err != nil {
		return 0, err
	}

	nextPollTime := lastPollTimestamp.Add(time.Duration(*retryInterval) * time.Second)
	return time.Until(nextPollTime), nil
}

// TimeoutReached returns whether the current operation has timed-out or not
// based on the InitializedAt status timestamp of the Operation and the provided
// timeout duration variable.
func (in *Operation) TimeoutReached(timeout time.Duration) bool {
	operationEndTime := in.Status.InitializedAt.Time.Add(timeout)

	return time.Now().After(operationEndTime)
}

// IsInProgress checks if the Operation's Phase is StateInProgress
func (in *Operation) IsInProgress() bool {
	return in.Status.Phase == StateInProgress
}

// RequestObject parses and returns the request object associated with
// the current operation. The request object is essential for the processing of
// webhook templates as part of the Operation Controller reconcile logic.
func (in *Operation) RequestObject() (*webhook.ApplicationLifecycleWebhookRequestObject, error) {
	str := struct {
		Application graphql.Application
		TenantID    string
		Headers     map[string]string
	}{}

	if err := json.Unmarshal([]byte(in.Spec.RequestObject), &str); err != nil {
		return &webhook.ApplicationLifecycleWebhookRequestObject{}, err
	}

	return &webhook.ApplicationLifecycleWebhookRequestObject{
		Application: &str.Application,
		TenantID:    str.TenantID,
		Headers:     str.Headers,
	}, nil
}
